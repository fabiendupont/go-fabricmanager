#!/bin/bash

# Simple version status script using Git tags
# This script shows the current version status and missing versions

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== NVIDIA FabricManager Go Package Version Status ===${NC}"
echo

# Get current version from Git tags
current_version=$(git tag --list "v*" --sort=-version:refname | grep -E "^v[0-9]+\.[0-9]+\.[0-9]+-[0-9]+\.[0-9]+$" | head -1 | sed 's/^v//' | sed 's/-[0-9]*\.[0-9]*$//')

if [ -z "$current_version" ]; then
    echo -e "${YELLOW}Current version: None (no tags found)${NC}"
else
    echo -e "${GREEN}Current version: $current_version${NC}"
fi

echo

# Get all existing tags
echo -e "${BLUE}Existing version tags:${NC}"
existing_tags=$(git tag --list "v*" --sort=version:refname | grep -E "^v[0-9]+\.[0-9]+\.[0-9]+-[0-9]+\.[0-9]+$" | sed 's/^v//' | sed 's/-[0-9]*\.[0-9]*$//')

if [ -z "$existing_tags" ]; then
    echo "  None"
else
    echo "$existing_tags" | sed 's/^/  /'
fi

echo

# Check for missing versions if we have a current version
if [ -n "$current_version" ]; then
    echo -e "${BLUE}Checking for missing versions since $current_version...${NC}"
    
    # Run the update script to find missing versions
    missing_versions=$(./scripts/update-fabricmanager.sh --status 2>/dev/null | grep -A 100 "Missing versions:" | tail -n +2)
    
    if [ -n "$missing_versions" ]; then
        echo -e "${YELLOW}Missing versions:${NC}"
        echo "$missing_versions" | sed 's/^/  /'
    else
        echo -e "${GREEN}No missing versions found${NC}"
    fi
fi

echo

# Show next steps
echo -e "${BLUE}Next steps:${NC}"
echo "  • Check for new versions: ./scripts/update-fabricmanager.sh --check-only"
echo "  • Update to latest: ./scripts/update-fabricmanager.sh --create-tag"
echo "  • Update to specific version: ./scripts/update-fabricmanager.sh --version <version> --create-tag" 