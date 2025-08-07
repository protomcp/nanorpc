package nanorpc

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"darvaza.org/core"
	"google.golang.org/protobuf/proto"
)

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = basicRequestTestCase{}
var _ core.TestCase = decodeResponseDataTestCase{}
var _ core.TestCase = decodeRequestDataTestCase{}
var _ core.TestCase = writeStringTestCase{}
var _ core.TestCase = pathTestCase{}
var _ core.TestCase = enumTestCase{}
var _ core.TestCase = responseEnumTestCase{}
var _ core.TestCase = statusEnumTestCase{}
var _ core.TestCase = protoMessageTestCase{}

type basicRequestTestCase struct {
	request *NanoRPCRequest
	name    string
}

func (tc basicRequestTestCase) Name() string {
	return tc.name
}

func (tc basicRequestTestCase) Test(t *testing.T) {
	helper := NewEncodeDecodeTestHelper(t)
	helper.TestRequestRoundTrip(tc.request, nil)

	// Also test the original JSON comparison for backward compatibility
	b1, err := EncodeRequest(tc.request, nil)
	if !core.AssertNoError(t, err, "EncodeRequest") {
		t.FailNow()
	}

	b2 := core.SliceCopy(b1)
	req2, n, err := DecodeRequest(b2)
	if !core.AssertNoError(t, err, "DecodeRequest") {
		t.FailNow()
	}
	if !core.AssertEqual(t, len(b1), n, "decode length") {
		t.FailNow()
	}

	j1, err := json.Marshal(tc.request)
	if !core.AssertNoError(t, err, "json.Marshal original") {
		t.FailNow()
	}

	j2, err := json.Marshal(req2)
	if !core.AssertNoError(t, err, "json.Marshal decoded") {
		t.FailNow()
	}

	if !core.AssertTrue(t, bytes.Equal(j1, j2), "request match") {
		t.FailNow()
	}

	t.Logf("Encoded: %q", b1)
}

func newBasicRequestTestCase(name string, request *NanoRPCRequest) basicRequestTestCase {
	return basicRequestTestCase{
		name:    name,
		request: request,
	}
}

func basicRequestTestCases() []basicRequestTestCase {
	return core.S(
		newBasicRequestTestCase("ping_request", &NanoRPCRequest{
			RequestId:   123,
			RequestType: NanoRPCRequest_TYPE_PING,
		}),
	)
}

func TestEncodeDecodeRequest(t *testing.T) {
	core.RunTestCases(t, basicRequestTestCases())
}

// decodeResponseDataTestCase represents a test case for DecodeResponseData
type decodeResponseDataTestCase struct {
	response *NanoRPCResponse
	name     string
	wantData bool
	wantErr  bool
}

func (tc decodeResponseDataTestCase) Name() string {
	return tc.name
}

func (tc decodeResponseDataTestCase) Test(t *testing.T) {
	t.Helper()

	out := &NanoRPCRequest{} // Use as proto.Message for testing
	result, hasData, err := DecodeResponseData(tc.response, out)

	if tc.wantErr {
		core.AssertError(t, err, "error")
		return
	}

	if !core.AssertNoError(t, err, "decode error") {
		t.FailNow()
	}
	if !core.AssertEqual(t, tc.wantData, hasData, "hasData") {
		t.FailNow()
	}
	if !core.AssertNotNil(t, result, "result") {
		t.FailNow()
	}
}

func newDecodeResponseDataTestCase(name string, response *NanoRPCResponse,
	wantData, wantErr bool) decodeResponseDataTestCase {
	return decodeResponseDataTestCase{
		name:     name,
		response: response,
		wantData: wantData,
		wantErr:  wantErr,
	}
}

func decodeResponseDataTestCases() []decodeResponseDataTestCase {
	// Create a test request to use as data
	testReq := &NanoRPCRequest{
		RequestId:   456,
		RequestType: NanoRPCRequest_TYPE_PING,
	}
	testData, _ := proto.Marshal(testReq)

	return []decodeResponseDataTestCase{
		newDecodeResponseDataTestCase("success_response", &NanoRPCResponse{
			RequestId:      123,
			ResponseType:   NanoRPCResponse_TYPE_RESPONSE,
			ResponseStatus: NanoRPCResponse_STATUS_OK,
			Data:           testData,
		}, true, false),

		newDecodeResponseDataTestCase("empty_data", &NanoRPCResponse{
			RequestId:      123,
			ResponseType:   NanoRPCResponse_TYPE_RESPONSE,
			ResponseStatus: NanoRPCResponse_STATUS_OK,
			Data:           nil,
		}, false, false),

		newDecodeResponseDataTestCase("error_response", &NanoRPCResponse{
			RequestId:       124,
			ResponseType:    NanoRPCResponse_TYPE_RESPONSE,
			ResponseStatus:  NanoRPCResponse_STATUS_NOT_FOUND,
			ResponseMessage: "not found",
		}, false, true),
	}
}

func TestDecodeResponseData(t *testing.T) {
	core.RunTestCases(t, decodeResponseDataTestCases())
}

// decodeRequestDataTestCase represents a test case for DecodeRequestData
type decodeRequestDataTestCase struct {
	request  *NanoRPCRequest
	name     string
	wantData bool
}

func (tc decodeRequestDataTestCase) Name() string {
	return tc.name
}

func (tc decodeRequestDataTestCase) Test(t *testing.T) {
	t.Helper()

	out := &NanoRPCResponse{} // Use as proto.Message for testing
	result, hasData, err := DecodeRequestData(tc.request, out)

	if !core.AssertNoError(t, err, "decode error") {
		t.FailNow()
	}
	if !core.AssertEqual(t, tc.wantData, hasData, "hasData") {
		t.FailNow()
	}
	if !core.AssertNotNil(t, result, "result") {
		t.FailNow()
	}
}

func newDecodeRequestDataTestCase(name string, request *NanoRPCRequest, wantData bool) decodeRequestDataTestCase {
	return decodeRequestDataTestCase{
		name:     name,
		request:  request,
		wantData: wantData,
	}
}

func decodeRequestDataTestCases() []decodeRequestDataTestCase {
	// Create a test response to use as data
	testResp := &NanoRPCResponse{
		RequestId:      789,
		ResponseType:   NanoRPCResponse_TYPE_RESPONSE,
		ResponseStatus: NanoRPCResponse_STATUS_OK,
	}
	testData, _ := proto.Marshal(testResp)

	return []decodeRequestDataTestCase{
		newDecodeRequestDataTestCase("request_with_data", &NanoRPCRequest{
			RequestId:   123,
			RequestType: NanoRPCRequest_TYPE_REQUEST,
			Data:        testData,
		}, true),

		newDecodeRequestDataTestCase("request_no_data", &NanoRPCRequest{
			RequestId:   124,
			RequestType: NanoRPCRequest_TYPE_REQUEST,
			Data:        nil,
		}, false),

		newDecodeRequestDataTestCase("request_empty_data", &NanoRPCRequest{
			RequestId:   125,
			RequestType: NanoRPCRequest_TYPE_REQUEST,
			Data:        []byte{},
		}, false),

		newDecodeRequestDataTestCase("nil_request", nil, false),
	}
}

func TestDecodeRequestData(t *testing.T) {
	core.RunTestCases(t, decodeRequestDataTestCases())
}

// writeStringTestCase represents a test case for writeString utility
type writeStringTestCase struct {
	expected string
	name     string
	input    []string
}

func (tc writeStringTestCase) Name() string {
	return tc.name
}

func (tc writeStringTestCase) Test(t *testing.T) {
	t.Helper()

	var buf strings.Builder
	writeString(&buf, tc.input...)

	result := buf.String()
	if !core.AssertEqual(t, tc.expected, result, "writeString output") {
		t.FailNow()
	}
}

func newWriteStringTestCase(name string, input []string, expected string) writeStringTestCase {
	return writeStringTestCase{
		expected: expected,
		name:     name,
		input:    input,
	}
}

func writeStringTestCases() []writeStringTestCase {
	return []writeStringTestCase{
		newWriteStringTestCase("single_string", []string{"hello"}, "hello"),
		newWriteStringTestCase("multiple_strings", []string{"hello", " ", "world"}, "hello world"),
		newWriteStringTestCase("empty_strings", []string{"", "test", ""}, "test"),
		newWriteStringTestCase("no_strings", []string{}, ""),
		newWriteStringTestCase("many_strings", []string{"a", "b", "c", "d", "e"}, "abcde"),
	}
}

func TestWriteString(t *testing.T) {
	core.RunTestCases(t, writeStringTestCases())
}

// pathTestCase represents a test case for RegisterPath
type pathTestCase struct {
	path string
	name string
}

func (tc pathTestCase) Name() string {
	return tc.name
}

func (tc pathTestCase) Test(t *testing.T) {
	t.Helper()

	// Test RegisterPath
	RegisterPath(tc.path)

	// Create a request with path to test DehashRequest
	req := &NanoRPCRequest{
		RequestId:   100,
		RequestType: NanoRPCRequest_TYPE_REQUEST,
		PathOneof:   &NanoRPCRequest_Path{Path: tc.path},
	}

	// Test DehashRequest - should return same request since it has path, not hash
	result, found := DehashRequest(req)
	if !core.AssertNotNil(t, result, "DehashRequest result") {
		t.FailNow()
	}
	if !core.AssertTrue(t, found, "DehashRequest found") {
		t.FailNow()
	}
	if !core.AssertEqual(t, tc.path, result.GetPath(), "path preserved") {
		t.FailNow()
	}
}

func newPathTestCase(name string, path string) pathTestCase {
	return pathTestCase{
		name: name,
		path: path,
	}
}

func pathTestCases() []pathTestCase {
	return []pathTestCase{
		newPathTestCase("simple_path", "/test"),
		newPathTestCase("nested_path", "/api/v1/users"),
		newPathTestCase("root_path", "/"),
		newPathTestCase("complex_path", "/api/v2/namespace/resource/action"),
	}
}

func TestPathOperations(t *testing.T) {
	core.RunTestCases(t, pathTestCases())
}

// enumTestCase represents test cases for enum methods
type enumTestCase struct {
	name        string
	requestType NanoRPCRequest_Type
}

func (tc enumTestCase) Name() string {
	return tc.name
}

func (tc enumTestCase) Test(t *testing.T) {
	t.Helper()

	// Test Enum() method
	enumVal := tc.requestType.Enum()
	if !core.AssertNotNil(t, enumVal, "Enum result") {
		t.FailNow()
	}
	if !core.AssertEqual(t, tc.requestType, *enumVal, "Enum value") {
		t.FailNow()
	}

	// Test String() method
	str := tc.requestType.String()
	if !core.AssertNotEqual(t, "", str, "String result") {
		t.FailNow()
	}

	// Test Type() method
	typeDesc := tc.requestType.Type()
	if !core.AssertNotNil(t, typeDesc, "Type result") {
		t.FailNow()
	}

	// Test Number() method
	num := tc.requestType.Number()
	if !core.AssertTrue(t, num >= 0, "Number non-negative") {
		t.FailNow()
	}
}

func newEnumTestCase(name string, requestType NanoRPCRequest_Type) enumTestCase {
	return enumTestCase{
		name:        name,
		requestType: requestType,
	}
}

func enumTestCases() []enumTestCase {
	return []enumTestCase{
		newEnumTestCase("ping_type", NanoRPCRequest_TYPE_PING),
		newEnumTestCase("request_type", NanoRPCRequest_TYPE_REQUEST),
		newEnumTestCase("subscribe_type", NanoRPCRequest_TYPE_SUBSCRIBE),
	}
}

func TestRequestTypeEnum(t *testing.T) {
	core.RunTestCases(t, enumTestCases())
}

// responseEnumTestCase for response enum testing
type responseEnumTestCase struct {
	name         string
	responseType NanoRPCResponse_Type
}

func (tc responseEnumTestCase) Name() string {
	return tc.name
}

func (tc responseEnumTestCase) Test(t *testing.T) {
	t.Helper()

	enumVal := tc.responseType.Enum()
	if !core.AssertNotNil(t, enumVal, "Enum result") {
		t.FailNow()
	}
	if !core.AssertEqual(t, tc.responseType, *enumVal, "Enum value") {
		t.FailNow()
	}

	str := tc.responseType.String()
	if !core.AssertNotEqual(t, "", str, "String result") {
		t.FailNow()
	}

	typeDesc := tc.responseType.Type()
	if !core.AssertNotNil(t, typeDesc, "Type result") {
		t.FailNow()
	}

	num := tc.responseType.Number()
	if !core.AssertTrue(t, num >= 0, "Number non-negative") {
		t.FailNow()
	}
}

func newResponseEnumTestCase(name string, responseType NanoRPCResponse_Type) responseEnumTestCase {
	return responseEnumTestCase{
		name:         name,
		responseType: responseType,
	}
}

func responseEnumTestCases() []responseEnumTestCase {
	return []responseEnumTestCase{
		newResponseEnumTestCase("response_type", NanoRPCResponse_TYPE_RESPONSE),
		newResponseEnumTestCase("update_type", NanoRPCResponse_TYPE_UPDATE),
	}
}

func TestResponseTypeEnum(t *testing.T) {
	core.RunTestCases(t, responseEnumTestCases())
}

// statusEnumTestCase for status enum testing
type statusEnumTestCase struct {
	name   string
	status NanoRPCResponse_Status
}

func (tc statusEnumTestCase) Name() string {
	return tc.name
}

func (tc statusEnumTestCase) Test(t *testing.T) {
	t.Helper()

	enumVal := tc.status.Enum()
	if !core.AssertNotNil(t, enumVal, "Enum result") {
		t.FailNow()
	}
	if !core.AssertEqual(t, tc.status, *enumVal, "Enum value") {
		t.FailNow()
	}

	str := tc.status.String()
	if !core.AssertNotEqual(t, "", str, "String result") {
		t.FailNow()
	}

	typeDesc := tc.status.Type()
	if !core.AssertNotNil(t, typeDesc, "Type result") {
		t.FailNow()
	}

	num := tc.status.Number()
	if !core.AssertTrue(t, num >= 0, "Number non-negative") {
		t.FailNow()
	}
}

func newStatusEnumTestCase(name string, status NanoRPCResponse_Status) statusEnumTestCase {
	return statusEnumTestCase{
		name:   name,
		status: status,
	}
}

func statusEnumTestCases() []statusEnumTestCase {
	return []statusEnumTestCase{
		newStatusEnumTestCase("ok_status", NanoRPCResponse_STATUS_OK),
		newStatusEnumTestCase("not_found_status", NanoRPCResponse_STATUS_NOT_FOUND),
		newStatusEnumTestCase("not_authorized_status", NanoRPCResponse_STATUS_NOT_AUTHORIZED),
		newStatusEnumTestCase("internal_error_status", NanoRPCResponse_STATUS_INTERNAL_ERROR),
	}
}

func TestResponseStatusEnum(t *testing.T) {
	core.RunTestCases(t, statusEnumTestCases())
}

// protoMessageTestCase for proto message methods
type protoMessageTestCase struct {
	message proto.Message
	name    string
}

func (tc protoMessageTestCase) Name() string {
	return tc.name
}

func (tc protoMessageTestCase) Test(t *testing.T) {
	t.Helper()

	// Test ProtoReflect
	reflectMsg := tc.message.ProtoReflect()
	if !core.AssertNotNil(t, reflectMsg, "ProtoReflect result") {
		t.FailNow()
	}

	// Test Descriptor
	desc := reflectMsg.Descriptor()
	if !core.AssertNotNil(t, desc, "Descriptor result") {
		t.FailNow()
	}

	// Test that message is valid (just check it's not nil)
	if !core.AssertNotNil(t, tc.message, "message") {
		t.FailNow()
	}
}

func newProtoMessageTestCase(name string, message proto.Message) protoMessageTestCase {
	return protoMessageTestCase{
		name:    name,
		message: message,
	}
}

func protoMessageTestCases() []protoMessageTestCase {
	return []protoMessageTestCase{
		newProtoMessageTestCase("request_message", &NanoRPCRequest{
			RequestId:   42,
			RequestType: NanoRPCRequest_TYPE_PING,
		}),
		newProtoMessageTestCase("response_message", &NanoRPCResponse{
			RequestId:      43,
			ResponseType:   NanoRPCResponse_TYPE_RESPONSE,
			ResponseStatus: NanoRPCResponse_STATUS_OK,
		}),
	}
}

func TestProtoMessageMethods(t *testing.T) {
	core.RunTestCases(t, protoMessageTestCases())
}
