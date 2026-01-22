#!/bin/bash
set -e

VERSION=${VERSION:-dev}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X github.com/eternisai/silo/internal/cli.version=$VERSION"
LDFLAGS="$LDFLAGS -X github.com/eternisai/silo/internal/cli.commit=$COMMIT"
LDFLAGS="$LDFLAGS -X github.com/eternisai/silo/internal/cli.buildDate=$BUILD_DATE"

OUTPUT=${OUTPUT:-./bin/silo}

echo "Building Silo CLI..."
echo "  Version:    $VERSION"
echo "  Commit:     $COMMIT"
echo "  Build Date: $BUILD_DATE"
echo "  Output:     $OUTPUT"

mkdir -p "$(dirname "$OUTPUT")"

go build -ldflags "$LDFLAGS" -o "$OUTPUT" ./cmd/silo

echo "✓ Build complete: $OUTPUT"
