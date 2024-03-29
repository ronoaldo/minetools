package contentdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/ronoaldo/minetools/api"
)

// Package is a downloadable content from ContentDB.
type Package struct {
	// Basic package info
	Author           string `json:"author,omitempty"`
	Name             string `json:"name,omitempty"`
	Title            string `json:"title,omitempty"`
	ShortDescription string `json:"short_description,omitempty"`
	Release          int32  `json:"release,omitempty"`
	Thumbnail        string `json:"string,omitempty"`
	Type             string `json:"type,omitempty"`

	// Package details
	LongDescription string   `json:"long_description,omitempty"`
	CreatedAt       string   `json:"created_at,omitempty"`
	License         string   `json:"license,omitempty"`
	MediaLicense    string   `json:"media_license,omitempty"`
	ContentWarnings []string `json:"content_warnings,omitempty"`
	Maintainers     []string `json:"maintainers,omitempty"`
	Provides        []string `json:"provides,omitempty"`
	ScreenShots     []string `json:"screenshots,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	State           string   `json:"state,omitempty"`

	// Statistics
	Score     float32 `json:"score,omitempty"`
	Downloads int32   `json:"downloads,omitempty"`
	Forums    int32   `json:"forums,omitempty"`

	// Links
	IssueTracker string `json:"issue_tracker,omitempty"`
	Repo         string `json:"repo,omitempty"`
	URL          string `json:"url,omitempty"`
}

// PackageRelease is a single downloadable version of a package.
type PackageRelease struct {
	ID          int    `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	ReleaseDate string `json:"release_date,omitempty"`
	URL         string `json:"url,omitempty"`
	Commit      string `json:"commit,omitempty"`
	Downlads    int    `json:"downloads,omitempty"`
}

// Query can be used to filter out the content returned by ListPackages.
type Query struct {
	Type            string
	Query           string
	Author          string
	Tag             []string
	Random          string
	Limit           string
	Hide            string
	Sort            string
	Order           string
	ProtocolVersion string
	EngineVersion   string
	Format          string
}

// NewQuery is a package query constructor that filter by the given query
func NewQuery(q string) *Query {
	return &Query{Query: q}
}

// QueryMods is a package query constructor that returns only Mods
func QueryMods() *Query {
	return &Query{Type: "mod"}
}

// WithType filter packages by type. Type must be 'mod', 'game' or 'txp'.
func (q *Query) WithType(t string) *Query {
	if q != nil {
		q.Type = t
	}
	return q
}

// WithAuthor filter packages by author.
func (q *Query) WithAuthor(author string) *Query {
	if q != nil {
		q.Author = author
	}
	return q
}

// WithTags filter packages by the given tags.
func (q *Query) WithTags(tag ...string) *Query {
	if q != nil {
		q.Tag = append(q.Tag, tag...)
	}
	return q
}

// OrderBy sorts the query by the given criteria. This value is passed to the
// remote endpoint. Allowed values are name, title, score, reviews, downloads,
// created_at, approved_at, last_release
func (q *Query) OrderBy(criteria string) *Query {
	if q != nil {
		q.Order = criteria
	}
	return q
}

// AsValues converts the current query into an url.Values
func (q *Query) AsValues() (qs url.Values) {
	qs = make(url.Values)
	if q == nil {
		return
	}

	if q.Query != "" {
		qs.Set("q", q.Query)
	}
	if q.Type != "" {
		qs.Set("type", q.Type)
	}
	if q.Author != "" {
		qs.Set("author", q.Author)
	}
	for _, t := range q.Tag {
		qs.Add("tag", t)
	}
	if q.Random != "" {
		qs.Set("random", q.Random)
	}
	if q.Limit != "" {
		qs.Set("limit", q.Limit)
	}
	if q.Sort != "" {
		qs.Set("sort", q.Sort)
	}
	if q.Order != "" {
		qs.Set("order", q.Order)
	}
	if q.EngineVersion != "" {
		qs.Set("engine_version", q.EngineVersion)
	}
	if q.ProtocolVersion != "" {
		qs.Set("protocol_version", q.ProtocolVersion)
	}
	if q.Format != "" {
		qs.Set("fmt", q.Format)
	}

	return
}

// ListPackages returns a list of packages with the given query.
// If query is nil, or has all fields empty, all pacakges are returned.
func (c *Client) ListPackages(q *Query) (pkgs []Package, err error) {
	resp, err := c.makeCall("GET", "/api/packages/", q.AsValues(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&pkgs); err != nil {
		return nil, err
	}

	return pkgs, nil
}

// GetPackage return details of the specified package.
func (c *Client) GetPackage(author, name string) (pkg *Package, err error) {
	resp, err := c.makeCall("GET", "/api/packages/"+author+"/"+name, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return nil, err
	}
	return pkg, err
}

func (c *Client) ListReleases(author, name string) (r []*PackageRelease, err error) {
	resp, err := c.makeCall("GET", "/api/packages/"+author+"/"+name+"/releases/", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r, err
}

func (c *Client) GetRelease(author, name, release string) (r *PackageRelease, err error) {
	resp, err := c.makeCall("GET", "/api/packages/"+author+"/"+name+"/releases/"+release, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r, err
}

// Download fetches the package archive from the ContentDB for the current revision.
func (c *Client) Download(author, name string) (*PackageArchive, error) {
	return c.fetchArchive("/packages/" + author + "/" + name + "/download/")
}

// DownloadRelease fetches the provided package from the ContentDB in the specified revision.
func (c *Client) DownloadRelease(author, name, release string) (*PackageArchive, error) {
	r, err := c.GetRelease(author, name, release)
	if err != nil {
		return nil, err
	}
	// We expect download URLs to be relative to the API endpoint.
	return c.fetchArchive(r.URL)
}
func (c *Client) fetchArchive(path string) (*PackageArchive, error) {
	start := time.Now()
	resp, err := c.makeCall("GET", path, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("contentdb: unable to download: %v", err)
	}
	defer resp.Body.Close()
	api.Debugf("Fetching bytes")
	w := &bytes.Buffer{}
	n, err := io.Copy(w, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("contentdb: unable to save downloaded bytes: %v", err)
	}
	api.Debugf("Wrote %d bytes (%v)", n, time.Since(start))
	return NewPackageArchive(w.Bytes())
}
