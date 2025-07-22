package http

import (
	"net/http"
	"testing"
)

type mockRoundTripper struct {
	capturedRequest *http.Request
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.capturedRequest = req
	return &http.Response{StatusCode: 200}, nil
}

var headers = map[string]string{
	"a": "aa",
	"b": "bb",
}

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClientWithHeaders(headers)

	if client.Transport == nil {
		t.Error("client.Transport was nil")
	}
}

func TestNewHTTPClientWithHeader(t *testing.T) {
	var name = "headername"
	var value = "headervalue"
	client := NewHTTPClientWithHeader(name, value)

	if client.Transport == nil {
		t.Error("client.Transport was nil")
	}
}

func TestRoundTrip(t *testing.T) {
	var mockRequest = http.Request{
		Header: http.Header{},
	}

	// Create a mock transport that captures the request
	mockTransport := &mockRoundTripper{}
	addHeaderTransport := AddHeaderTransport{mockTransport, headers}

	_, err := addHeaderTransport.RoundTrip(&mockRequest)

	if err != nil {
		t.Errorf("Error is not nil but '%v'", err)
	}

	// Check headers on the captured request (the cloned one)
	if mockTransport.capturedRequest == nil {
		t.Fatal("Request was not captured")
	}

	for name, value := range headers {
		if mockTransport.capturedRequest.Header.Get(name) != value {
			t.Errorf("Expected %v but found %v", value, mockTransport.capturedRequest.Header.Get(name))
		}
	}
}
