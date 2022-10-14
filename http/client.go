package http

import "net/http"

type AddHeaderTransport struct {
	T              http.RoundTripper
	defaultHeaders map[string]string
}

func (adt *AddHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for name, value := range adt.defaultHeaders {
		req.Header.Add(name, value)
	}
	return adt.T.RoundTrip(req)
}

func NewAddHeaderTransport(T http.RoundTripper, headers map[string]string) *AddHeaderTransport {
	if T == nil {
		T = http.DefaultTransport
	}

	return &AddHeaderTransport{T, headers}
}

// creates an http client which injects a http header for each request
func NewHttpClientWithHeader(headerName string, headerValue string) *http.Client {
	headers := make(map[string]string)
	headers[headerName] = headerValue
	client := http.Client{
		Transport: NewAddHeaderTransport(nil, headers),
	}
	return &client
}

// creates an http client which injects a http headers for each request
func NewHttpClientWithHeaders(defaultHeaders map[string]string) *http.Client {
	client := http.Client{
		Transport: NewAddHeaderTransport(nil, defaultHeaders),
	}
	return &client
}
