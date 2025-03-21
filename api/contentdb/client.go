package contentdb

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ronoaldo/minetools/api"
)

var (
	Host = "https://content.minetest.net"
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

var maxRetries int64 = 8
var backoffFactor int64 = 2

func (c *Client) makeCall(method, path string, query url.Values, body io.Reader) (*http.Response, error) {
	var retryCount int64
	for {
		url := Host + path + "?" + query.Encode()
		api.Debugf("Request %v %v", method, url)
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "minetools/go1")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		api.Debugf("Response %v: %v", resp.StatusCode, resp.Status)
		if resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusServiceUnavailable {
			retryCount += 1
			if retryCount > maxRetries {
				return nil, fmt.Errorf("error making API call, after %d retries: %v", retryCount, err)
			}
			backoff := time.Duration(backoffFactor*retryCount) * time.Second
			api.Debugf("Response took too long, waiting %v (Retry %d/%d)", backoff, retryCount, maxRetries)
			time.Sleep(backoff)
			continue
		}
		if resp.Header.Get("x-cache") != "" {
			api.Debugf("Cache Status: %s", resp.Header.Get("x-cache"))
		}
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("contentdb: package not found")
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			b, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("contentdb: error calling API, status=%v, payload=%v", resp.StatusCode, string(b))
		}
		return resp, nil
	}
}
