#!/bin/sh

ARCHS="386 amd64 arm64"
OSES="linux darwin windows"

rm -rf dist
mkdir -p dist

for GOOS in $OSES; do
    for GOARCH in $ARCHS; do
        echo "Building for GOOS=$GOOS GOARCH=$GOARCH"
        GOOS=$GOOS GOARCH=$GOARCH go build -o dist/contentdb ./cmd/contentdb
        zip dist/contentdb-$GOOS-$GOARCH.zip dist/contentdb README.md LICENSE
    done
done