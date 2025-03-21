package contentdb

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ronoaldo/minetools/api"
)

func TestClient_makeCall(t *testing.T) {
	// setUp
	testServer := httptest.NewServer(http.HandlerFunc(mockServer))
	origHost := Host
	Host = testServer.URL
	if testing.Verbose() {
		api.SetLogLevel(api.Debug)
	}
	origMaxRetries, origBackoffFactor := maxRetries, backoffFactor
	maxRetries, backoffFactor = 2, 1

	// tearDown
	defer func() {
		testServer.Close()
		Host = origHost
		maxRetries, backoffFactor = origMaxRetries, origBackoffFactor
	}()

	// Test cases
	type fields struct {
		hc http.Client
	}
	type args struct {
		method string
		path   string
		query  url.Values
		body   io.Reader
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		want         *http.Response
		wantErr      bool
		wantReqCount int
	}{
		{
			name:         "when HTTP 429 (too many requests) is returned, retry up to 8 times",
			fields:       fields{http.Client{}},
			args:         args{method: "GET", path: "/mock/429"},
			want:         nil,
			wantErr:      true,
			wantReqCount: int(maxRetries) + 1,
		},
		{
			name:         "when HTTP 500 (internal server error) is returned, do not retry",
			fields:       fields{http.Client{}},
			args:         args{method: "GET", path: "/mock/500"},
			want:         nil,
			wantErr:      true,
			wantReqCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize state
			mockServerReqCount = 0
			c := &Client{
				hc: tt.fields.hc,
			}

			// Test
			got, err := c.makeCall(tt.args.method, tt.args.path, tt.args.query, tt.args.body)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.makeCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil && got != nil {
				t.Errorf("Client.makeCall() = %v, want %v", got, tt.want)
			} else if tt.want != nil && got == nil {
				t.Errorf("Client.makeCall() = %v, want %v", got, tt.want)
			} else if tt.want != nil && got != nil {
				if tt.want.StatusCode != got.StatusCode {
					t.Errorf("Client.makeCall().StatusCode = %v, want %v", got.StatusCode, tt.want.StatusCode)
				}
			}

			if tt.wantReqCount != mockServerReqCount {
				t.Errorf("Request count = %v, want %v", mockServerReqCount, tt.wantReqCount)
			}
		})
	}
}
