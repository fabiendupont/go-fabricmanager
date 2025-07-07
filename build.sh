#!/bin/bash

# Build script for NVIDIA FabricManager Go Package

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
FABRICMANAGER_BASE_URL="https://developer.download.nvidia.com/compute/nvidia-driver/redist/fabricmanager"
ARCH="linux-x86_64"
TEMP_DIR="tmp"
RUNTIME_LIB_DIR="$TEMP_DIR/runtime-libs"

echo -e "${BLUE}=== Building NVIDIA FabricManager Go Package ===${NC}"

# Check if CGO is enabled
if [ "$CGO_ENABLED" != "1" ]; then
    echo -e "${YELLOW}Warning: CGO_ENABLED is not set to 1${NC}"
    echo -e "${YELLOW}Setting CGO_ENABLED=1 for this build...${NC}"
    export CGO_ENABLED=1
fi

# Check if headers exist
if [ ! -f "headers/nv_fm_agent.h" ] || [ ! -f "headers/nv_fm_types.h" ]; then
    echo -e "${RED}Error: Required headers not found in headers/ directory${NC}"
    echo -e "${YELLOW}Please ensure the headers are present:${NC}"
    echo "  - headers/nv_fm_agent.h"
    echo "  - headers/nv_fm_types.h"
    exit 1
fi

echo -e "${GREEN}✓ Headers found${NC}"

# Function to download runtime library
download_runtime_lib() {
    local version=$1
    
    echo -e "${BLUE}Downloading FabricManager runtime library version $version...${NC}"
    
    # Create temp directory
    mkdir -p "$RUNTIME_LIB_DIR"
    
    # Download the archive
    local archive_name="fabricmanager-linux-x86_64-${version}-archive.tar.xz"
    local download_url="$FABRICMANAGER_BASE_URL/$ARCH/$archive_name"
    local archive_path="$TEMP_DIR/$archive_name"
    
    if ! curl -L -o "$archive_path" "$download_url"; then
        echo -e "${RED}Failed to download $download_url${NC}"
        return 1
    fi
    
    # Extract the archive
    if ! tar -xf "$archive_path" -C "$TEMP_DIR"; then
        echo -e "${RED}Failed to extract archive${NC}"
        return 1
    fi
    
    # Find the extracted directory
    local extracted_dir=$(find "$TEMP_DIR" -maxdepth 1 -type d -name "fabricmanager-linux-x86_64-${version}-*" | head -1)
    
    if [ -z "$extracted_dir" ]; then
        echo -e "${RED}Could not find extracted directory${NC}"
        return 1
    fi
    
    # Copy the library
    if [ -f "$extracted_dir/lib/libnvfm.so" ]; then
        cp "$extracted_dir/lib/libnvfm.so" "$RUNTIME_LIB_DIR/"
        echo -e "${GREEN}✓ Runtime library downloaded to $RUNTIME_LIB_DIR/libnvfm.so${NC}"
    else
        echo -e "${RED}Library file not found in extracted archive${NC}"
        return 1
    fi
    
    # Clean up
    rm -rf "$extracted_dir" "$archive_path"
}

# Function to detect current FabricManager version from headers
detect_fabricmanager_version() {
    # Try to extract version from header files
    local version=$(grep -r "FM_VERSION\|FABRICMANAGER_VERSION" headers/ 2>/dev/null | head -1 | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1)
    
    if [ -z "$version" ]; then
        # Fallback: try to get latest version from NVIDIA
        curl -s "$FABRICMANAGER_BASE_URL/$ARCH/" | grep -o 'fabricmanager-linux-x86_64-[0-9]\+\.[0-9]\+\.[0-9]\+-archive\.tar\.xz' | sed 's/fabricmanager-linux-x86_64-\(.*\)-archive\.tar\.xz/\1/' | sort -V | tail -1
    else
        echo "$version"
    fi
}

# Try to build with system library first
echo -e "${BLUE}Attempting build with system library...${NC}"

if CGO_ENABLED=1 go build -o fmpm cmd/fmpm/main.go; then
    echo -e "${GREEN}✓ Build successful with system library!${NC}"
    echo -e "${BLUE}Binary created: ./fmpm${NC}"
else
    echo -e "${YELLOW}System library not found, downloading runtime library...${NC}"
    
    detected_version=$(detect_fabricmanager_version)
    if [ -z "$detected_version" ]; then
        echo -e "${YELLOW}Could not detect version from headers, trying latest version...${NC}"
        detected_version=$(curl -s "$FABRICMANAGER_BASE_URL/$ARCH/" | grep -o 'fabricmanager-linux-x86_64-[0-9]\+\.[0-9]\+\.[0-9]\+-archive\.tar\.xz' | sed 's/fabricmanager-linux-x86_64-\(.*\)-archive\.tar\.xz/\1/' | sort -V | tail -1)
    fi
    echo -e "${BLUE}Detected version: $detected_version${NC}"
    
    if download_runtime_lib "$detected_version"; then
        # Set library path for build
        export LD_LIBRARY_PATH="$RUNTIME_LIB_DIR:$LD_LIBRARY_PATH"
        export LIBRARY_PATH="$RUNTIME_LIB_DIR:$LIBRARY_PATH"
        
        echo -e "${BLUE}Building with downloaded runtime library...${NC}"
        
        if CGO_ENABLED=1 go build -o fmpm cmd/fmpm/main.go; then
            echo -e "${GREEN}✓ Build successful with downloaded library!${NC}"
            echo -e "${BLUE}Binary created: ./fmpm${NC}"
            echo -e "${YELLOW}Note: Runtime library is in $RUNTIME_LIB_DIR/ (will be cleaned up)${NC}"
        else
            echo -e "${RED}✗ Build failed even with downloaded library${NC}"
            exit 1
        fi
    else
        echo -e "${RED}✗ Failed to download runtime library${NC}"
        exit 1
    fi
fi

# Test the binary
if [ -f "./fmpm" ]; then
    echo -e "${BLUE}Testing binary...${NC}"
    if ./fmpm --help > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Binary works correctly${NC}"
    else
        echo -e "${YELLOW}⚠ Binary created but may have runtime issues${NC}"
        echo -e "${YELLOW}This is normal if the FabricManager runtime library is not installed${NC}"
    fi
fi

# Clean up temporary files
if [ -d "$TEMP_DIR" ]; then
    echo -e "${BLUE}Cleaning up temporary files...${NC}"
    rm -rf "$TEMP_DIR"
fi

echo ""
echo -e "${GREEN}Build completed!${NC}" 