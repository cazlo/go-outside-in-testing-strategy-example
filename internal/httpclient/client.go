package httpclient

import "net/http"

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type DefaultClient struct {
	Client *http.Client
}

func (c *DefaultClient) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}
