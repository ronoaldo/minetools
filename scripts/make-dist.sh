#!/bin/sh
set -e
# set -x

ARCHS="amd64 386 arm64"
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
        zip -q dist/contentdb-$GOOS-$GOARCH.zip dist/contentdb README.md LICENSE
	export FILES="$FILES dist/contentdb-$GOOS-$GOARCH.zip"
	rm -f dist/contentdb
    done
done

# Prepare a Github release
LAST=$(git tag --sort creatordate | tail -n 1)
read -p "New tag/version (latest is $LAST): " VERSION

CHANGELOG="$(mktemp)"
echo "# Changelog" > $CHANGELOG
echo >> $CHANGELOG
git log --pretty="* %an: %s" ${LAST}..HEAD >> $CHANGELOG
cat $CHANGELOG

echo "Drafting a Github release (uploading $FILES) ..."
gh release --repo ronoaldo/minetools create \
    --draft \
    --title "minetools $VERSION" \
    --notes-file $CHANGELOG \
    $VERSION \
    $FILES