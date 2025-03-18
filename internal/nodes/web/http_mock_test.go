package web

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

// mockTransport is a custom http.RoundTripper that mocks HTTP responses
type mockTransport struct {
	mockResponses map[string]*http.Response
	mockErrors    map[string]error
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		mockResponses: make(map[string]*http.Response),
		mockErrors:    make(map[string]error),
	}
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	url := req.URL.String()

	// Check if we have a mock error for this URL
	if err, exists := t.mockErrors[url]; exists {
		return nil, err
	}

	// Check if we have a mock response for this URL
	if resp, exists := t.mockResponses[url]; exists {
		return resp, nil
	}

	// Return a default error if no mock is defined
	return nil, fmt.Errorf("no mock defined for URL: %s", url)
}

func (t *mockTransport) AddMockResponse(url string, statusCode int, body string) {
	t.mockResponses[url] = &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(io.ReadSeeker(nil)),
	}
}

func (t *mockTransport) AddMockError(url string, err error) {
	t.mockErrors[url] = err
}

// TestMockTransport tests the mock transport implementation
func TestMockTransport(t *testing.T) {
	transport := newMockTransport()
	transport.AddMockResponse("https://example.com", 200, "{}")
	transport.AddMockError("https://error.com", fmt.Errorf("mock error"))

	// Simple validation that we can create the mock transport
	if len(transport.mockResponses) != 1 {
		t.Errorf("Expected 1 mock response, got %d", len(transport.mockResponses))
	}

	if len(transport.mockErrors) != 1 {
		t.Errorf("Expected 1 mock error, got %d", len(transport.mockErrors))
	}
}
