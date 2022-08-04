package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/ronoaldo/minetools/api"
	"github.com/urfave/cli/v2"
)

var (
	worldPath string
	targetDir string
)

var (
	join = filepath.Join
)

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "minetool"
	app.Usage = "Minetest command line utilities."
	app.Commands = []*cli.Command{
		{
			Name:  "world-textures",
			Usage: "extract textures from all mods in the given world path",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "world-path",
					Usage:       "the path to the world directory to be processed",
					Required:    true,
					Destination: &worldPath,
				},
				&cli.StringFlag{
					Name:        "target-dir",
					Usage:       "the path where the textures will be written to",
					Required:    true,
					Destination: &targetDir,
				},
			},
			Action: worldTextures,
		},
	}

	api.LogLevel = api.Debug

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected error: %v", err)
	}
}

func worldTextures(c *cli.Context) error {
	log.Printf("Processing world at %s", worldPath)

	// Load world.mt at worldPath
	worldMt := join(worldPath, "world.mt")
	b, err := os.ReadFile(worldMt)
	if err != nil {
		log.Printf("Could not read file %s: %v", worldMt, err)
		return err
	}

	mods := []string{}
	sc := bufio.NewScanner(bytes.NewReader(b))
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		if strings.HasPrefix(line, "load_mod_") {
			parts := strings.Split(line, "=")
			if len(parts) != 2 {
				log.Printf("Error parsing line %d: '%s': expected only one = sign", lineNo, line)
				continue
			}
			mod, hint := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			enabled := true
			if hint == "false" {
				enabled = false
			}
			mod = strings.Replace(mod, "load_mod_", "", -1)
			log.Printf("Mod %s is enabled: %v", mod, enabled)

			if enabled {
				modPath, err := api.LookupModByName(mod, hint)
				if err != nil {
					log.Printf("Mod %s could not be found. Is this mod installed?", mod)
					continue
				}
				mods = append(mods, modPath)
			}
		}
	}

	log.Printf("Extracting textures from: %v", mods)

	if err := os.MkdirAll(targetDir, 0766); err != nil {
		log.Fatalf("error creating target dir: %v", err)
	}

	for _, modDir := range mods {
		filepath.WalkDir(join(modDir, "textures"), copyTexture)
	}
	return nil
}

func copyTexture(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return nil
	}
	if d != nil && d.IsDir() {
		return nil
	}
	log.Printf("> path=%s, d=%v, err=%v", path, d, err)
	ext := filepath.Ext(path)
	mType := mime.TypeByExtension(ext)
	switch mType {
	case "image/png", "image/jpeg", "image/bmp", "image/x-tga":
		var (
			src io.Reader
			dst io.WriteCloser
			err error
		)
		if src, err = os.Open(path); err != nil {
			log.Fatalf("error opening texture %v: %v", path, err)
		}
		dstName := join(targetDir, filepath.Base(path))
		if dst, err = os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644); err != nil {
			log.Fatalf("error opening target file %v: %v", dstName, err)
		}
		if b, err := io.Copy(dst, src); err != nil {
			log.Fatalf("error copying file: %v", err)
		} else {
			log.Printf("> %d bytes written.", b)
		}
	default:
		log.Printf("Unsupported mime: %v", mType)
	}
	return nil
}
