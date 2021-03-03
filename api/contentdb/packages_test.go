package contentdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ronoaldo/minetools/api"
)

func logJSON(t *testing.T, v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Error(err)
	}

	t.Logf("logJSON: %v", string(b))
}

func mockServer(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/download/") {
		http.Redirect(w, r, "/uploads/content.zip", 302)
		return
	}
	if strings.HasSuffix(r.URL.Path, ".zip") {
		fmt.Fprint(w, strings.Repeat("*", 1024))
		return
	}
	http.Error(w, "Not found", http.StatusNotFound)
}

func TestListPackages(t *testing.T) {
	type args struct {
		q *Query
	}
	tests := []struct {
		name         string
		args         args
		wantPkgs     []Package
		wantMoreThan int
		wantErr      bool
	}{
		{
			name:         "nil filter returns all",
			args:         args{q: nil},
			wantMoreThan: 500,
		},
		{
			name:         "search for sfinv returns results",
			args:         args{NewQuery("sfinv")},
			wantMoreThan: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(context.Background())
			gotPkgs, err := c.ListPackages(tt.args.q)
			t.Logf("ListPackages(): %d results", len(gotPkgs))
			logJSON(t, gotPkgs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPackages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantPkgs != nil && !reflect.DeepEqual(gotPkgs, tt.wantPkgs) {
				t.Errorf("ListPackages() = %v, want %v", gotPkgs, tt.wantPkgs)
			}
			if tt.wantMoreThan > 0 && len(gotPkgs) <= tt.wantMoreThan {
				t.Errorf("ListPackages() len() = %v, want more than %v", len(gotPkgs), tt.wantMoreThan)
			}
		})
	}
}
func TestGetPackage(t *testing.T) {
	type args struct {
		author string
		name   string
	}
	tests := []struct {
		name    string
		args    args
		wantPkg *Package
		wantErr bool
	}{
		{
			name:    "not found package returns error",
			args:    args{author: "", name: ""},
			wantErr: true,
		},
		{
			name:    "valid package returns package info",
			args:    args{author: "rubenwardy", name: "sfinv"},
			wantPkg: &Package{Author: "rubenwardy", Name: "sfinv"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(context.Background())
			gotPkg, err := c.GetPackage(tt.args.author, tt.args.name)
			logJSON(t, gotPkg)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPackage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantPkg != nil {
				if tt.wantPkg.Author != gotPkg.Author || tt.wantPkg.Name != gotPkg.Name {
					t.Errorf("GetPackage() = %v, want %v", tt.wantPkg, gotPkg)
				}
			}
		})
	}
}

func TestPackageDownload(t *testing.T) {
	// setUp
	testServer := httptest.NewServer(http.HandlerFunc(mockServer))
	origHost := host
	host = testServer.URL
	if testing.Verbose() {
		api.LogLevel = api.Debug
	}

	// tearDown
	defer func() {
		testServer.Close()
		host = origHost
	}()

	// Test
	type args struct {
		author string
		name   string
	}
	tests := []struct {
		name         string
		args         args
		wantBytesLen int
		wantErr      bool
	}{
		{
			name: "download suceeds when URL is valid",
			args: args{
				author: "rubenwardy",
				name:   "sfinv",
			},
			wantBytesLen: 1024,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(context.Background())
			w := &bytes.Buffer{}
			if err := c.Download(tt.args.author, tt.args.name, w); (err != nil) != tt.wantErr {
				t.Errorf("Package.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBytesLen := len(w.Bytes()); tt.wantBytesLen < gotBytesLen {
				t.Errorf("Package.Download() = %v, want min len %v", gotBytesLen, tt.wantBytesLen)
			}
		})
	}
}
