#!/bin/bash

# NVIDIA FabricManager Coverage Analysis Script
# Analyzes coverage between C API and Go bindings, and between CLI tools

set -e

# Configuration
HEADERS_DIR="headers"
GO_PACKAGE_FILE="fabricmanager.go"
CLI_MAIN_FILE="cmd/fmpm/main.go"
OUTPUT_DIR="coverage"
TEMP_DIR="/tmp/fabricmanager-coverage"

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

# Function to extract C API symbols from headers
extract_c_api() {
    local output_file="$1"
    
    log_info "Extracting C API symbols from headers..."
    
    # Create temp directory
    mkdir -p "$TEMP_DIR"
    
    # Find all header files
    local header_files=$(find "$HEADERS_DIR" -name "*.h" 2>/dev/null || true)
    
    if [ -z "$header_files" ]; then
        log_warning "No header files found in $HEADERS_DIR"
        return 1
    fi
    
    # Extract function declarations, types, and constants
    {
        # Functions
        for header in $header_files; do
            grep -h "^[a-zA-Z_][a-zA-Z0-9_]*\s\+[a-zA-Z_][a-zA-Z0-9_]*\s*(" "$header" 2>/dev/null || true
        done
        
        # Type definitions
        for header in $header_files; do
            grep -h "^typedef\s\+.*\s\+[a-zA-Z_][a-zA-Z0-9_]*\s*;" "$header" 2>/dev/null || true
        done
        
        # Constants and macros
        for header in $header_files; do
            grep -h "^#define\s\+[a-zA-Z_][a-zA-Z0-9_]*" "$header" 2>/dev/null || true
        done
        
        # Struct definitions
        for header in $header_files; do
            grep -h "^struct\s\+[a-zA-Z_][a-zA-Z0-9_]*" "$header" 2>/dev/null || true
        done
    } | sed 's/^[[:space:]]*//' | sort -u > "$output_file"
    
    log_success "Extracted $(wc -l < "$output_file") C API symbols"
}

# Function to extract Go API symbols
extract_go_api() {
    local output_file="$1"
    
    log_info "Extracting Go API symbols..."
    
    # Extract Go functions, types, and constants
    {
        # Functions
        grep -h "^func\s\+[A-Z][a-zA-Z0-9_]*" "$GO_PACKAGE_FILE" 2>/dev/null || true
        
        # Types
        grep -h "^type\s\+[A-Z][a-zA-Z0-9_]*" "$GO_PACKAGE_FILE" 2>/dev/null || true
        
        # Constants
        grep -h "^const\s\+[A-Z][a-zA-Z0-9_]*" "$GO_PACKAGE_FILE" 2>/dev/null || true
        
        # Variables
        grep -h "^var\s\+[A-Z][a-zA-Z0-9_]*" "$GO_PACKAGE_FILE" 2>/dev/null || true
    } | sed 's/^[[:space:]]*//' | sort -u > "$output_file"
    
    log_success "Extracted $(wc -l < "$output_file") Go API symbols"
}

# Function to extract CLI commands
extract_cli_commands() {
    local output_file="$1"
    
    log_info "Extracting CLI commands..."
    
    # Extract command flags and options from Go CLI
    {
        # Short flags (-a, -d, -l, etc.)
        grep -h "flag\.String\|flag\.Bool\|flag\.Int" "$CLI_MAIN_FILE" 2>/dev/null || true
        
        # Long flags (--hostname, etc.)
        grep -h "flag\.String\|flag\.Bool\|flag\.Int" "$CLI_MAIN_FILE" 2>/dev/null || true
    } | grep -o '"[^"]*"' | sed 's/"//g' | sort -u > "$output_file"
    
    log_success "Extracted $(wc -l < "$output_file") CLI commands"
}

# Function to run CLI comparison if version is provided
run_cli_comparison() {
    local version="$1"
    
    if [ -n "$version" ]; then
        log_info "Running CLI comparison for version $version..."
        if ./scripts/cli-comparison.sh --version "$version" > /dev/null 2>&1; then
            log_success "CLI comparison completed"
        else
            log_warning "CLI comparison failed"
        fi
    fi
}

# Function to generate coverage report
generate_coverage_report() {
    local c_api_file="$TEMP_DIR/c_api.txt"
    local go_api_file="$TEMP_DIR/go_api.txt"
    local cli_commands_file="$TEMP_DIR/cli_commands.txt"
    local report_file="$OUTPUT_DIR/coverage-report.md"
    
    log_info "Generating coverage report..."
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR"
    
    # Generate report
    cat > "$report_file" << EOF
# NVIDIA FabricManager Go Package Coverage Report

Generated on: $(date)

## C API Coverage

### Summary
- **Total C API symbols:** $(wc -l < "$c_api_file")
- **Total Go API symbols:** $(wc -l < "$go_api_file")
- **Coverage percentage:** $(calculate_coverage "$c_api_file" "$go_api_file")%

### Missing C API Symbols (Not Wrapped in Go)
\`\`\`
$(comm -23 "$c_api_file" "$go_api_file" 2>/dev/null || echo "None found")
\`\`\`

### Go API Symbols (No Corresponding C Symbol)
\`\`\`
$(comm -13 "$c_api_file" "$go_api_file" 2>/dev/null || echo "None found")
\`\`\`

## CLI Coverage

### Summary
- **Total CLI commands:** $(wc -l < "$cli_commands_file")

### Available CLI Commands
\`\`\`
$(cat "$cli_commands_file" 2>/dev/null || echo "None found")
\`\`\`

## Recommendations

$(generate_recommendations "$c_api_file" "$go_api_file")

---

*This report was automatically generated by the coverage analysis script.*
EOF
    
    log_success "Coverage report generated: $report_file"
}

# Function to calculate coverage percentage
calculate_coverage() {
    local c_file="$1"
    local go_file="$2"
    
    local c_count=$(wc -l < "$c_file")
    local go_count=$(wc -l < "$go_file")
    
    if [ "$c_count" -eq 0 ]; then
        echo "0"
    else
        echo "scale=1; $go_count * 100 / $c_count" | bc -l 2>/dev/null || echo "0"
    fi
}

# Function to generate recommendations
generate_recommendations() {
    local c_file="$1"
    local go_file="$2"
    
    local missing_count=$(comm -23 "$c_file" "$go_file" 2>/dev/null | wc -l)
    
    if [ "$missing_count" -gt 0 ]; then
        echo "- **High Priority:** $missing_count C API symbols are not wrapped in Go"
        echo "- Consider implementing missing C API wrappers"
    else
        echo "- **Excellent:** All C API symbols are wrapped in Go"
    fi
    
    echo "- Run tests to ensure Go wrappers work correctly"
    echo "- Update this report when adding new API bindings"
}

# Function to generate JSON report for CI
generate_json_report() {
    local c_api_file="$TEMP_DIR/c_api.txt"
    local go_api_file="$TEMP_DIR/go_api.txt"
    local json_file="$OUTPUT_DIR/coverage.json"
    
    log_info "Generating JSON report for CI..."
    
    local c_count=$(wc -l < "$c_api_file")
    local go_count=$(wc -l < "$go_api_file")
    local coverage_percent=$(calculate_coverage "$c_api_file" "$go_api_file")
    
    cat > "$json_file" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "c_api_count": $c_count,
  "go_api_count": $go_count,
  "coverage_percentage": $coverage_percent,
  "missing_c_symbols": [
$(comm -23 "$c_api_file" "$go_api_file" 2>/dev/null | sed 's/^/    "/; s/$/",/' | sed '$ s/,$//' || echo "")
  ],
  "extra_go_symbols": [
$(comm -13 "$c_api_file" "$go_api_file" 2>/dev/null | sed 's/^/    "/; s/$/",/' | sed '$ s/,$//' || echo "")
  ]
}
EOF
    
    log_success "JSON report generated: $json_file"
}

# Function to update README with coverage badge
update_readme_coverage() {
    local json_file="$OUTPUT_DIR/coverage.json"
    
    if [ ! -f "$json_file" ]; then
        log_warning "JSON report not found, skipping README update"
        return
    fi
    
    local coverage_percent=$(jq -r '.coverage_percentage' "$json_file" 2>/dev/null || echo "0")
    
    log_info "Updating README with coverage information..."
    
    # Create coverage badge URL (using shields.io)
    local badge_url="https://img.shields.io/badge/API%20Coverage-${coverage_percent}%25-brightgreen"
    
    # Update README if it exists
    if [ -f "README.md" ]; then
        # Add coverage badge after the title
        if ! grep -q "API Coverage" README.md; then
            sed -i "1a\\
![API Coverage]($badge_url)\\
" README.md
        else
            # Update existing badge
            sed -i "s|https://img.shields.io/badge/API%20Coverage-[0-9.]*%25-[a-z]*|$badge_url|" README.md
        fi
        
        log_success "Updated README with coverage badge"
    fi
}

# Function to show coverage summary
show_coverage_summary() {
    local json_file="$OUTPUT_DIR/coverage.json"
    
    if [ ! -f "$json_file" ]; then
        log_error "Coverage report not found. Run analysis first."
        return 1
    fi
    
    echo "=== Coverage Summary ==="
    echo "C API symbols: $(jq -r '.c_api_count' "$json_file")"
    echo "Go API symbols: $(jq -r '.go_api_count' "$json_file")"
    echo "Coverage: $(jq -r '.coverage_percentage' "$json_file")%"
    echo "Missing C symbols: $(jq -r '.missing_c_symbols | length' "$json_file")"
    echo "Extra Go symbols: $(jq -r '.extra_go_symbols | length' "$json_file")"
}

# Main function
main() {
    local generate_json=false
    local update_readme=false
    local show_summary=false
    local clean_temp=false
    local cli_version=""
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --json)
                generate_json=true
                shift
                ;;
            --update-readme)
                update_readme=true
                shift
                ;;
            --summary)
                show_summary=true
                shift
                ;;
            --clean)
                clean_temp=true
                shift
                ;;
            --cli-version)
                cli_version="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --json          Generate JSON report for CI"
                echo "  --update-readme Update README with coverage badge"
                echo "  --summary       Show coverage summary"
                echo "  --clean         Clean temporary files"
                echo "  --cli-version   Run CLI comparison with specific version"
                echo "  --help          Show this help message"
                echo ""
                echo "Examples:"
                echo "  $0 --json --update-readme                    # Full analysis"
                echo "  $0 --cli-version 575.57.08                  # Include CLI comparison"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Clean temp files if requested
    if [ "$clean_temp" = true ]; then
        rm -rf "$TEMP_DIR"
        log_success "Cleaned temporary files"
        exit 0
    fi
    
    # Show summary if requested
    if [ "$show_summary" = true ]; then
        show_coverage_summary
        exit 0
    fi
    
    # Create temp directory
    mkdir -p "$TEMP_DIR"
    
    # Extract APIs
    extract_c_api "$TEMP_DIR/c_api.txt"
    extract_go_api "$TEMP_DIR/go_api.txt"
    extract_cli_commands "$TEMP_DIR/cli_commands.txt"
    
    # Generate reports
    generate_coverage_report
    
    if [ "$generate_json" = true ]; then
        generate_json_report
    fi
    
    if [ "$update_readme" = true ]; then
        update_readme_coverage
    fi
    
    # Run CLI comparison if version provided
    run_cli_comparison "$cli_version"
    
    # Show summary
    show_coverage_summary
    
    log_success "Coverage analysis completed"
}

# Run main function
main "$@" 