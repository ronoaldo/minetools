# minetools

Project minetools has several small command line utilities to manage a
Minetest server from the command line.

This is a work in progress.

## contentdb

Command `contentdb` can be used to download mods from
https://content.minetest.net and install them into the local `mods` folder.


### Install

To install, you need a working Go (> 1.16) toolchain:

    go get github.com/ronoaldo/minetools/cmd/contentdb