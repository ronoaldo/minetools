#!/bin/sh
set -e
set -x

ARCHS="amd64 386"
OSES="linux windows"

# Cleanup previous builds
rm -rf dist
mkdir -p dist

# Build for all os/arch combos
FILES=""
for GOOS in $OSES; do
    for GOARCH in $ARCHS; do
        echo "Building for GOOS=$GOOS GOARCH=$GOARCH"
        GOOS=$GOOS GOARCH=$GOARCH go build -o dist/contentdb ./cmd/contentdb
        zip dist/contentdb-$GOOS-$GOARCH.zip dist/contentdb README.md LICENSE
	export FILES="$FILES dist/contentdb-$GOOS-$GOARCH.zip"
	rm -f dist/contentdb
    done
done

# Prepare a Github release
git tag --sort creatordate
LAST=$(git tag --sort creatordate | tail -n 1)
read -p "Release version (latest is $LAST): " VERSION
CHANGES="$(mktemp)"
git log --pretty="* %an: %s" ${LAST}..HEAD > $CHANGES

echo "Releasing '$FILES' ..."

gh release --repo ronoaldo/minetools create \
    --draft \
    --notes-file $CHANGES \
    $VERSION \
    $FILES