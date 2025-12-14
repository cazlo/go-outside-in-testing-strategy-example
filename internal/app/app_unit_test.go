package app

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockHTTPClient is a mock implementation of the HTTPClient interface
type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, nil
}

// mockReadCloser is a mock implementation of io.ReadCloser
type mockReadCloser struct {
	io.Reader
	CloseFunc func() error
}

func (m *mockReadCloser) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// mockResponseWriter is a mock implementation of http.ResponseWriter that can fail on Write
type mockResponseWriter struct {
	header          http.Header
	WriteFunc       func([]byte) (int, error)
	WriteHeaderFunc func(statusCode int)
}

func (m *mockResponseWriter) Header() http.Header {
	if m.header == nil {
		m.header = make(http.Header)
	}
	return m.header
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(b)
	}
	return len(b), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	if m.WriteHeaderFunc != nil {
		m.WriteHeaderFunc(statusCode)
	}
}

func TestHelloHandler_Unit_NewRequestError(t *testing.T) {
	// Setup app with invalid URL to trigger http.NewRequest error
	// A control character in the URL will cause http.NewRequest to fail
	a := &App{
		ExternalURL: string([]byte{0x7f}),
		HTTPClient:  &mockHTTPClient{},
	}

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()

	a.HelloHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestHelloHandler_Unit_BodyCloseError(t *testing.T) {
	// Setup mock client to return a body that fails on Close
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body: &mockReadCloser{
					Reader: bytes.NewBufferString("response"),
					CloseFunc: func() error {
						return errors.New("close error")
					},
				},
			}, nil
		},
	}

	a := &App{
		ExternalURL: "http://example.com",
		HTTPClient:  mockClient,
	}

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()

	// We just want to ensure it doesn't panic and runs the line that logs the error
	a.HelloHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHelloHandler_Unit_WriteResponseError(t *testing.T) {
	// Setup mock client to return success
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString("response")),
			}, nil
		},
	}

	a := &App{
		ExternalURL: "http://example.com",
		HTTPClient:  mockClient,
	}

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)

	// Use mock response writer that fails on Write
	w := &mockResponseWriter{
		WriteFunc: func(b []byte) (int, error) {
			return 0, errors.New("write error")
		},
	}

	// We just want to ensure it doesn't panic and runs the line that logs the error
	a.HelloHandler(w, req)
}
