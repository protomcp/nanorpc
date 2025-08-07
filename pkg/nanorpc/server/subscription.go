package server

import (
	"context"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/container/list"

	"protomcp.org/nanorpc/pkg/nanorpc"
	"protomcp.org/nanorpc/pkg/nanorpc/utils"
)

// SubscriptionMap manages subscriptions organized by path hash
type SubscriptionMap map[uint32]*list.List[*ActiveSubscription]

// AddSubscription adds a subscription to the map
func (sm SubscriptionMap) AddSubscription(pathHash uint32, sub *ActiveSubscription) {
	subList := sm[pathHash]
	if subList == nil {
		subList = list.New[*ActiveSubscription]()
		sm[pathHash] = subList
	}
	subList.PushBack(sub)
}

// GetSubscribers returns the list of subscribers for a path hash
func (sm SubscriptionMap) GetSubscribers(pathHash uint32) *list.List[*ActiveSubscription] {
	return sm[pathHash]
}

// RemoveForSession removes all subscriptions for a given session ID
func (sm SubscriptionMap) RemoveForSession(sessionID string) {
	for pathHash, subList := range sm {
		if subList == nil {
			continue
		}

		subList.DeleteMatchFn(func(sub *ActiveSubscription) bool {
			return sub.Session != nil && sub.Session.ID() == sessionID
		})

		// Remove empty lists to prevent memory leaks
		if subList.Len() == 0 {
			delete(sm, pathHash)
		}
	}
}

// ActiveSubscription tracks a live subscription in a session
type ActiveSubscription struct {
	// Session identification (8-byte aligned fields first)
	Session   Session   // Reference to client session
	CreatedAt time.Time // When subscription was created
	Filter    []byte    // Request data used as filter criteria

	// 4-byte aligned fields
	RequestID int32  // Client's original request ID for correlation
	PathHash  uint32 // FNV-1a hash of path (primary lookup key)
}

// Subscribe adds a new subscription for the given path and request
func (h *DefaultMessageHandler) Subscribe(_ context.Context, session Session, req *nanorpc.NanoRPCRequest) error {
	if h == nil {
		return core.ErrNilReceiver
	}

	// Resolve path from hash or string using existing logic
	_, pathHash, err := h.hashCache.ResolvePath(req)
	if err != nil {
		return sendErrorResponse(session, req, nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR,
			"failed to resolve subscription path")
	}

	// Validate that we have a valid path
	if pathHash == 0 {
		return sendErrorResponse(session, req, nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR,
			"invalid subscription path")
	}

	// Create subscription
	subscription := &ActiveSubscription{
		Session:   session,
		RequestID: req.RequestId,
		PathHash:  pathHash,
		CreatedAt: time.Now(),
		Filter:    req.Data, // Use request data as filter criteria
	}

	// Add to subscription list
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add subscription using the map's method
	h.subscriptions.AddSubscription(pathHash, subscription)

	// Send acknowledgment response
	response := &nanorpc.NanoRPCResponse{
		RequestId:      req.RequestId,
		ResponseType:   nanorpc.NanoRPCResponse_TYPE_RESPONSE,
		ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
	}

	return session.SendResponse(req, response)
}

// Publish sends an update to all subscribers of a given path
func (h *DefaultMessageHandler) Publish(path string, data []byte) error {
	if h == nil {
		return core.ErrNilReceiver
	}

	// Get path hash
	pathHash, err := h.hashCache.Hash(path)
	if err != nil {
		return core.Wrapf(err, "failed to hash path %q", path)
	}

	return h.PublishByHash(pathHash, data)
}

// PublishByHash sends an update to all subscribers of a given path hash
func (h *DefaultMessageHandler) PublishByHash(pathHash uint32, data []byte) error {
	if h == nil {
		return core.ErrNilReceiver
	}

	// Collect updates while holding the lock
	updates := h.collectPendingUpdates(pathHash, data)

	// Send all updates outside the lock to prevent blocking
	var firstErr error
	for _, update := range updates {
		if err := update.session.SendResponse(nil, update.message); err != nil {
			// Report error via callback
			fields := slog.Fields{
				utils.FieldPathHash:  pathHash,
				utils.FieldSessionID: update.session.ID(),
			}
			h.onError(err, update.session, fields, "failed to send subscription update")
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// pendingUpdate represents an update ready to be sent to a subscriber
type pendingUpdate struct {
	session Session
	message *nanorpc.NanoRPCResponse
}

// collectPendingUpdates gathers all updates for a path hash while holding the lock
func (h *DefaultMessageHandler) collectPendingUpdates(pathHash uint32, data []byte) []pendingUpdate {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subList := h.subscriptions.GetSubscribers(pathHash)
	if subList == nil || subList.Len() == 0 {
		return nil
	}

	// Start with no pre-allocation to avoid memory waste
	// List may contain expired sessions
	var updates []pendingUpdate

	// Iterate through all subscriptions for this path
	subList.ForEach(func(sub *ActiveSubscription) bool {
		if sub.Session != nil {
			// Create update message
			update := &nanorpc.NanoRPCResponse{
				RequestId:      sub.RequestID, // Use original request ID for correlation
				ResponseType:   nanorpc.NanoRPCResponse_TYPE_UPDATE,
				ResponseStatus: nanorpc.NanoRPCResponse_STATUS_OK,
				Data:           data,
			}
			updates = append(updates, pendingUpdate{
				session: sub.Session,
				message: update,
			})
		}
		return true
	})

	return updates
}

// RemoveSubscriptionsForSession removes all subscriptions for a given session
// This should be called when a session disconnects
func (h *DefaultMessageHandler) RemoveSubscriptionsForSession(sessionID string) {
	if h == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Use the map's method to remove subscriptions
	h.subscriptions.RemoveForSession(sessionID)
}

// unsubscribeByRequestID removes a specific subscription identified by
// session ID, request ID, and path hash
func (h *DefaultMessageHandler) unsubscribeByRequestID(sessionID string,
	requestID int32, pathHash uint32) bool {
	if h == nil || sessionID == "" || pathHash == 0 {
		return false
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	subList := h.subscriptions[pathHash]
	if subList == nil {
		return false
	}

	// Remove the subscription with matching session and request ID
	var removed bool
	subList.DeleteMatchFn(func(sub *ActiveSubscription) bool {
		match := sub.Session != nil &&
			sub.Session.ID() == sessionID &&
			sub.RequestID == requestID
		if match {
			removed = true
		}
		return match
	})
	return removed
}
