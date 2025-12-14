package blackbox

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestHello(t *testing.T) {
	baseURL := getenv("BASE_URL", "http://localhost:8080")

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
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	text := string(body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	if !strings.Contains(text, "hello blackbox-test") {
		t.Fatalf("unexpected body: %s", text)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
