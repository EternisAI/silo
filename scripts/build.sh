#!/bin/bash
set -e

VERSION=${VERSION:-dev}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

MAIN=${MAIN:-cmd/silo/main.go}
OUTPUT=${OUTPUT:-./bin/silo}

# Determine the package to inject version info into
if [[ "$MAIN" == *"silod"* ]]; then
    PKG="main"
    NAME="Silo Daemon"
else
    PKG="github.com/eternisai/silo/internal/cli"
    NAME="Silo CLI"
fi

LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X $PKG.version=$VERSION"
LDFLAGS="$LDFLAGS -X $PKG.commit=$COMMIT"
LDFLAGS="$LDFLAGS -X $PKG.buildDate=$BUILD_DATE"

echo "Building $NAME..."
echo "  Version:    $VERSION"
echo "  Commit:     $COMMIT"
echo "  Build Date: $BUILD_DATE"
echo "  Output:     $OUTPUT"

mkdir -p "$(dirname "$OUTPUT")"

go build -ldflags "$LDFLAGS" -o "$OUTPUT" "$MAIN"

echo "✓ Build complete: $OUTPUT"
