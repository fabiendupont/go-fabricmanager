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

	// Connect to FabricManager (default: localhost:6666)
	client, err := fabricmanager.Connect("127.0.0.1:6666", 5000)
	if err != nil {
		log.Fatalf("Failed to connect to FabricManager: %v", err)
	}
	defer client.Disconnect()

	fmt.Println("Connected to FabricManager successfully!")

	// List all supported partitions
	partitions, err := client.GetSupportedPartitions()
	if err != nil {
		log.Fatalf("Failed to get partitions: %v", err)
	}

	fmt.Printf("\nFound %d partition(s):\n", len(partitions))
	for _, partition := range partitions {
		status := "Inactive"
		if partition.IsActive {
			status = "Active"
		}
		fmt.Printf("  Partition %d: %s (%d GPUs)\n", partition.ID, status, partition.NumGPUs)
	}

	// Example: Activate the first inactive partition
	for _, partition := range partitions {
		if !partition.IsActive {
			fmt.Printf("\nActivating partition %d...\n", partition.ID)
			if err := client.ActivatePartition(partition.ID); err != nil {
				log.Printf("Failed to activate partition %d: %v", partition.ID, err)
				continue
			}
			fmt.Printf("Successfully activated partition %d\n", partition.ID)
			break
		}
	}

	// Get NVLink failed devices information
	failedDevices, err := client.GetNvlinkFailedDevices()
	if err != nil {
		log.Printf("Failed to get NVLink failed devices: %v", err)
	} else {
		fmt.Printf("\nNVLink Status:\n")
		fmt.Printf("  GPUs with failed NVLinks: %d\n", failedDevices.NumGPUs)
		fmt.Printf("  NVSwitches with failed NVLinks: %d\n", failedDevices.NumSwitches)
	}

	// Get unsupported partitions
	unsupported, err := client.GetUnsupportedPartitions()
	if err != nil {
		log.Printf("Failed to get unsupported partitions: %v", err)
	} else {
		fmt.Printf("  Unsupported partitions: %d\n", len(unsupported))
	}
}
