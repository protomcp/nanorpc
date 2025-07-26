package client

import (
	"crypto/rand"
	"math"
	"math/big"
	"sync/atomic"
)

// RequestCounter emits sequential int31 values
// starting at a random value and skipping 0.
type RequestCounter struct {
	counter atomic.Int32
}

// Next atomically increments the int31 counter,
// skipping zero, and returns the value.
func (c *RequestCounter) Next() int32 {
	if c == nil {
		return c.getRandom()
	}

	for {
		prev := c.counter.Load()
		next := prev + 1
		if next < 1 {
			next = 1
		}

		if c.counter.CompareAndSwap(prev, next) {
			// success
			return next
		}
	}
}

func (*RequestCounter) getRandom() int32 {
	n, err := NewRandomRequestID()
	if err != nil {
		panic(err)
	}
	return n
}

// NewRequestCounter creates a new [RequestCounter]
// with a random starting point.
func NewRequestCounter() (*RequestCounter, error) {
	next, err := NewRandomRequestID()
	if err != nil {
		return nil, err
	}

	c := &RequestCounter{}
	c.counter.Store(next)
	return c, nil
}

// NewRandomRequestID returns a random int31 value.
// To allow using the zero value, 0 is never returned.
func NewRandomRequestID() (int32, error) {
	maxV := big.NewInt(math.MaxInt32)
	for {
		nBig, err := rand.Int(rand.Reader, maxV)
		if err != nil {
			return 0, err
		}

		if n := int32(nBig.Int64()); n > 0 {
			return n, nil
		}
	}
}
