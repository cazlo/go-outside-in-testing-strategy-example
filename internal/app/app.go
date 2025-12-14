package app

import (
	"fmt"
	"log"
	"net/http"
)

type App struct {
	HTTPClient interface {
		Do(req *http.Request) (*http.Response, error)
	}
	ExternalURL string
}

func (a *App) HelloHandler(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(http.MethodGet, a.ExternalURL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("error closing response body: %v", closeErr)
		}
	}()

	ua := r.UserAgent()
	msg := fmt.Sprintf(
		"hello %s. I called to %s and got code %d\n",
		ua,
		a.ExternalURL,
		resp.StatusCode,
	)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(msg)); err != nil {
		log.Printf("error writing response: %v", err)
	}
}
