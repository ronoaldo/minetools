# minetools

[![Go Reference](https://pkg.go.dev/badge/github.com/ronoaldo/minetools.svg)](https://pkg.go.dev/github.com/ronoaldo/minetools)

Project minetools has several small utilities to manage a Minetest server
from the command line.

**This is a work in progress.**

## contentdb

Command `contentdb` can be used to download mods from
https://content.minetest.net and install them into a local `mods` folder.

To install the contentdb cli, you need a working Go (> 1.16) toolchain:

    git clone https://github.com/ronoaldo/minetools
    go install ./minetools/cmd/contentdb

### Usage

To get the online help with all parameters, run

    contentdb --help

To search for content, use the `search` subcommand:

    contentdb search mesecons

To install a mod/modpack, use the `install` subcommand:

    contentdb install mod rubenwardy/sfinv

To update all mods in the mods folder (including those installed with git!), use the `update` subdommand:

    contentdb update

You can check what would be updated by running:

    contentdb --dry-run update
