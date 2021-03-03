package contentdb

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ronoaldo/minetools/api"
)

var (
	host = "https://content.minetest.net"
)

// Client implements a basic HTTP client that can be used to talk to the remote
// API endpoitns.
type Client struct {
	hc http.Client
}

// NewClient initializes a Client with the required values to properly operate.
func NewClient(ctx context.Context) *Client {
	return &Client{
		hc: http.Client{},
	}
}

func (c *Client) makeCall(method, path string, query url.Values, body io.Reader) (*http.Response, error) {
	url := host + path + "?" + query.Encode()
	api.Debugf("request %v %v", method, url)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "minetools/0.0.1")
	resp, err := http.DefaultClient.Do(req)
	api.Debugf("response %v: %v", resp.StatusCode, resp.Status)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("contentdb: internal server error")
	}
	return resp, nil
}
