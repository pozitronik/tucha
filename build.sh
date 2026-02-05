#!/bin/bash
# Build script for Tucha
# Produces both Linux and Windows builds
#
# Usage:
#   ./build.sh

set -e

echo "======================================"
echo "Building Tucha (Linux + Windows)"
echo "======================================"
echo ""

# Step 1: Cleanup old builds
echo "[1/5] Cleaning old builds..."
rm -f cmd/tucha/*.syso
rm -f tucha tucha.exe
rm -f winres/*.syso
echo "[OK] Cleanup complete"
echo ""

# Step 2: Check for go-winres
echo "[2/5] Checking for go-winres..."
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

# Step 3: Generate Windows resources (.syso files)
echo "[3/5] Generating Windows resources..."
if [ ! -f "winres/winres.json" ]; then
    echo "[!] winres/winres.json not found"
    echo "    Skipping resource generation"
    echo ""
else
    if $WINRES_CMD make --in winres/winres.json --out cmd/tucha/rsrc 2>&1; then
        echo "[OK] Resource files generated in cmd/tucha/"
    else
        echo "[!] Warning: go-winres failed"
        echo "    Continuing without embedded resources"
    fi
fi
echo ""

# Step 4: Build Linux executable
echo "[4/5] Building Linux executable..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o tucha ./cmd/tucha
echo "[OK] Linux build complete: tucha"
echo ""

# Step 5: Build Windows executable
echo "[5/5] Building Windows executable..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o tucha.exe ./cmd/tucha
echo "[OK] Windows build complete: tucha.exe"
echo ""

# Cleanup intermediate files
rm -f cmd/tucha/*.syso
rm -f winres/*.syso

# Summary
echo "======================================"
echo "Build Summary"
echo "======================================"
ls -lh tucha tucha.exe 2>/dev/null || true
echo ""
echo "[OK] Build complete!"
