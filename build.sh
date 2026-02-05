#!/bin/bash
# Build script for Tucha
# Produces both Linux and Windows builds
#
# Usage:
#   ./build.sh [version]
#
# If version is not provided, it will be extracted from git tags or default to 0.0.0

set -e

echo "======================================"
echo "Building Tucha (Linux + Windows)"
echo "======================================"
echo ""

# Step 1: Determine version
echo "[1/6] Determining version..."
if [ -n "$1" ]; then
    VERSION="$1"
    echo "[OK] Using provided version: $VERSION"
elif git describe --tags --abbrev=0 >/dev/null 2>&1; then
    VERSION=$(git describe --tags --abbrev=0 | sed 's/^v//')
    echo "[OK] Version from git tag: $VERSION"
else
    VERSION="0.0.0"
    echo "[!] No git tag found, using default: $VERSION"
fi

# Convert to 4-component format for Windows resources
DOT_COUNT=$(echo "$VERSION" | tr -cd '.' | wc -c)
if [ "$DOT_COUNT" -eq 2 ]; then
    VERSION_FULL="${VERSION}.0"
elif [ "$DOT_COUNT" -eq 3 ]; then
    VERSION_FULL="${VERSION}"
else
    VERSION_FULL="${VERSION}.0.0.0"
fi
echo "    Version (4-component): $VERSION_FULL"
echo ""

# Step 2: Cleanup old builds
echo "[2/6] Cleaning old builds..."
rm -f cmd/tucha/*.syso
rm -f tucha tucha.exe
rm -f winres/*.syso
echo "[OK] Cleanup complete"
echo ""

# Step 3: Check for go-winres
echo "[3/6] Checking for go-winres..."
WINRES_CMD=""

if command -v go-winres &> /dev/null; then
    WINRES_CMD="go-winres"
    echo "[OK] go-winres found in PATH"
elif [ -f "$HOME/go/bin/go-winres" ]; then
    WINRES_CMD="$HOME/go/bin/go-winres"
    echo "[OK] go-winres found in $HOME/go/bin"
else
    echo "[!] go-winres not found"
    echo ""
    echo "Installing go-winres..."
    if go install github.com/tc-hib/go-winres@latest; then
        WINRES_CMD="$HOME/go/bin/go-winres"
        echo "[OK] go-winres installed successfully"
    else
        echo "[X] Failed to install go-winres"
        echo ""
        echo "Please install manually:"
        echo "  go install github.com/tc-hib/go-winres@latest"
        echo ""
        exit 1
    fi
fi
echo ""

# Step 4: Update winres.json with version and generate resources
echo "[4/6] Generating Windows resources..."
if [ ! -f "winres/winres.json" ]; then
    echo "[!] winres/winres.json not found"
    echo "    Skipping resource generation"
else
    # Update version in winres.json
    if command -v jq &> /dev/null; then
        cp winres/winres.json winres/winres.json.bak
        cat winres/winres.json.bak | \
        jq ".RT_MANIFEST[\"#1\"][\"0409\"].identity.version = \"$VERSION_FULL\" | \
            .RT_VERSION[\"#1\"][\"0000\"].fixed.file_version = \"$VERSION_FULL\" | \
            .RT_VERSION[\"#1\"][\"0000\"].fixed.product_version = \"$VERSION_FULL\" | \
            .RT_VERSION[\"#1\"][\"0000\"].info[\"0409\"].FileVersion = \"$VERSION\" | \
            .RT_VERSION[\"#1\"][\"0000\"].info[\"0409\"].ProductVersion = \"$VERSION\"" \
        > winres/winres.json
        echo "[OK] Updated winres.json with version $VERSION"
    else
        echo "[!] jq not found, skipping version update in winres.json"
    fi

    # Generate resources
    if $WINRES_CMD make --in winres/winres.json --out cmd/tucha/rsrc 2>&1; then
        echo "[OK] Resource files generated in cmd/tucha/"
    else
        echo "[!] Warning: go-winres failed"
        echo "    Continuing without embedded resources"
    fi

    # Restore original winres.json
    if [ -f "winres/winres.json.bak" ]; then
        mv winres/winres.json.bak winres/winres.json
    fi
fi
echo ""

# Step 5: Build Linux executable
echo "[5/6] Building Linux executable..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o tucha ./cmd/tucha
echo "[OK] Linux build complete: tucha"
echo ""

# Step 6: Build Windows executable
echo "[6/6] Building Windows executable..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o tucha.exe ./cmd/tucha
echo "[OK] Windows build complete: tucha.exe"
echo ""

# Cleanup intermediate files
rm -f cmd/tucha/*.syso
rm -f winres/*.syso

# Summary
echo "======================================"
echo "Build Summary (v$VERSION)"
echo "======================================"
ls -lh tucha tucha.exe 2>/dev/null || true
echo ""
echo "[OK] Build complete!"
