package contentdb

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// ArchiveType is the type of the archive.
type ArchiveType int

// Possible archive type values.
const (
	Invalid     ArchiveType = 0
	Mod         ArchiveType = 1
	Modpack     ArchiveType = 2
	TexturePack ArchiveType = 3
	Game        ArchiveType = 4
)

func (a ArchiveType) String() string {
	switch a {
	case Modpack:
		return "modpack"
	case Mod:
		return "mod"
	case TexturePack:
		return "txp"
	case Game:
		return "game"
	}
	return "INVALID"
}

// PackageArchive provides utility functions to analyse the contents of
// downloaded archives from ContentDB.
type PackageArchive struct {
	b        *bytes.Buffer
	z        *zip.Reader
	contents []string
}

// NewPackageArchive initializes a PackageArchive struct with the given byte
// slice b. It's expected that b points to a valid in-memory Zip archive.
func NewPackageArchive(b []byte) (*PackageArchive, error) {
	buff := bytes.NewBuffer(b)
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}
	p := &PackageArchive{
		b: buff,
		z: z,
	}
	return p, nil
}

type fileList []string

func (fl fileList) Swap(i, j int) { fl[i], fl[j] = fl[j], fl[i] }
func (fl fileList) Len() int      { return len(fl) }
func (fl fileList) Less(i, j int) bool {
	iSlashes, jSlashes := strings.Count(fl[i], "/"), strings.Count(fl[j], "/")
	if iSlashes == jSlashes {
		// in the same dir, compare strings
		return fl[i] < fl[j]
	}
	// sort by depth in the dir tree
	return iSlashes < jSlashes
}

// Contents returns a list of the package contents, ignoring directories.
func (p *PackageArchive) Contents() []string {
	if p.contents == nil {
		// Cache a sorted list of file names
		p.contents = make([]string, 0)
		for _, f := range p.z.File {
			if f.FileInfo().IsDir() {
				continue
			}
			p.contents = append(p.contents, f.Name)
		}
		sort.Sort(fileList(p.contents))
	}
	return p.contents
}

// FindFile returns the count of files with the given name and at what directory
// it was found. If max is grather than zero, search will stop after max files
// are found.
func (p *PackageArchive) FindFile(name string, max int) (count int, dir string) {
	for _, f := range p.Contents() {
		if strings.HasSuffix(f, "/"+name) || f == name {
			if dir == "" {
				dir, _ = path.Split(f)
			}
			count++
		}
		if max > 0 && count >= max {
			break
		}
	}
	return count, dir
}

// Type detects the archive type. The algorithm uses the expected contents, and
// has some backwards compatibility.
func (p *PackageArchive) Type() ArchiveType {
	if found, _ := p.FindFile("modpack.conf", 0); found == 1 {
		return Modpack
	}
	if found, _ := p.FindFile("modpack.txt", 0); found == 1 {
		return Modpack
	}
	if found, _ := p.FindFile("mod.conf", 0); found == 1 {
		return Mod
	}
	// If we have only one init.lua file, assume it is a mod
	if found, _ := p.FindFile("init.lua", 0); found == 1 {
		return Mod
	}
	// TODO: add/detect texture pack or game
	return Invalid
}

// Bytes returns the underlying byte slice this archive was initialized with.
func (p *PackageArchive) Bytes() []byte {
	if p.b == nil {
		return nil
	}
	return p.b.Bytes()
}

// ErrFileNotFound .
var ErrFileNotFound = fmt.Errorf("PackageArchive.ReadFile(): not found")

// ReadFile returns a byte slice with the extracted file contents.
func (p *PackageArchive) ReadFile(pattern string) ([]byte, error) {
	for _, f := range p.z.File {
		fname := filepath.Clean(f.Name)
		if matches, _ := filepath.Match(pattern, fname); matches {
			reader, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer reader.Close()
			return ioutil.ReadAll(reader)
		}
	}
	return nil, ErrFileNotFound
}
