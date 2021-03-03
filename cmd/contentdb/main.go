package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ronoaldo/minetools/api"
	"github.com/ronoaldo/minetools/api/contentdb"
	"github.com/urfave/cli"
	"gopkg.in/ini.v1"
)

func init() {
	ini.PrettyFormat = false
	ini.PrettyEqual = true
}

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:  "install",
			Usage: "install new content",
			Subcommands: []cli.Command{
				{
					Name:   "mod",
					Usage:  "install a new mod",
					Action: installMod,
				},
			},
		},
	}
	app.Before = func(c *cli.Context) error {
		api.LogLevel = api.Debug
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		api.Warningf("unexpected error: %v", err)
		os.Exit(1)
	}
}

func installMod(c *cli.Context) error {
	mods := c.Args()
	cdb := contentdb.NewClient(context.Background())

	for i, mod := range mods {
		var (
			pkg *contentdb.Package
			err error
		)

		api.Debugf("install: installing %v (%v/%v)", mod, i, len(mods))
		if strings.Count(mod, "/") > 0 {
			// Get package details
			s := strings.Split(mod, "/")
			pkg, err = cdb.GetPackage(s[0], s[1])
			if err != nil {
				return fmt.Errorf("install: unable to find %v", mod)
			}
		} else {
			// Lookup package using query
			query := contentdb.NewQuery(mod).WithType("mod").OrderBy("score")
			pkgs, err := cdb.ListPackages(query)
			if err != nil {
				return fmt.Errorf("install: unable to install '%v': %v", mod, err)
			}
			if len(pkgs) == 0 {
				api.Warningf("install: no packages found")
				continue
			}
			api.Debugf("install: found %d packages:", len(pkgs))
			for _, pkg := range pkgs {
				api.Debugf("install: - %v/%v (revision=%v)", pkg.Author, pkg.Name, pkg.Release)
			}
			// TODO(ronoaldo): allow package selection if more than 1 returned
			pkg = &pkgs[0]
		}

		// Download zip file
		buff := &bytes.Buffer{}
		if err := cdb.Download(pkg.Author, pkg.Name, buff); err != nil {
			return err
		}
		r, len := bytes.NewReader(buff.Bytes()), int64(buff.Len())
		z, err := zip.NewReader(r, len)
		if err != nil {
			return err
		}

		// init.lua: mods are expected to have init.lua
		validMod, stripPrefix := findInitLua(z)
		if !validMod {
			return fmt.Errorf("install: unsupported mod: missing */init.lua file")
		}

		// mod.conf: try to load from zip, create empty one if not found.
		var modconf *ini.File
		b, err := readZipFile(z, stripPrefix+"mod.conf")
		switch err {
		case errFileNotFound:
			api.Debugf(stripPrefix + "mod.conf not found, creating one")
			modconf = ini.Empty()
		case nil:
			api.Debugf(stripPrefix + "mod.conf found, using provided one")
			if modconf, err = ini.Load(b); err != nil {
				return err
			}
		default:
			return err
		}

		// Use modName from mod.conf, otherwise from contentdb
		cfg := modconf.Section("")
		modName := pkg.Name
		if cfg.HasKey("name") {
			modName = cfg.Key("name").String()
		} else {
			cfg.Key("name").SetValue(pkg.Name)
		}
		// Update mod.conf with contentdb data so we keep track of updates.
		cfg.Key("author").SetValue(pkg.Author)
		cfg.Key("release").SetValue(fmt.Sprintf("%d", pkg.Release))
		if !cfg.HasKey("title") {
			cfg.Key("title").SetValue(pkg.Title)
		}
		if !cfg.HasKey("description") {
			cfg.Key("description").SetValue(pkg.ShortDescription)
		}

		// Unpack mod contents
		destdir := filepath.Join("mods", modName)
		os.MkdirAll(destdir, 0755)
		modconf.SaveTo(filepath.Join(destdir, "mod.conf"))
		for _, f := range z.File {
			// Ignore directories as they will be auto-created bellow
			if f.FileInfo().IsDir() {
				continue
			}
			if stripPrefix+"mod.conf" == f.Name {
				continue
			}
			// Strip folder prefix while unpacking
			noprefix := strings.Replace(f.Name, stripPrefix, "", 1)
			fname := filepath.FromSlash(noprefix) // zip is '/' separated, convert to filepath.Separator
			// Sanity check: verify target to prevent Zip Slip vul.
			target := filepath.Clean(filepath.Join(destdir, fname))
			if !strings.HasPrefix(target, destdir) {
				api.Warningf("possible Zip Slip found, ignoring %v", f.Name)
				continue
			}
			api.Debugf("Extracting '%v' into '%v'", f.Name, target)
			// Create destination directory
			os.MkdirAll(filepath.Dir(target), 0755)
			// Extract file contents
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("install: error reading from zip: %v", err)
			}
			b, err := ioutil.ReadAll(rc)
			rc.Close()
			if err != nil {
				return fmt.Errorf("install: error reading from zip: %v", err)
			}
			if err = ioutil.WriteFile(target, b, 0644); err != nil {
				return fmt.Errorf("install: error writing to %v: %v", target, err)
			}
		}

	}

	return nil
}

// TODO(ronoaldo): refactor these misc funcs into a archive helper

// findInitLua identifies the first directory where init.lua is found. This
// represents the mod root folder.
func findInitLua(z *zip.Reader) (found bool, stripPrefix string) {
	patterns := []string{"init.lua", "*/init.lua"}

	for _, p := range patterns {
		for _, f := range z.File {
			if ok, _ := path.Match(p, f.Name); ok {
				dir, _ := path.Split(f.Name)
				return true, dir
			}
		}
	}

	return false, ""
}

var errFileNotFound = fmt.Errorf("readZipFile: not found")

func readZipFile(z *zip.Reader, pattern string) ([]byte, error) {
	for _, f := range z.File {
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
	return nil, errFileNotFound
}
