package http

import (
	"net/http"
	"testing"
)

type mockRoundTripper struct{}

func (mockRoundTripper *mockRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, nil
}

var headers = map[string]string{
	"a": "aa",
	"b": "bb",
}

func TestNewHttpClient(t *testing.T) {
	client := NewHttpClientWithHeaders(headers)

	if client.Transport == nil {
		t.Error("client.Transport was nil")
	}
}

func TestNewHttpClientWithHeader(t *testing.T) {
	var name = "headername"
	var value = "headervalue"
	client := NewHttpClientWithHeader(name, value)

	if client.Transport == nil {
		t.Error("client.Transport was nil")
	}
}

func TestRoundTrip(t *testing.T) {
	var mockRequest = http.Request{
		Header: http.Header{},
	}
	addHeaderTransport := AddHeaderTransport{&mockRoundTripper{}, headers}

	_, err := addHeaderTransport.RoundTrip(&mockRequest)

	if err != nil {
		t.Errorf("Error is not nil but '%v'", err)
	}

	for name, value := range headers {
		if mockRequest.Header.Get(name) != value {
			t.Errorf("Expected %v but found %v", value, mockRequest.Header.Get(name))
		}
	}
}
