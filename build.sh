#!/bin/bash

# Configuration
APP_NAME="kuccps"
GO_ARCH="amd64"             # Can be amd64, 386, arm, arm64
GO_OS="linux"               # Target OS (Linux)
OUTPUT_DIR="../../bin"            # Output directory
ICON_FILE="app.ico"         # Optional icon file (Not applicable for Linux)
LD_FLAGS="-s -w"            # Linker flags to reduce binary size
BUILD_FLAGS="-trimpath"     # Additional build flags

# Create output directory if it doesn't exist
if [ ! -d "$OUTPUT_DIR" ]; then
    mkdir -p "$OUTPUT_DIR"
fi

# Check for Go installation
if ! command -v go &> /dev/null
then
    echo "Error: Go compiler not found in PATH"
    exit 1
fi

# Build the project
echo "Building the project..."
go get kuccps
go build $BUILD_FLAGS -ldflags="$LD_FLAGS" -o "$OUTPUT_DIR/$APP_NAME"

# Verify build
if [ -f "$OUTPUT_DIR/$APP_NAME" ]; then
    echo "Build successful!"
    echo "Output: $OUTPUT_DIR/$APP_NAME"
    echo "File size: $(stat -c %s "$OUTPUT_DIR/$APP_NAME") bytes"
else
    echo "Build failed!"
    exit 1
fi
