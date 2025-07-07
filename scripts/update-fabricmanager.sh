#!/bin/bash

# NVIDIA FabricManager Auto-Update Script
# This script automatically downloads and updates FabricManager versions

set -e

# Configuration
FABRICMANAGER_BASE_URL="https://developer.download.nvidia.com/compute/nvidia-driver/redist/fabricmanager"
ARCH="linux-x86_64"
TEMP_DIR="${TMPDIR:-/tmp}/fabricmanager-update"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to get the latest version from NVIDIA
get_latest_version() {
    log_info "Fetching latest FabricManager versions..."
    
    # Create temp directory
    mkdir -p "$TEMP_DIR"
    
    # Download the index page
    local index_url="$FABRICMANAGER_BASE_URL/$ARCH/"
    local index_file="$TEMP_DIR/index.html"
    
    if ! curl -s -o "$index_file" "$index_url"; then
        log_error "Failed to download index from $index_url"
        return 1
    fi
    
    # Extract version numbers from the index
    local versions=$(grep -o 'fabricmanager-linux-x86_64-[0-9]\+\.[0-9]\+\.[0-9]\+-archive\.tar\.xz' "$index_file" | \
                    sed 's/fabricmanager-linux-x86_64-\(.*\)-archive\.tar\.xz/\1/' | \
                    sort -V | tail -1)
    
    if [ -z "$versions" ]; then
        log_error "No versions found in index"
        return 1
    fi
    
    echo "$versions"
}

# Function to get current version from Git tags
get_current_version() {
    # Get the latest tag that matches our versioning pattern
    local latest_tag=$(git tag --list "v*" --sort=-version:refname | grep -E "^v[0-9]+\.[0-9]+\.[0-9]+-[0-9]+\.[0-9]+$" | head -1)
    
    if [ -z "$latest_tag" ]; then
        echo "none"
    else
        # Extract FabricManager version from tag (remove 'v' prefix and '-<X.Y>' suffix)
        echo "$latest_tag" | sed 's/^v//' | sed 's/-[0-9]*\.[0-9]*$//'
    fi
}

# Function to get all available versions from NVIDIA
get_all_versions() {
    log_info "Fetching all available FabricManager versions..."
    
    # Create temp directory
    mkdir -p "$TEMP_DIR"
    
    # Download the index page
    local index_url="$FABRICMANAGER_BASE_URL/$ARCH/"
    local index_file="$TEMP_DIR/index.html"
    
    if ! curl -s -o "$index_file" "$index_url"; then
        log_error "Failed to download index from $index_url"
        return 1
    fi
    
    # Extract all version numbers from the index
    grep -o 'fabricmanager-linux-x86_64-[0-9]\+\.[0-9]\+\.[0-9]\+-archive\.tar\.xz' "$index_file" | \
    sed 's/fabricmanager-linux-x86_64-\(.*\)-archive\.tar\.xz/\1/' | \
    sort -V
}

# Function to get all existing Git tags
get_existing_tags() {
    git tag --list "v*" --sort=version:refname | grep -E "^v[0-9]+\.[0-9]+\.[0-9]+-[0-9]+\.[0-9]+$" | sed 's/^v//' | sed 's/-[0-9]*\.[0-9]*$//'
}

# Function to find missing versions
find_missing_versions() {
    local current_version=$1
    
    log_info "Finding missing versions since $current_version..."
    
    # Get all available versions
    local all_versions=$(get_all_versions)
    
    # Get existing tags
    local existing_tags=$(get_existing_tags)
    
    # Find versions that are newer than current but not tagged
    local missing_versions=""
    
    while IFS= read -r version; do
        # Skip if version is older than or equal to current
        if [ "$(printf '%s\n' "$current_version" "$version" | sort -V | head -1)" != "$current_version" ]; then
            continue
        fi
        
        # Check if this version is already tagged
        if echo "$existing_tags" | grep -q "^$version$"; then
            continue
        fi
        
        # Add to missing versions
        if [ -z "$missing_versions" ]; then
            missing_versions="$version"
        else
            missing_versions="$missing_versions"$'\n'"$version"
        fi
    done <<< "$all_versions"
    
    echo "$missing_versions"
}

# Function to download and extract FabricManager
download_fabricmanager() {
    local version=$1
    local archive_name="fabricmanager-linux-x86_64-${version}-archive.tar.xz"
    local download_url="$FABRICMANAGER_BASE_URL/$ARCH/$archive_name"
    local archive_path="$TEMP_DIR/$archive_name"
    
    log_info "Downloading FabricManager version $version..."
    
    # Ensure temp directory exists and is writable
    mkdir -p "$TEMP_DIR"
    if [ ! -w "$TEMP_DIR" ]; then
        log_error "Temp directory $TEMP_DIR is not writable"
        return 1
    fi
    
    # Clean up any existing archive
    rm -f "$archive_path"
    
    # Download the archive with better error handling
    log_info "Downloading from: $download_url"
    if ! curl -L --fail --show-error -o "$archive_path" "$download_url"; then
        log_error "Failed to download $download_url"
        return 1
    fi
    
    # Verify the download
    if [ ! -f "$archive_path" ] || [ ! -s "$archive_path" ]; then
        log_error "Downloaded file is empty or missing"
        return 1
    fi
    
    log_success "Downloaded $archive_name ($(du -h "$archive_path" | cut -f1))"
    
    # Extract the archive
    log_info "Extracting archive..."
    if ! tar -xf "$archive_path" -C "$TEMP_DIR"; then
        log_error "Failed to extract archive"
        return 1
    fi
    
    # Find the extracted directory
    local extracted_dir=$(find "$TEMP_DIR" -maxdepth 1 -type d -name "fabricmanager-linux-x86_64-${version}-*" | head -1)
    
    if [ -z "$extracted_dir" ]; then
        log_error "Could not find extracted directory"
        log_info "Contents of $TEMP_DIR:"
        ls -la "$TEMP_DIR" || true
        return 1
    fi
    
    log_info "Found extracted directory: $extracted_dir"
    echo "$extracted_dir"
}

# Function to update headers directory
update_headers() {
    local extracted_dir=$1
    local version=$2
    
    log_info "Updating headers directory with version $version..."
    
    # Create headers directory if it doesn't exist
    mkdir -p "headers"
    
    # Backup current headers if they exist
    if [ -d "headers" ] && [ "$(ls -A headers 2>/dev/null)" ]; then
        local backup_dir="headers.backup.$(date +%Y%m%d_%H%M%S)"
        log_info "Backing up current headers to $backup_dir"
        if ! mv headers "$backup_dir"; then
            log_warning "Failed to backup headers, removing old headers"
            rm -rf headers
        fi
    fi
    
    # Create fresh headers directory
    mkdir -p "headers"
    
    # Copy new headers
    if [ -d "$extracted_dir/include" ]; then
        log_info "Copying headers from $extracted_dir/include"
        if ! cp -r "$extracted_dir/include"/* headers/; then
            log_error "Failed to copy headers"
            return 1
        fi
        log_success "Successfully copied headers"
    else
        log_error "No include directory found in extracted archive"
        log_info "Contents of $extracted_dir:"
        ls -la "$extracted_dir" || true
        log_info "Looking for include directory in subdirectories..."
        find "$extracted_dir" -type d -name "include" -exec ls -la {} \; || true
        return 1
    fi
    
    log_success "Updated headers to version $version"
}



# Function to update Go module version
update_go_version() {
    local version=$1
    
    log_info "Updating Go module version to $version..."
    
    # Update fabricmanager.go version constant
    if sed -i "s/Version = \"[^\"]*\"/Version = \"$version\"/" fabricmanager.go; then
        log_success "Updated fabricmanager.go version to $version"
    else
        log_warning "Failed to update fabricmanager.go version"
    fi
}

# Function to run tests after update
run_tests() {
    log_info "Running tests after update..."
    
    if CGO_ENABLED=1 go test -v ./...; then
        log_success "All tests passed"
    else
        log_warning "Some tests failed - please review"
        return 1
    fi
}

# Function to build after update
build_after_update() {
    log_info "Building after update..."
    
    if CGO_ENABLED=1 go build -o fmpm cmd/fmpm/main.go; then
        log_success "Build successful"
    else
        log_error "Build failed"
        return 1
    fi
}

# Function to create or checkout version-specific branch
create_version_branch() {
    local version=$1
    local branch_name="fm/$version"
    
    log_info "Setting up branch for version $version..."
    
    # Check if we're already on the correct branch
    local current_branch=$(git branch --show-current 2>/dev/null || git rev-parse --abbrev-ref HEAD)
    if [ "$current_branch" = "$branch_name" ]; then
        log_info "Already on branch $branch_name"
        return 0
    fi
    
    # Check if branch exists locally
    if git show-ref --verify --quiet refs/heads/"$branch_name"; then
        log_info "Checking out existing branch $branch_name"
        if ! git checkout "$branch_name"; then
            log_error "Failed to checkout branch $branch_name"
            return 1
        fi
    else
        # Check if branch exists remotely
        if git ls-remote --heads origin "$branch_name" | grep -q "$branch_name"; then
            log_info "Checking out remote branch $branch_name"
            if ! git checkout -b "$branch_name" "origin/$branch_name"; then
                log_error "Failed to checkout remote branch $branch_name"
                return 1
            fi
        else
            # Create new branch from main
            log_info "Creating new branch $branch_name from main"
            if ! git checkout main; then
                log_error "Failed to checkout main branch"
                return 1
            fi
            if ! git pull origin main; then
                log_warning "Failed to pull latest main, continuing anyway"
            fi
            if ! git checkout -b "$branch_name"; then
                log_error "Failed to create branch $branch_name"
                return 1
            fi
        fi
    fi
    
    log_success "Successfully set up branch $branch_name"
}

# Function to create version-specific tag and release
create_version_tag() {
    local version=$1
    
    # Find the next Go change number for this FabricManager version
    # Start with 1.0 for new versions, increment Y for fixes, X for features
    local next_x=1
    local next_y=0
    local existing_tags=$(git tag --list "v${version}-*" | sed 's/.*-//' | sort -V | tail -1)
    
    if [ -n "$existing_tags" ]; then
        # Parse existing X.Y format
        if [[ "$existing_tags" =~ ^([0-9]+)\.([0-9]+)$ ]]; then
            next_x=${BASH_REMATCH[1]}
            next_y=${BASH_REMATCH[2]}
            # For now, just increment Y (bug fixes)
            next_y=$((next_y + 1))
        else
            # Fallback: assume it's a single number and convert to X.Y
            next_x=$((existing_tags + 1))
            next_y=0
        fi
    fi
    
    local tag_name="v${version}-${next_x}.${next_y}"
    
    log_info "Creating release tag $tag_name..."
    
    # Check if tag already exists
    if git tag -l "$tag_name" | grep -q "$tag_name"; then
        log_warning "Tag $tag_name already exists"
        return 0
    fi
    
    # Create tag for current state (after updates)
    if git tag "$tag_name"; then
        log_success "Created release tag $tag_name"
        
        # Push tag
        if git push origin "$tag_name"; then
            log_success "Pushed release tag $tag_name"
        else
            log_warning "Failed to push tag"
        fi
    else
        log_warning "Failed to create tag"
    fi
}

# Function to check for API changes
check_api_changes() {
    local old_version=$1
    local new_version=$2
    
    log_info "Checking for API changes between $old_version and $new_version..."
    
    # This is a placeholder - in a real implementation, you might:
    # 1. Compare header files
    # 2. Check for new functions/constants
    # 3. Validate that existing bindings still work
    
    log_warning "API change detection not implemented - please review manually"
}

# Function to show version status
show_version_status() {
    local current_version=$(get_current_version)
    local latest_version=$(get_latest_version)
    
    echo "=== Version Status ==="
    echo "Current version: $current_version"
    echo "Latest available: $latest_version"
    
    if [ "$current_version" != "none" ] && [ "$current_version" != "$latest_version" ]; then
        echo ""
        echo "Missing versions:"
        find_missing_versions "$current_version"
    fi
}

# Main function
main() {
    local force_update=false
    local target_version=""
    local skip_tests=false
    local skip_build=false
    local create_tag=false
    local check_only=false
    local show_status=false
    local create_branch=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --force)
                force_update=true
                shift
                ;;
            --version)
                target_version="$2"
                shift 2
                ;;
            --skip-tests)
                skip_tests=true
                shift
                ;;
            --skip-build)
                skip_build=true
                shift
                ;;
            --create-tag)
                create_tag=true
                shift
                ;;
            --check-only)
                check_only=true
                shift
                ;;
            --status)
                show_status=true
                shift
                ;;
            --create-branch)
                create_branch=true
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --force        Force update even if version is the same"
                echo "  --version VER  Update to specific version"
                echo "  --skip-tests   Skip running tests after update"
                echo "  --skip-build   Skip building after update"
                echo "  --create-tag   Create git tag for the new version"
                echo "  --create-branch Create/checkout version-specific branch (fm/VERSION)"
                echo "  --check-only   Only check for new versions, don't update"
                echo "  --status       Show current version status"
                echo "  --help         Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Show status if requested
    if [ "$show_status" = true ]; then
        show_version_status
        exit 0
    fi
    
    # Get current version
    local current_version=$(get_current_version)
    log_info "Current FabricManager version: $current_version"
    
    # Get target version
    if [ -z "$target_version" ]; then
        target_version=$(get_latest_version)
        if [ $? -ne 0 ]; then
            log_error "Failed to get latest version"
            exit 1
        fi
    fi
    
    log_info "Target FabricManager version: $target_version"
    
    # Check if update is needed
    if [ "$current_version" = "$target_version" ] && [ "$force_update" = false ]; then
        log_success "Already at latest version $target_version"
        exit 0
    fi
    
    # If check-only mode, just report the version difference
    if [ "$check_only" = true ]; then
        if [ "$current_version" != "$target_version" ]; then
            log_info "New version available: $current_version -> $target_version"
            exit 0
        else
            log_info "No new version available"
            exit 1
        fi
    fi
    
    # Create or checkout version branch if requested
    if [ "$create_branch" = true ]; then
        create_version_branch "$target_version"
        if [ $? -ne 0 ]; then
            log_error "Failed to set up version branch"
            exit 1
        fi
    fi
    
    # Download and extract new version
    local extracted_dir=$(download_fabricmanager "$target_version")
    if [ $? -ne 0 ]; then
        log_error "Failed to download FabricManager"
        exit 1
    fi
    
    # Update headers directory
    update_headers "$extracted_dir" "$target_version"
    if [ $? -ne 0 ]; then
        log_error "Failed to update headers directory"
        exit 1
    fi
    
    # Update Go module version
    update_go_version "$target_version"
    
    # Check for API changes
    if [ "$current_version" != "none" ]; then
        check_api_changes "$current_version" "$target_version"
    fi
    
    # Run tests
    if [ "$skip_tests" = false ]; then
        run_tests
    fi
    
    # Build
    if [ "$skip_build" = false ]; then
        build_after_update
    fi
    
    # Create tag
    if [ "$create_tag" = true ]; then
        create_version_tag "$target_version"
    fi
    
    # Cleanup
    rm -rf "$TEMP_DIR"
    
    log_success "Successfully updated to FabricManager version $target_version"
}

# Run main function
main "$@" 