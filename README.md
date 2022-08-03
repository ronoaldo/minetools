# minetools

[![Go Reference](https://pkg.go.dev/badge/github.com/ronoaldo/minetools.svg)](https://pkg.go.dev/github.com/ronoaldo/minetools)

Project minetools has several small utilities to manage a Minetest server
from the command line.

**This is a work in progress.**

## contentdb

Command `contentdb` can be used to download mods from
https://content.minetest.net and install them into a local `mods` folder.

### Install pre-compiled binaries on Linux

From the Releases page you can download the file for your Linux machine.
Depending on your architecture, you may want to download either the 386,
amd64 or arm64 versions. To decide which one you need, try the following
command:

```
uname -m
```

Select contentdb-linux-**386**.zip for `x86`, contentdb-linux-**amd64**.zip for
`x86_64` and contentdb-linux-**arm64**.zip for `aarch64`.  Then you can download
with `curl` like this (change amd64 for your architecture):

```
curl -L https://github.com/ronoaldo/minetools/releases/download/v0.2.2/contentdb-linux-amd64.zip > /tmp/contentdb.zip
```

Next you need to unpack the zip and install the program on your `$PATH`.
One way to complete that step is to use the `unzip` program, like this:

```
cd /tmp/
unzip /tmp/contentdb.zip
sudo mv dist/contentdb /usr/local/bin
```

Check the installation is working with:

```
contentdb --help
```

If you don't have `unzip` or `curl`, you can install these tools using your
package manager like `apt` for Debian/Ubuntu or derivatives:

```
apt-get update
apt-get install curl unzip -yq
```

### Install from source

To install the contentdb cli, you need a working Go (> 1.16) toolchain:

    git clone https://github.com/ronoaldo/minetools
    go install ./minetools/cmd/contentdb

### Usage

To get the online help with all parameters, run

    contentdb --help

To search for content, use the `search` subcommand:

    contentdb search mesecons

This command will install the requested mod into the `mods` folder in the
current working directory.  For instance, to install a system-wide mod, first
change into that directory with:

    cd /usr/share/minetest

To install in your local home directory, first change to the `$HOME/.minetest` folder
with:

    cd $HOME/.minetest

To install a mod/modpack, use the `install` subcommand. 

    contentdb install rubenwardy/sfinv

Or alternativelly, specify a specific release to install:

    contentdb install rubenwardy/sfinv@52

To update all mods in the mods folder (including those installed with git!), use the `update` subdommand:

    contentdb update

You can check what would be updated by running:

    contentdb update --dry-run
