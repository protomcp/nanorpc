package server

import (
	"encoding/json"
	"errors"
	"testing"

	"darvaza.org/core"
	"google.golang.org/protobuf/proto"

	"protomcp.org/nanorpc/pkg/nanorpc"
)

// testData is a test struct for JSON marshaling tests
type testData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// sendOKTestCase tests the SendOK method
type sendOKTestCase struct {
	name        string
	rc          *RequestContext
	data        []byte
	expectError bool
}

func newSendOKTestCase(name string) *sendOKTestCase {
	return &sendOKTestCase{name: name}
}

func (tc *sendOKTestCase) withNilReceiver(data []byte) *sendOKTestCase {
	tc.rc = nil
	tc.data = data
	tc.expectError = true
	return tc
}

func (tc *sendOKTestCase) withRequestContext(requestID int32) *sendOKTestCase {
	tc.rc = &RequestContext{
		Session: &mockSession{},
		Request: &nanorpc.NanoRPCRequest{
			RequestId: requestID,
		},
	}
	return tc
}

func (tc *sendOKTestCase) withData(data []byte) *sendOKTestCase {
	tc.data = data
	return tc
}

func (tc *sendOKTestCase) test(t *testing.T) {
	t.Helper()
	err := tc.rc.SendOK(tc.data)
	if (err != nil) != tc.expectError {
		t.Errorf("SendOK() error = %v, expectError %v", err, tc.expectError)
	}

	if !tc.expectError && tc.rc != nil {
		verifyOKResponse(t, tc.rc, tc.data)
	}
}

// verifyOKResponse checks the OK response is correct
func verifyOKResponse(t *testing.T, rc *RequestContext, expectedData []byte) {
	t.Helper()
	session := getSessionFromContext(t, rc)
	core.AssertEqual(t, nanorpc.NanoRPCResponse_STATUS_OK, session.lastResponse.ResponseStatus, "status")
	core.AssertEqual(t, string(expectedData), string(session.lastResponse.Data), "data")
}

// getSessionFromContext safely extracts the mock session from RequestContext
func getSessionFromContext(t *testing.T, rc *RequestContext) *mockSession {
	t.Helper()
	session, ok := rc.Session.(*mockSession)
	if !ok {
		t.Fatal("expected Session to be *mockSession")
	}
	return session
}

// TestRequestContext_SendOK tests the SendOK method
func TestRequestContext_SendOK(t *testing.T) {
	tests := []*sendOKTestCase{
		newSendOKTestCase("nil receiver").
			withNilReceiver([]byte("test")),
		newSendOKTestCase("send with data").
			withRequestContext(123).
			withData([]byte("response data")),
		newSendOKTestCase("send without data").
			withRequestContext(456).
			withData(nil),
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

// sendErrorTestCase tests the SendError method
type sendErrorTestCase struct {
	rc          *RequestContext
	name        string
	message     string
	status      nanorpc.NanoRPCResponse_Status
	expectError bool
}

func newSendErrorTestCase(name string) *sendErrorTestCase {
	return &sendErrorTestCase{name: name}
}

func (tc *sendErrorTestCase) withNilReceiver() *sendErrorTestCase {
	tc.rc = nil
	tc.expectError = true
	return tc
}

func (tc *sendErrorTestCase) withRequestContext(requestID int32) *sendErrorTestCase {
	tc.rc = &RequestContext{
		Session: &mockSession{},
		Request: &nanorpc.NanoRPCRequest{
			RequestId: requestID,
		},
	}
	return tc
}

func (tc *sendErrorTestCase) withError(status nanorpc.NanoRPCResponse_Status, message string) *sendErrorTestCase {
	tc.status = status
	tc.message = message
	return tc
}

func (tc *sendErrorTestCase) test(t *testing.T) {
	t.Helper()
	err := tc.rc.SendError(tc.status, tc.message)
	if (err != nil) != tc.expectError {
		t.Errorf("SendError() error = %v, expectError %v", err, tc.expectError)
	}

	if !tc.expectError && tc.rc != nil {
		verifyErrorResponse(t, tc.rc, tc.status, tc.message)
	}
}

// verifyErrorResponse checks the error response is correct
func verifyErrorResponse(t *testing.T, rc *RequestContext,
	expectedStatus nanorpc.NanoRPCResponse_Status, expectedMessage string) {
	t.Helper()
	session := getSessionFromContext(t, rc)

	// STATUS_OK is converted to STATUS_INTERNAL_ERROR for errors
	if expectedStatus == nanorpc.NanoRPCResponse_STATUS_OK {
		expectedStatus = nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR
	}

	core.AssertEqual(t, expectedStatus, session.lastResponse.ResponseStatus, "status")
	core.AssertEqual(t, expectedMessage, session.lastResponse.ResponseMessage, "message")
}

// TestRequestContext_SendError tests the SendError method
func TestRequestContext_SendError(t *testing.T) {
	tests := []*sendErrorTestCase{
		newSendErrorTestCase("nil receiver").
			withNilReceiver().
			withError(nanorpc.NanoRPCResponse_STATUS_NOT_FOUND, "not found"),
		newSendErrorTestCase("send error response").
			withRequestContext(789).
			withError(nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR, "something went wrong"),
		newSendErrorTestCase("STATUS_OK converted to INTERNAL_ERROR").
			withRequestContext(999).
			withError(nanorpc.NanoRPCResponse_STATUS_OK, "this should not be OK"),
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

// specificErrorTestCase tests specific error helper methods
type specificErrorTestCase struct {
	name           string
	method         func(*RequestContext, string) error
	message        string
	defaultMessage string
	expectedStatus nanorpc.NanoRPCResponse_Status
}

func newSpecificErrorTestCase(name string) *specificErrorTestCase {
	return &specificErrorTestCase{name: name}
}

func (tc *specificErrorTestCase) withMethod(method func(*RequestContext, string) error) *specificErrorTestCase {
	tc.method = method
	return tc
}

func (tc *specificErrorTestCase) withMessage(message string) *specificErrorTestCase {
	tc.message = message
	return tc
}

func (tc *specificErrorTestCase) withDefaultMessage(defaultMessage string) *specificErrorTestCase {
	tc.defaultMessage = defaultMessage
	return tc
}

func (tc *specificErrorTestCase) expectingStatus(status nanorpc.NanoRPCResponse_Status) *specificErrorTestCase {
	tc.expectedStatus = status
	return tc
}

func (tc *specificErrorTestCase) test(t *testing.T) {
	t.Helper()
	rc := &RequestContext{
		Session: &mockSession{},
		Request: &nanorpc.NanoRPCRequest{
			RequestId: 100,
		},
	}

	err := tc.method(rc, tc.message)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	session, ok := rc.Session.(*mockSession)
	if !ok {
		t.Fatal("expected Session to be *mockSession")
	}
	core.AssertEqual(t, tc.expectedStatus, session.lastResponse.ResponseStatus, "status")

	expectedMessage := tc.message
	if expectedMessage == "" && tc.defaultMessage != "" {
		expectedMessage = tc.defaultMessage
	}
	core.AssertEqual(t, expectedMessage, session.lastResponse.ResponseMessage, "message")
}

// TestRequestContext_SpecificErrors tests specific error helper methods
func TestRequestContext_SpecificErrors(t *testing.T) {
	tests := []*specificErrorTestCase{
		newSpecificErrorTestCase("SendNotFound with message").
			withMethod((*RequestContext).SendNotFound).
			withMessage("user not found").
			expectingStatus(nanorpc.NanoRPCResponse_STATUS_NOT_FOUND),
		newSpecificErrorTestCase("SendNotFound without message").
			withMethod((*RequestContext).SendNotFound).
			withMessage("").
			withDefaultMessage("resource not found").
			expectingStatus(nanorpc.NanoRPCResponse_STATUS_NOT_FOUND),
		newSpecificErrorTestCase("SendBadRequest with message").
			withMethod((*RequestContext).SendBadRequest).
			withMessage("invalid input").
			expectingStatus(nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR),
		newSpecificErrorTestCase("SendBadRequest without message").
			withMethod((*RequestContext).SendBadRequest).
			withMessage("").
			withDefaultMessage("bad request").
			expectingStatus(nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR),
		newSpecificErrorTestCase("SendUnauthorized with message").
			withMethod((*RequestContext).SendUnauthorized).
			withMessage("invalid token").
			expectingStatus(nanorpc.NanoRPCResponse_STATUS_NOT_AUTHORIZED),
		newSpecificErrorTestCase("SendUnauthorized without message").
			withMethod((*RequestContext).SendUnauthorized).
			withMessage("").
			withDefaultMessage("not authorized").
			expectingStatus(nanorpc.NanoRPCResponse_STATUS_NOT_AUTHORIZED),
		newSpecificErrorTestCase("SendInternalError with message").
			withMethod((*RequestContext).SendInternalError).
			withMessage("database error").
			expectingStatus(nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR),
		newSpecificErrorTestCase("SendInternalError without message").
			withMethod((*RequestContext).SendInternalError).
			withMessage("").
			withDefaultMessage("internal server error").
			expectingStatus(nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR),
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

// sendJSONTestCase tests the SendJSON method
type sendJSONTestCase struct {
	rc          *RequestContext
	value       any
	name        string
	expectError bool
	checkStruct bool
}

func newSendJSONTestCase(name string) *sendJSONTestCase {
	return &sendJSONTestCase{name: name}
}

func (tc *sendJSONTestCase) withNilReceiver() *sendJSONTestCase {
	tc.rc = nil
	tc.expectError = true
	return tc
}

func (tc *sendJSONTestCase) withRequestContext(requestID int32) *sendJSONTestCase {
	tc.rc = &RequestContext{
		Session: &mockSession{},
		Request: &nanorpc.NanoRPCRequest{
			RequestId: requestID,
		},
	}
	return tc
}

func (tc *sendJSONTestCase) withValue(value any) *sendJSONTestCase {
	tc.value = value
	return tc
}

func (tc *sendJSONTestCase) expectingError() *sendJSONTestCase {
	tc.expectError = true
	return tc
}

func (tc *sendJSONTestCase) checkingStruct() *sendJSONTestCase {
	tc.checkStruct = true
	return tc
}

func (tc *sendJSONTestCase) test(t *testing.T) {
	t.Helper()
	err := tc.rc.SendJSON(tc.value)
	if (err != nil) != tc.expectError {
		t.Errorf("SendJSON() error = %v, expectError %v", err, tc.expectError)
	}

	if !tc.expectError && tc.rc != nil {
		verifyJSONResponse(t, tc)
	}
}

// verifyJSONResponse checks the JSON response is correct
func verifyJSONResponse(t *testing.T, tc *sendJSONTestCase) {
	t.Helper()
	session := getSessionFromContext(t, tc.rc)
	core.AssertEqual(t, nanorpc.NanoRPCResponse_STATUS_OK, session.lastResponse.ResponseStatus, "status")

	if tc.checkStruct {
		verifyJSONData(t, session.lastResponse.Data, tc.value)
	}
}

// verifyJSONData verifies the JSON data matches the original
func verifyJSONData(t *testing.T, data []byte, expectedValue any) {
	t.Helper()
	var decoded testData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("failed to unmarshal response data: %v", err)
		return
	}
	original, ok := expectedValue.(testData)
	if !ok {
		t.Fatal("expected value to be testData")
	}
	if decoded.Name != original.Name || decoded.Value != original.Value {
		t.Errorf("decoded data doesn't match: got %+v, want %+v", decoded, original)
	}
}

// TestRequestContext_SendJSON tests the SendJSON method
func TestRequestContext_SendJSON(t *testing.T) {
	tests := []*sendJSONTestCase{
		newSendJSONTestCase("nil receiver").
			withNilReceiver().
			withValue(testData{Name: "test", Value: 42}),
		newSendJSONTestCase("valid struct").
			withRequestContext(200).
			withValue(testData{Name: "test", Value: 42}).
			checkingStruct(),
		newSendJSONTestCase("valid map").
			withRequestContext(201).
			withValue(map[string]any{"key": "value", "number": 123}),
		newSendJSONTestCase("unmarshallable value").
			withRequestContext(202).
			withValue(make(chan int)).
			expectingError(),
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

// sendProtobufTestCase tests the SendProtobuf method
type sendProtobufTestCase struct {
	rc          *RequestContext
	msg         proto.Message
	name        string
	expectError bool
}

func newSendProtobufTestCase(name string) *sendProtobufTestCase {
	return &sendProtobufTestCase{name: name}
}

func (tc *sendProtobufTestCase) withNilReceiver() *sendProtobufTestCase {
	tc.rc = nil
	tc.expectError = true
	return tc
}

func (tc *sendProtobufTestCase) withRequestContext(requestID int32) *sendProtobufTestCase {
	tc.rc = &RequestContext{
		Session: &mockSession{},
		Request: &nanorpc.NanoRPCRequest{
			RequestId: requestID,
		},
	}
	return tc
}

func (tc *sendProtobufTestCase) withMessage(msg proto.Message) *sendProtobufTestCase {
	tc.msg = msg
	return tc
}

func (tc *sendProtobufTestCase) test(t *testing.T) {
	t.Helper()
	err := tc.rc.SendProtobuf(tc.msg)
	if (err != nil) != tc.expectError {
		t.Errorf("SendProtobuf() error = %v, expectError %v", err, tc.expectError)
	}

	if !tc.expectError && tc.rc != nil {
		verifyProtobufResponse(t, tc.rc)
	}
}

// verifyProtobufResponse checks the protobuf response is correct
func verifyProtobufResponse(t *testing.T, rc *RequestContext) {
	t.Helper()
	session := getSessionFromContext(t, rc)
	core.AssertEqual(t, nanorpc.NanoRPCResponse_STATUS_OK, session.lastResponse.ResponseStatus, "status")

	// Verify protobuf data can be unmarshaled
	var decoded nanorpc.NanoRPCRequest
	core.AssertNoError(t, proto.Unmarshal(session.lastResponse.Data, &decoded), "unmarshal")
}

// TestRequestContext_SendProtobuf tests the SendProtobuf method
func TestRequestContext_SendProtobuf(t *testing.T) {
	tests := []*sendProtobufTestCase{
		newSendProtobufTestCase("nil receiver").
			withNilReceiver().
			withMessage(&nanorpc.NanoRPCRequest{RequestId: 300}),
		newSendProtobufTestCase("valid protobuf message").
			withRequestContext(301).
			withMessage(&nanorpc.NanoRPCRequest{
				RequestId:   400,
				RequestType: nanorpc.NanoRPCRequest_TYPE_PING,
			}),
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

// unmarshalJSONTestCase tests the UnmarshalRequestJSON method
type unmarshalJSONTestCase struct {
	rc          *RequestContext
	target      any
	name        string
	expectError bool
	checkValid  bool
}

func newUnmarshalJSONTestCase(name string) *unmarshalJSONTestCase {
	return &unmarshalJSONTestCase{name: name}
}

func (tc *unmarshalJSONTestCase) withNilReceiver() *unmarshalJSONTestCase {
	tc.rc = nil
	tc.expectError = true
	return tc
}

func (tc *unmarshalJSONTestCase) withRequestData(data []byte) *unmarshalJSONTestCase {
	tc.rc = &RequestContext{
		Request: &nanorpc.NanoRPCRequest{
			Data: data,
		},
	}
	return tc
}

func (tc *unmarshalJSONTestCase) withTarget(target any) *unmarshalJSONTestCase {
	tc.target = target
	return tc
}

func (tc *unmarshalJSONTestCase) expectingError() *unmarshalJSONTestCase {
	tc.expectError = true
	return tc
}

func (tc *unmarshalJSONTestCase) checkingValid() *unmarshalJSONTestCase {
	tc.checkValid = true
	return tc
}

func (tc *unmarshalJSONTestCase) test(t *testing.T) {
	t.Helper()
	err := tc.rc.UnmarshalRequestJSON(tc.target)
	if (err != nil) != tc.expectError {
		t.Errorf("UnmarshalRequestJSON() error = %v, expectError %v", err, tc.expectError)
	}

	if !tc.expectError && tc.checkValid {
		verifyUnmarshaledJSON(t, tc.target)
	}
}

// verifyUnmarshaledJSON checks the unmarshaled JSON data
func verifyUnmarshaledJSON(t *testing.T, target any) {
	t.Helper()
	decoded, ok := target.(*testData)
	if !ok {
		t.Fatal("expected target to be *testData")
	}
	if decoded.Name != "test" || decoded.Value != 42 {
		t.Errorf("unexpected decoded data: %+v", decoded)
	}
}

// TestRequestContext_UnmarshalRequestJSON tests the UnmarshalRequestJSON method
func TestRequestContext_UnmarshalRequestJSON(t *testing.T) {
	validJSON, err := json.Marshal(testData{Name: "test", Value: 42})
	if err != nil {
		t.Fatalf("Failed to marshal test JSON data: %v", err)
	}

	tests := []*unmarshalJSONTestCase{
		newUnmarshalJSONTestCase("nil receiver").
			withNilReceiver().
			withTarget(&testData{}),
		newUnmarshalJSONTestCase("valid JSON data").
			withRequestData(validJSON).
			withTarget(&testData{}).
			checkingValid(),
		newUnmarshalJSONTestCase("no data").
			withRequestData(nil).
			withTarget(&testData{}).
			expectingError(),
		newUnmarshalJSONTestCase("invalid JSON").
			withRequestData([]byte("{invalid json}")).
			withTarget(&testData{}).
			expectingError(),
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

// unmarshalProtobufTestCase tests the UnmarshalRequestProtobuf method
type unmarshalProtobufTestCase struct {
	rc          *RequestContext
	target      proto.Message
	name        string
	expectError bool
	checkValid  bool
}

func newUnmarshalProtobufTestCase(name string) *unmarshalProtobufTestCase {
	return &unmarshalProtobufTestCase{name: name}
}

func (tc *unmarshalProtobufTestCase) withNilReceiver() *unmarshalProtobufTestCase {
	tc.rc = nil
	tc.expectError = true
	return tc
}

func (tc *unmarshalProtobufTestCase) withRequestData(data []byte) *unmarshalProtobufTestCase {
	tc.rc = &RequestContext{
		Request: &nanorpc.NanoRPCRequest{
			Data: data,
		},
	}
	return tc
}

func (tc *unmarshalProtobufTestCase) withTarget(target proto.Message) *unmarshalProtobufTestCase {
	tc.target = target
	return tc
}

func (tc *unmarshalProtobufTestCase) expectingError() *unmarshalProtobufTestCase {
	tc.expectError = true
	return tc
}

func (tc *unmarshalProtobufTestCase) checkingValid() *unmarshalProtobufTestCase {
	tc.checkValid = true
	return tc
}

func (tc *unmarshalProtobufTestCase) test(t *testing.T) {
	t.Helper()
	err := tc.rc.UnmarshalRequestProtobuf(tc.target)
	if (err != nil) != tc.expectError {
		t.Errorf("UnmarshalRequestProtobuf() error = %v, expectError %v", err, tc.expectError)
	}

	if !tc.expectError && tc.checkValid {
		verifyUnmarshaledProtobuf(t, tc.target)
	}
}

// verifyUnmarshaledProtobuf checks the unmarshaled protobuf data
func verifyUnmarshaledProtobuf(t *testing.T, target proto.Message) {
	t.Helper()
	decoded, ok := target.(*nanorpc.NanoRPCRequest)
	if !ok {
		t.Fatal("expected target to be *nanorpc.NanoRPCRequest")
	}
	if decoded.RequestId != 500 || decoded.RequestType != nanorpc.NanoRPCRequest_TYPE_REQUEST {
		t.Errorf("unexpected decoded data: %+v", decoded)
	}
}

// TestRequestContext_UnmarshalRequestProtobuf tests the UnmarshalRequestProtobuf method
func TestRequestContext_UnmarshalRequestProtobuf(t *testing.T) {
	validProto, err := proto.Marshal(&nanorpc.NanoRPCRequest{
		RequestId:   500,
		RequestType: nanorpc.NanoRPCRequest_TYPE_REQUEST,
	})
	if err != nil {
		t.Fatalf("Failed to marshal test protobuf data: %v", err)
	}

	tests := []*unmarshalProtobufTestCase{
		newUnmarshalProtobufTestCase("nil receiver").
			withNilReceiver().
			withTarget(&nanorpc.NanoRPCRequest{}),
		newUnmarshalProtobufTestCase("valid protobuf data").
			withRequestData(validProto).
			withTarget(&nanorpc.NanoRPCRequest{}).
			checkingValid(),
		newUnmarshalProtobufTestCase("no data").
			withRequestData(nil).
			withTarget(&nanorpc.NanoRPCRequest{}).
			expectingError(),
		newUnmarshalProtobufTestCase("invalid protobuf").
			withRequestData([]byte("not a protobuf")).
			withTarget(&nanorpc.NanoRPCRequest{}).
			expectingError(),
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

// gettersTestCase tests the getter methods
type gettersTestCase struct {
	rc          *RequestContext
	name        string
	wantData    []byte
	wantID      int32
	wantHasData bool
}

func newGettersTestCase(name string) *gettersTestCase {
	return &gettersTestCase{name: name}
}

func (tc *gettersTestCase) withNilReceiver() *gettersTestCase {
	tc.rc = nil
	tc.wantID = 0
	tc.wantData = nil
	tc.wantHasData = false
	return tc
}

func (tc *gettersTestCase) withNilRequest() *gettersTestCase {
	tc.rc = &RequestContext{}
	tc.wantID = 0
	tc.wantData = nil
	tc.wantHasData = false
	return tc
}

func (tc *gettersTestCase) withRequest(id int32, data []byte) *gettersTestCase {
	tc.rc = &RequestContext{
		Request: &nanorpc.NanoRPCRequest{
			RequestId: id,
			Data:      data,
		},
	}
	tc.wantID = id
	tc.wantData = data
	tc.wantHasData = data != nil
	return tc
}

func (tc *gettersTestCase) test(t *testing.T) {
	t.Helper()
	if got := tc.rc.GetRequestID(); got != tc.wantID {
		t.Errorf("GetRequestID() = %v, want %v", got, tc.wantID)
	}
	if got := tc.rc.GetData(); string(got) != string(tc.wantData) {
		t.Errorf("GetData() = %v, want %v", got, tc.wantData)
	}
	if got := tc.rc.HasData(); got != tc.wantHasData {
		t.Errorf("HasData() = %v, want %v", got, tc.wantHasData)
	}
}

// TestRequestContext_Getters tests the getter methods
func TestRequestContext_Getters(t *testing.T) {
	tests := []*gettersTestCase{
		newGettersTestCase("nil receiver").
			withNilReceiver(),
		newGettersTestCase("nil request").
			withNilRequest(),
		newGettersTestCase("request with data").
			withRequest(600, []byte("test data")),
		newGettersTestCase("request without data").
			withRequest(700, nil),
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

// errorPropagationTestCase tests error propagation
type errorPropagationTestCase struct {
	rc         *RequestContext
	sessionErr error
	name       string
}

func newErrorPropagationTestCase(name string) *errorPropagationTestCase {
	return &errorPropagationTestCase{name: name}
}

func (tc *errorPropagationTestCase) withSessionError(err error) *errorPropagationTestCase {
	tc.sessionErr = err
	tc.rc = &RequestContext{
		Session: &mockSessionWithError{
			sendError: err,
		},
		Request: &nanorpc.NanoRPCRequest{
			RequestId: 800,
		},
	}
	return tc
}

func (tc *errorPropagationTestCase) test(t *testing.T) {
	t.Helper()

	// Test that all send methods propagate the error
	if err := tc.rc.SendOK([]byte("data")); err != tc.sessionErr {
		t.Errorf("SendOK() didn't propagate session error: %v", err)
	}

	if err := tc.rc.SendError(nanorpc.NanoRPCResponse_STATUS_INTERNAL_ERROR, "msg"); err != tc.sessionErr {
		t.Errorf("SendError() didn't propagate session error: %v", err)
	}

	if err := tc.rc.SendJSON(map[string]string{"key": "value"}); err != tc.sessionErr {
		t.Errorf("SendJSON() didn't propagate session error: %v", err)
	}

	if err := tc.rc.SendProtobuf(&nanorpc.NanoRPCRequest{}); err != tc.sessionErr {
		t.Errorf("SendProtobuf() didn't propagate session error: %v", err)
	}
}

// TestRequestContext_ErrorPropagation tests that session errors are properly propagated
func TestRequestContext_ErrorPropagation(t *testing.T) {
	tc := newErrorPropagationTestCase("session error propagation").
		withSessionError(errors.New("session error"))

	t.Run(tc.name, tc.test)
}
