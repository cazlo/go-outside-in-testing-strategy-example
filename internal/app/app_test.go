package app_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cazlo/go-outside-in-testing-strategy-example/internal/app"
	"github.com/cazlo/go-outside-in-testing-strategy-example/internal/httpclient"
)

func TestHelloHandler_Success(t *testing.T) {
	// Setup ephemeral mock external service using httptest
	mockExternal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request to external service
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		// Return a successful response
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(mockExternal.Close)

	// Setup the app with REAL httpclient
	a := &app.App{
		ExternalURL: mockExternal.URL,
		HTTPClient: &httpclient.DefaultClient{
			Client: &http.Client{
				Timeout: 5 * time.Second,
			},
		},
	}

	// Create a handler and test server
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", a.HelloHandler)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	// Make a request with a custom User-Agent
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/hello", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "test-agent")

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("error closing response body: %v", closeErr)
		}
	}()

	// Validate the response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, "hello test-agent") {
		t.Errorf("expected body to contain 'hello test-agent', got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, mockExternal.URL) {
		t.Errorf("expected body to mention external URL, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "204") {
		t.Errorf("expected body to mention status code 204, got: %s", bodyStr)
	}
}

func TestHelloHandler_ExternalServiceError(t *testing.T) {
	// Setup mock that immediately closes the connection
	mockExternal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate error by panicking, which will close the connection
		panic("simulated error")
	}))
	// Close immediately to simulate connection error
	mockExternal.Close()

	a := &app.App{
		ExternalURL: mockExternal.URL,
		HTTPClient: &httpclient.DefaultClient{
			Client: &http.Client{
				Timeout: 5 * time.Second,
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", a.HelloHandler)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	resp, err := srv.Client().Get(srv.URL + "/hello")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("error closing response body: %v", closeErr)
		}
	}()

	// Should return 502 Bad Gateway when external service fails
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", resp.StatusCode)
	}
}

func TestHelloHandler_InvalidExternalURL(t *testing.T) {
	a := &app.App{
		ExternalURL: "://invalid-url",
		HTTPClient: &httpclient.DefaultClient{
			Client: &http.Client{
				Timeout: 5 * time.Second,
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", a.HelloHandler)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	resp, err := srv.Client().Get(srv.URL + "/hello")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("error closing response body: %v", closeErr)
		}
	}()

	// Should return 500 Internal Server Error for invalid URL
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}
}

func TestHelloHandler_DefaultUserAgent(t *testing.T) {
	mockExternal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(mockExternal.Close)

	a := &app.App{
		ExternalURL: mockExternal.URL,
		HTTPClient: &httpclient.DefaultClient{
			Client: &http.Client{
				Timeout: 5 * time.Second,
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", a.HelloHandler)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	// Make request without setting User-Agent
	resp, err := srv.Client().Get(srv.URL + "/hello")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("error closing response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	bodyStr := string(body)
	// The default Go HTTP client sets User-Agent to "Go-http-client/1.1" or similar
	if !strings.Contains(bodyStr, "hello") {
		t.Errorf("expected body to contain greeting, got: %s", bodyStr)
	}
}
