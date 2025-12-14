package app

import (
	"fmt"
	"net/http"
)

type App struct {
	ExternalURL string
	HTTPClient  interface {
		Do(req *http.Request) (*http.Response, error)
	}
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
	defer resp.Body.Close()

	ua := r.UserAgent()
	msg := fmt.Sprintf(
		"hello %s. I called to %s and got code %d\n",
		ua,
		a.ExternalURL,
		resp.StatusCode,
	)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg))
}
