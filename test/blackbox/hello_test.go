package blackbox

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cazlo/go-outside-in-testing-strategy-example/test/wiremock"
)

func TestHello(t *testing.T) {
	baseURL := getenv("BASE_URL", "http://localhost:8080")
	wiremockURL := getenv("WIREMOCK_URL", "http://localhost:8081")

	// Setup Wiremock if available
	if wiremockURL != "" {
		setupWiremock(t, wiremockURL)
	}

	req, err := http.NewRequest(
		http.MethodGet,
		baseURL+"/hello",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "blackbox-test")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("error closing response body: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	if !strings.Contains(text, "hello blackbox-test") {
		t.Fatalf("unexpected body: %s", text)
	}
}

func setupWiremock(t *testing.T, wiremockURL string) {
	t.Helper()

	client := wiremock.NewClient(wiremockURL)

	// Wait for Wiremock to be ready (with retry logic)
	var lastErr error
	for i := 0; i < 10; i++ {
		if err := client.HealthCheck(); err == nil {
			break
		} else {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
		}
	}
	if lastErr != nil {
		t.Fatalf("Wiremock not available at %s: %v", wiremockURL, lastErr)
	}

	// Reset any existing stubs
	if err := client.Reset(); err != nil {
		t.Fatalf("Failed to reset Wiremock: %v", err)
	}

	// Create the status/204 stub dynamically
	stub := wiremock.StubMapping{
		Request: wiremock.RequestPattern{
			Method: "GET",
			URL:    "/status/204",
		},
		Response: wiremock.ResponseDef{
			Status: 204,
		},
	}

	if err := client.CreateStub(stub); err != nil {
		t.Fatalf("Failed to create Wiremock stub: %v", err)
	}

	t.Logf("Wiremock configured at %s", wiremockURL)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
