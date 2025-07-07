#!/usr/bin/env bash
set -e

# Default values
LIBDIR=".fabricmanager-lib"
VERSION=""

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --version)
      VERSION="$2"
      shift 2
      ;;
    --libdir)
      LIBDIR="$2"
      shift 2
      ;;
    --help)
      echo "Usage: $0 --version <FM_VERSION> [--libdir <dir>]"
      exit 0
      ;;
    *)
      echo "Unknown argument: $1"
      exit 1
      ;;
  esac
done

if [[ -z "$VERSION" ]]; then
  echo "Error: --version is required" >&2
  exit 1
fi

ARCHIVE_URL="https://developer.download.nvidia.com/compute/nvidia-driver/redist/fabricmanager/linux-x86_64/fabricmanager-linux-x86_64-${VERSION}-archive.tar.xz"
ARCHIVE_NAME="fabricmanager-linux-x86_64-${VERSION}-archive.tar.xz"
EXTRACTED_DIR="fabricmanager-linux-x86_64-${VERSION}-archive"

# Download
echo "Downloading FabricManager runtime for version $VERSION..."
curl -L -o "$ARCHIVE_NAME" "$ARCHIVE_URL"

# Extract
rm -rf "$EXTRACTED_DIR"
tar -xf "$ARCHIVE_NAME"

# Install to user-writable directory
mkdir -p "$LIBDIR"
cp "$EXTRACTED_DIR/lib/libnvfm.so.1" "$LIBDIR/"
cp -P "$EXTRACTED_DIR/lib/libnvfm.so" "$LIBDIR/"

# Clean up
rm -rf "$ARCHIVE_NAME" "$EXTRACTED_DIR"

# Print export instructions
echo ""
echo "✅ FabricManager runtime installed to: $LIBDIR"
echo "To use it for build and run, export these variables:"
echo "  export LD_LIBRARY_PATH=\"$PWD/$LIBDIR:\$LD_LIBRARY_PATH\""
echo "  export CGO_LDFLAGS=\"-L$PWD/$LIBDIR\""
echo ""
echo "Example build:"
echo "  CGO_LDFLAGS=\"-L$PWD/$LIBDIR\" LD_LIBRARY_PATH=\"$PWD/$LIBDIR\" CGO_ENABLED=1 go build -o fmpm cmd/fmpm/main.go"
echo ""
echo "Done." 