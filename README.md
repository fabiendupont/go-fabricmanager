# NVIDIA FabricManager Go Package

![API Coverage](https://img.shields.io/badge/API%20Coverage-0%25-red)

A Go package and CLI tool for managing GPU partitions for NVIDIA Fabric Manager's Shared NVSwitch feature.

## Overview

This package provides a Go interface to NVIDIA FabricManager, allowing you to:
- List available fabric partitions
- Activate and deactivate partitions
- Query NVLink failed devices
- List unsupported partitions
- Set activated partition lists for resiliency mode

## Installation

### For a Specific FabricManager Version

```bash
# Install for FabricManager 575.57.08
go get github.com/NVIDIA/go-fabricmanager@fm/575.57.08

# Or use a specific release tag
go get github.com/NVIDIA/go-fabricmanager@v575.57.08-1.0
```

### For Latest Development Version

```bash
# Install from main branch (development)
go get github.com/NVIDIA/go-fabricmanager@main
```

## Prerequisites

- NVIDIA FabricManager runtime package installed (for runtime library)
- Go 1.22 or later
- CGO enabled (required for C library bindings)

### Runtime Dependencies

The Go package requires the NVIDIA FabricManager runtime library at runtime:

**Ubuntu/Debian:**
```bash
sudo apt-get install nvidia-fabricmanager-<version>
```

**RHEL/CentOS:**
```bash
sudo yum install nvidia-fabricmanager-<version>
```

**Note:** The development package is NOT required for using this Go package. The headers are included in the repository for build-time compilation.

### Installing FabricManager Development Package (For Development Only)

**On Ubuntu/Debian:**
```bash
sudo apt-get install nvidia-fabricmanager-dev-<version>
```

**On RHEL/CentOS:**
```bash
sudo yum install nvidia-fabricmanager-devel-<version>
```

**Note:** The development package is only needed if you're developing or building the Go package from source. For regular usage, only the runtime package is required.

## Branch Strategy

This repository uses a **branch-per-FM-version** strategy:

- **`main`**: Development branch (no specific headers)
- **`fm/<version>`**: Branch for specific FabricManager version (e.g., `fm/575.57.08`)
- **Tags**: Release tags on version branches (e.g., `v575.57.08-1.0`)

### Available Versions

- **`fm/575.57.08`**: FabricManager 575.57.08 support
- **`main`**: Latest development version

### How to Choose

1. **For production**: Use a specific version branch or tag
2. **For development**: Use `main` branch
3. **For compatibility**: Match your system's FabricManager version

## Usage

### As a Library

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/NVIDIA/go-fabricmanager"
)

func main() {
    // Initialize the FabricManager library
    if err := fabricmanager.Init(); err != nil {
        log.Fatalf("Failed to initialize FabricManager: %v", err)
    }
    defer fabricmanager.Shutdown()

    // Connect to FabricManager
    client, err := fabricmanager.Connect("127.0.0.1:6666", 5000)
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer client.Disconnect()

    // List partitions
    partitions, err := client.GetSupportedPartitions()
    if err != nil {
        log.Fatalf("Failed to get partitions: %v", err)
    }

    for _, partition := range partitions {
        fmt.Printf("Partition ID: %d, Active: %t, GPUs: %d\n", 
            partition.ID, partition.IsActive, partition.NumGPUs)
    }
}
```

### CLI Tool

The package includes a command-line interface similar to the C++ `fmpm` tool:

```bash
# List all partitions
./fmpm -l

# Activate a partition
./fmpm -a 1

# Deactivate a partition
./fmpm -d 1

# Connect to remote FabricManager
./fmpm --hostname 192.168.1.100 -l

# Use Unix domain socket
./fmpm --unix-domain-socket /tmp/fabricmanager.sock -l

# Get NVLink failed devices
./fmpm --get-nvlink-failed-devices

# List unsupported partitions
./fmpm --list-unsupported-partitions

# Set activated partition list (for resiliency mode)
./fmpm --set-activated-list 1,2,3

# Show version
./fmpm -v
```

## Building

```bash
# Build the CLI tool
go build -o fmpm cmd/fmpm/main.go

# Build with CGO (required)
CGO_ENABLED=1 go build -o fmpm cmd/fmpm/main.go

# Or use the build script
./build.sh

# Or use make
make build
```

## Versioning

This project uses a **branch-per-FM-version** strategy for managing multiple FabricManager versions in parallel.

- **Branch format:** `fm/<fabricmanager-version>` (e.g., `fm/575.57.08`)
- **Tag format:** `v<fabricmanager-version>-<X.Y>` (e.g., `v575.57.08-1.0`)
- **Headers:** Each branch contains headers for its specific FM version in the `headers/` directory.

**Release workflow:**
1. Create a new branch for the FM version: `git checkout -b fm/<version>`
2. Update headers and Go bindings for that version.
3. Test thoroughly.
4. Create a tag for the release: `git tag v<version>-1.0`
5. Push branch and tag to trigger GitHub release.

**Version management:**
- Each FM version has its own branch for parallel maintenance
- Tags provide immutable release points on each branch
- Main branch is for development (no specific headers)

**Example**

- `fm/575.57.08` branch with `v575.57.08-1.0` tag (initial release)
- `fm/575.57.08` branch with `v575.57.08-1.1` tag (bug fix)
- `fm/575.57.08` branch with `v575.57.08-2.0` tag (new features)
- `fm/580.0.0` branch with `v580.0.0-1.0` tag (initial release)
- `main` branch for development

## API Reference

### Core Functions

- `Init()` - Initialize the FabricManager library
- `Shutdown()` - Shutdown the FabricManager library
- `Connect(address string, timeoutMs int) (*Client, error)` - Connect to FabricManager
- `Client.Disconnect()` - Disconnect from FabricManager

### Client Methods

- `GetSupportedPartitions() ([]Partition, error)` - Get list of supported partitions
- `ActivatePartition(id uint32) error` - Activate a partition
- `DeactivatePartition(id uint32) error` - Deactivate a partition
- `GetNvlinkFailedDevices() (*NvlinkFailedDevices, error)` - Get NVLink failed devices
- `GetUnsupportedPartitions() ([]UnsupportedPartition, error)` - Get unsupported partitions
- `SetActivatedPartitions(ids []uint32) error` - Set activated partition list

## Error Handling

The package provides comprehensive error handling with specific error types:

```go
if err != nil {
    switch {
    case fabricmanager.IsConnectionError(err):
        // Handle connection issues
    case fabricmanager.IsResourceError(err):
        // Handle resource-related errors
    case fabricmanager.IsPartitionError(err):
        // Handle partition-related errors
    default:
        // Handle other errors
    }
}
```

## Configuration

The FabricManager connection can be configured via:

1. **Environment variables:**
   - `FM_HOST` - Default hostname/IP (default: 127.0.0.1)
   - `FM_PORT` - Default port (default: 6666)
   - `FM_TIMEOUT` - Default timeout in milliseconds (default: 5000)

2. **Configuration file:**
   - Default location: `/usr/share/nvidia/nvswitch/fabricmanager.cfg`
   - Can be overridden with `FM_CONFIG_FILE` environment variable

## Examples

See the `examples/` directory for complete working examples.

## License

- **Go bindings, CLI, and all original code:** Licensed under the [MIT License](LICENSE)
- **NVIDIA FabricManager components:** © NVIDIA Corporation, subject to their own license terms

### Important Notice

This project is **NOT affiliated with or endorsed by NVIDIA Corporation**. The Go bindings and CLI implementation are independent open-source work that provides Go language bindings for NVIDIA FabricManager.

For details about NVIDIA's components, see the [NOTICE](NOTICE) file.

## Coverage Analysis

This project includes automated coverage analysis to ensure complete API and CLI compatibility.

### API Coverage
- **C API Coverage:** Measures how many C API symbols from `libnvfm.so` are wrapped in the Go package
- **CLI Coverage:** Compares the original `fmpm` tool with our Go CLI implementation
- **Version-specific:** Downloads the exact FabricManager version for comparison

### Running Coverage Analysis

```bash
# Run full coverage analysis
./scripts/coverage-analysis.sh --json --update-readme

# Run coverage analysis with CLI comparison
./scripts/coverage-analysis.sh --json --update-readme --cli-version 575.57.08

# Show coverage summary
./scripts/coverage-analysis.sh --summary

# Compare CLI implementations (standalone)
./scripts/cli-comparison.sh --version 575.57.08

# Clean temporary files
./scripts/coverage-analysis.sh --clean
```

### Coverage Reports
- **Markdown reports:** Generated in `coverage/` directory
- **JSON reports:** For CI integration
- **README badges:** Automatically updated with coverage percentage

### CI Integration
The CI/CD pipeline runs automatically on:
- **Pull requests** - Full validation with coverage analysis
- **Pushes to branches** - Extended testing and coverage
- **Tag pushes** - Release creation with assets

### CI/CD Pipeline
The repository uses a single, comprehensive CI/CD pipeline with incremental complexity:

**Basic Checks (All Scenarios):**
- ✅ Package builds successfully
- ✅ Unit tests pass
- ✅ Code passes linting

**Extended Checks (PRs & Pushes):**
- ✅ Race condition detection
- ✅ Security vulnerability scanning
- ✅ API coverage analysis (100% required)
- ✅ CLI functionality testing
- ✅ Cross-compilation testing
- ✅ Coverage reports and badges

**Release Steps (Tag Pushes):**
- ✅ CLI comparison with original tool
- ✅ Release creation with assets
- ✅ Binary uploads for multiple architectures

### Quality Requirements
All changes must pass the quality gates defined in the CI/CD pipeline. The pipeline automatically adapts its complexity based on the trigger:

- **PRs & Pushes:** Full validation with coverage analysis
- **Releases:** Basic validation + release creation



## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run coverage analysis: `./scripts/coverage-analysis.sh`
6. Submit a pull request

## Support

For issues related to this Go package, please file an issue on GitHub.

For NVIDIA FabricManager support, refer to the [official documentation](https://docs.nvidia.com/datacenter/tesla/fabric-manager-user-guide/).