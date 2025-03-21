package api

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
)

// TopLevelDirs is a list of directories to look for Minetest data, like mods
var TopLevelDirs = []string{homeDir() + "/.minetest", "/usr/share/minetest"}

func homeDir() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return u.HomeDir
}

// ErrModNotFound is returned when a mod lookup is not successfull
var ErrModNotFound = errors.New("mod not found")

// LookupModByName searches for the provided mod name into the /mods/
// subfolder on all top level directories where Minetest data can be
// stored.
//
// For Minetest >= 5.6.0, the hint provided is used to desambiguate
// from mods in more than one directory.
func LookupModByName(name string, hint string) (path string, err error) {
	Debugf("searching for mod %s", name)
	// Syntax suggar
	join := filepath.Join
	ls := filepath.Glob

	// Lookup on top level dirs where the mod can be installed
	for _, topDir := range TopLevelDirs {
		Debugf("searching mod %s under %s/mods", name, topDir)
		modDir := join(topDir, "mods", name)
		if isDir(modDir) {
			// Found a directory with the requested mod name!
			Debugf("found mod at %v", modDir)
			return modDir, nil
		}
	}

	// Lookup on modpacks in the top dirs where the mod ban be installed
	for _, topDir := range TopLevelDirs {
		globPattern := join(topDir, "mods", "*", "modpack.conf")
		Debugf("searching mod %s under modpacks at %s", name, globPattern)
		// Mod packs have a modpack.conf file so look for them
		modPacks, err := ls(globPattern)
		if err != nil {
			Warningf("error searching for modpacks: %v", modPacks)
			return "", err
		}
		for _, modPack := range modPacks {
			modDir := join(filepath.Dir(modPack), name)
			if isDir(modDir) {
				// Found a directory within a mod with the requested name!
				Debugf("found mod at %v", modDir)
				return modDir, nil
			}
		}
	}

	Debugf("mod %v not found", name)
	return "", ErrModNotFound
}

func isDir(path string) bool {
	if path == "" {
		return false
	}
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	if stat.IsDir() {
		return true
	}
	return false
}
