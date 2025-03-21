package contentdb

import (
	"context"
	"encoding/json"
	"log"
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

var mockServerReqCount int

func mockServer(w http.ResponseWriter, r *http.Request) {
	mockServerReqCount++
	if strings.HasSuffix(r.URL.Path, "/download/") {
		http.Redirect(w, r, "/sfinv.zip", http.StatusFound)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/mock/429") {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/mock/500") {
		http.Error(w, "Too many requests", http.StatusInternalServerError)
		return
	}

	if strings.Contains(r.URL.Path, "/sfinv/releases/6537") {
		http.ServeFile(w, r, "./testdata/releases_6537.json")
		return
	} else if strings.Contains(r.URL.Path, "/sfinv/releases") {
		http.ServeFile(w, r, "./testdata/releases.json")
		return
	}
	if strings.HasSuffix(r.URL.Path, ".zip") {
		http.ServeFile(w, r, "./testdata"+r.URL.Path)
		return
	}
	log.Printf("Not found: " + r.URL.Path)
	http.Error(w, "Not found: "+r.URL.Path, http.StatusNotFound)
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
	origHost := Host
	Host = testServer.URL
	if testing.Verbose() {
		api.SetLogLevel(api.Debug)
	}

	// tearDown
	defer func() {
		testServer.Close()
		Host = origHost
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
			wantBytesLen: 38606,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(context.Background())
			var p *PackageArchive
			var err error
			if p, err = c.Download(tt.args.author, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Package.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBytesLen := len(p.b.Bytes()); tt.wantBytesLen < gotBytesLen {
				t.Errorf("Package.Download() = %v, want min len %v", gotBytesLen, tt.wantBytesLen)
			}
		})
	}
}

func TestDownloadRelease(t *testing.T) {
	// setUp
	testServer := httptest.NewServer(http.HandlerFunc(mockServer))
	origHost := Host
	Host = testServer.URL
	if testing.Verbose() {
		api.SetLogLevel(api.Debug)
	}

	// tearDown
	defer func() {
		testServer.Close()
		Host = origHost
	}()

	type args struct {
		author  string
		name    string
		release string
	}
	tests := []struct {
		name         string
		args         args
		wantBytesLen int
		wantErr      bool
	}{
		{
			name:         "can download specified release",
			args:         args{author: "rubenwardy", name: "sfinv", release: "6537"},
			wantBytesLen: 38606,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(context.Background())
			got, err := c.DownloadRelease(tt.args.author, tt.args.name, tt.args.release)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.DownloadRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				if len(got.Bytes()) != tt.wantBytesLen {
					t.Errorf("Client.DownloadRelease().Bytes() = %v, want %v", len(got.Bytes()), tt.wantBytesLen)
				}
			}
		})
	}
}
