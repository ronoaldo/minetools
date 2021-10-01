package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/ronoaldo/minetools/api"
	"github.com/ronoaldo/minetools/api/contentdb"
	"github.com/urfave/cli"
	"gopkg.in/ini.v1"
)

var (
	apiDebug         bool
	removeOldpackage bool
	dryRun           bool
)

// Helper functions
var (
	warnf  = color.New(color.FgHiRed, color.Bold).PrintfFunc()
	green  = color.New(color.FgGreen, color.Bold).PrintfFunc()
	yellow = color.New(color.FgYellow, color.Bold).PrintfFunc()
)

func init() {
	ini.PrettyFormat = false
	ini.PrettyEqual = true
}

func main() {
	app := cli.NewApp()
	app.Name = "contentdb"
	app.Usage = "Minetest ContentDB client implementation for headless server administration"
	app.Commands = []cli.Command{
		{
			Name:   "search",
			Usage:  "search for content",
			Action: search,
		},
		{
			Name:  "install",
			Usage: "installs new mod into ./mods folder",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "update",
					Usage:       "update if the mod is already installed, removing old contents",
					Destination: &removeOldpackage,
				},
			},
			Action: func(c *cli.Context) error {
				return installMod(c.Args())
			},
		},
		{
			Name:  "update",
			Usage: "updates all mods in the ./mods folder",
			Action: func(c *cli.Context) error {
				removeOldpackage = true
				return update()
			},
		},
	}
	app.Flags = append(app.Flags, cli.BoolFlag{
		Name:        "debug",
		EnvVar:      "CDB_DEBUG",
		Usage:       "show debug information on console",
		Destination: &apiDebug,
	})
	app.Flags = append(app.Flags, cli.BoolFlag{
		Name:        "dry-run",
		Usage:       "do not actually perform any opertaions, just report what would be done",
		Destination: &dryRun,
	})
	app.Before = func(c *cli.Context) error {
		if apiDebug {
			api.LogLevel = api.Debug
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		warnf("unexpected error: %v", err)
		os.Exit(1)
	}
}

func search(c *cli.Context) error {
	queryString := strings.Join(c.Args(), " ")
	cdb := contentdb.NewClient(context.Background())
	query := contentdb.NewQuery(queryString).OrderBy("score")
	pkgs, err := cdb.ListPackages(query)
	if err != nil {
		return fmt.Errorf("search: unable to call endpoint '%v': %v", queryString, err)
	}
	api.Debugf("search: found %d packages", len(pkgs))
	t := tabwriter.NewWriter(os.Stdout, 18, 0, 1, ' ', 0)
	fmt.Fprintf(t, "Key\tType\tTitle\tRelease\tShort description\n")
	for _, pkg := range pkgs {
		fmt.Fprintf(t, "%s/%s\t%s\t%s\t%d\t%s\n",
			pkg.Author, pkg.Name, pkg.Type, pkg.Title, pkg.Release, fmt.Sprintf("%.60s", pkg.ShortDescription))
	}
	t.Flush()
	return nil
}

func installMod(mods []string) error {
	cdb := contentdb.NewClient(context.Background())

	for _, mod := range mods {
		var (
			pkg *contentdb.Package
			err error
		)

		if strings.Count(mod, "/") != 1 {
			warnf("install: provide a valid package key: author/name (like rubenwardy/sfinv)")
			continue
		}

		// Get package details
		fmt.Println("Searching for package ", mod)
		s := strings.Split(mod, "/")
		pkg, err = cdb.GetPackage(s[0], s[1])
		if err != nil {
			warnf("install: unable to find %v", mod)
			continue
		}

		// Download zip file
		fmt.Printf("Downloading %v/%v@%v ...\n", pkg.Author, pkg.Name, pkg.Release)
		archive, err := cdb.Download(pkg.Author, pkg.Name)
		if err != nil {
			return err
		}

		pkgType := archive.Type()
		if pkgType != contentdb.Mod && pkgType != contentdb.Modpack {
			warnf("install: package is not a mod/modpack: %s", pkgType)
			continue
		}

		// mod.conf/modpack.conf: try to load from zip, create empty one if not found
		var modconf *ini.File

		modconfFilename := "mod.conf"
		// mod root dir is where the first init.lua is
		_, stripPrefix := archive.FindFile("init.lua", 1)
		if pkgType == contentdb.Modpack {
			var found = 0
			// For modpack, load a diferent config name and adjust the stripPrefix
			modconfFilename = "modpack.conf"
			found, stripPrefix = archive.FindFile(modconfFilename, 1)
			if found == 0 {
				// Backwards compatibility
				_, stripPrefix = archive.FindFile("modpack.txt", 1)
			}
		}
		api.Debugf("Processing archive of type %s (stripPrefix=%s)", pkgType, stripPrefix)

		// Initialize package configuration file
		b, err := archive.ReadFile(stripPrefix + modconfFilename)
		switch err {
		case contentdb.ErrFileNotFound:
			api.Debugf(stripPrefix + " " + modconfFilename + " not found, creating one")
			modconf = ini.Empty()
		case nil:
			api.Debugf(stripPrefix + " " + modconfFilename + " found, using provided one")
			if modconf, err = ini.Load(b); err != nil {
				return err
			}
		default:
			return err
		}
		cfg := modconf.Section("")

		// Update mod.conf with contentdb data so we keep track of updates.
		modName := pkg.Name
		cfg.Key("name").SetValue(modName)
		cfg.Key("author").SetValue(pkg.Author)
		cfg.Key("release").SetValue(fmt.Sprintf("%d", pkg.Release))
		if !cfg.HasKey("title") {
			cfg.Key("title").SetValue(pkg.Title)
		}
		if !cfg.HasKey("description") {
			cfg.Key("description").SetValue(pkg.ShortDescription)
		}

		if dryRun {
			yellow("[Package installation skipped, just a dry-run]")
			fmt.Println("")
			return nil
		}

		// Avoid overwrite destination directory, if updatePackage is not provided.
		destdir := filepath.Join("mods", modName)
		if _, err = os.Stat(destdir); !os.IsNotExist(err) {
			if !removeOldpackage {
				return fmt.Errorf("install: %v already exists, exiting (err=%v)", destdir, err)
			}
			fmt.Println("Removing previous installation (performing in-place update as requested) ...")
			if err = os.RemoveAll(destdir); err != nil {
				return fmt.Errorf("install: unable to clean previous install at %v: %v", destdir, err)
			}
		}

		// Unpack mod contents
		fmt.Println("Extracting package contents ...")

		os.MkdirAll(destdir, 0755)
		modconf.SaveTo(filepath.Join(destdir, modconfFilename))
		for _, f := range archive.Contents() {
			// Skip mod.conf as we already created it.
			if stripPrefix+modconfFilename == f {
				continue
			}
			// Strip folder prefix while unpacking
			noprefix := strings.Replace(f, stripPrefix, "", 1)
			fname := filepath.FromSlash(noprefix) // zip is '/' separated, convert to filepath.Separator
			// Sanity check: verify target to prevent Zip Slip vul.
			target := filepath.Clean(filepath.Join(destdir, fname))
			if !strings.HasPrefix(target, destdir) {
				warnf("possible Zip Slip found, ignoring %v", f)
				continue
			}
			api.Debugf("Extracting '%v' into '%v'", f, target)
			// Create destination directory
			os.MkdirAll(filepath.Dir(target), 0755)
			// Extract file contents
			b, err := archive.ReadFile(f)
			if err != nil {
				return fmt.Errorf("install: error reading from zip: %v", err)
			}
			if err = ioutil.WriteFile(target, b, 0644); err != nil {
				return fmt.Errorf("install: error writing to %v: %v", target, err)
			}
		}
		green("Installed %v into %v\n", mod, destdir)
		fmt.Printf("Add load_mod_%s = true to world.mt to use it.\n", pkg.Name)
		fmt.Printf("* Dependencies: %v\n", cfg.Key("depends").String())
		fmt.Printf("* Optional dependencies: %v\n", cfg.Key("optional_depends").String())
	}

	return nil
}

func update() error {
	cdb := contentdb.NewClient(context.Background())
	mods, err := filepath.Glob("mods/*")
	if err != nil {
		warnf("update: error loading mods directory: %v", err)
		return err
	}

	fmt.Printf("Updating %d mods in ./mods ...\n", len(mods))
	for _, mod := range mods {
		if err := updateMod(cdb, mod); err != nil {
			warnf("update: error upgrading mod %v: %v\n", mod, err)
		}
	}
	return nil
}

func updateMod(cdb *contentdb.Client, modDir string) error {
	yellow("* Checking mod dir ./%v ", modDir)
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(modDir); err != nil {
		return err
	}
	defer os.Chdir(pwd)

	_, err = os.Stat(".git")
	if err == nil {
		fmt.Printf("Mod managed using git. Running `git pull`...\n")
		// .git exists, upgrade from upstream
		b, err := exec.Command("git", "pull").CombinedOutput()
		if err != nil {
			warnf("  updateMod: error executing `git pull`: %v", err)
		}
		green("* Git output:\n%s\n", string(b))
		return err
	}

	// Check if it's a mod
	var err1 error
	b, err := ioutil.ReadFile("mod.conf")
	if err != nil {
		// Check if it's a modpack
		b, err1 = ioutil.ReadFile("modpack.conf")
		if err1 != nil {
			return fmt.Errorf("upgradeMod: unable to load configuration (mod=%v; modpack=%v)", err, err1)
		}
	}

	// Check local version
	modconf, err := ini.Load(b)
	if err != nil {
		return err
	}
	cfg := modconf.Section("")
	author := cfg.Key("author").String()
	name := cfg.Key("name").String()
	release, _ := cfg.Key("release").Int64()
	fmt.Printf("Installed version %v/%v@%v, ", author, name, release)

	// Check remote version
	pkg, err := cdb.GetPackage(author, name)
	if err != nil {
		return err
	}
	fmt.Printf("remote version %v/%v@%v: ", pkg.Author, pkg.Name, pkg.Release)

	// Update?
	if int64(pkg.Release) > release && release > 0 {
		fmt.Printf("updating to version %v ...\n", pkg.Release)
		if dryRun {
			yellow("[Update skipped, just a dry-run]")
			fmt.Println("")
			return nil
		}
		os.Chdir(pwd)
		return installMod([]string{fmt.Sprintf("%s/%s", author, name)})
	}
	fmt.Println("package is already up to date.")
	return nil
}
