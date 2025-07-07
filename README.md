# NVIDIA FabricManager Go Package

A Go package and CLI tool for managing GPU partitions for NVIDIA Fabric Manager's Shared NVSwitch feature.

## Overview

This package provides a Go interface to NVIDIA FabricManager, allowing you to:
- List available fabric partitions
- Activate and deactivate partitions
- Query NVLink failed devices
- List unsupported partitions
- Set activated partition lists for resiliency mode

## Installation

```bash
go get github.com/NVIDIA/go-fabricmanager
```

## Prerequisites

- NVIDIA FabricManager runtime package installed (for runtime library)
- Go 1.21 or later
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

This project tracks NVIDIA FabricManager versions. Each Git tag is a complete, tested release of the Go bindings and headers for a specific FabricManager version.

- **Tag format:** `v<fabricmanager-version>-<go-change-number>` (e.g., `v575.57.08-1`)
- **Headers:** Only the current version's headers are present in the `headers/` directory.
- **Go bindings:** Always match the headers in the current repository state.

**Release workflow:**
1. Update `headers/` to the new FabricManager version.
2. Update Go bindings if needed.
3. Test thoroughly.
4. Create a tag for the new release (e.g., `v575.57.08-1`).
5. Push the tag to trigger a GitHub release.

**Historical versions:**
- Only the latest version is present in the main branch.
- Previous versions are available via Git tags.

**Example**

- `v575.57.08-1`: First release for FabricManager 575.57.08
- `v575.57.08-2`: Go-specific bug fix for 575.57.08
- `v580.0.0-1`: First release for FabricManager 580.0.0

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

This project is licensed under the same terms as the NVIDIA FabricManager library.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

For issues related to this Go package, please file an issue on GitHub.

For NVIDIA FabricManager support, refer to the [official documentation](https://docs.nvidia.com/datacenter/tesla/fabric-manager-user-guide/).