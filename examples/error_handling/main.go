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

	// Example 1: Connection error handling
	fmt.Println("=== Connection Error Handling ===")

	// Try to connect to a non-existent FabricManager
	client, err := fabricmanager.Connect("192.168.1.999:6666", 1000)
	if err != nil {
		if fabricmanager.IsConnectionError(err) {
			fmt.Printf("Connection error (expected): %v\n", err)
		} else {
			fmt.Printf("Unexpected error: %v\n", err)
		}
	} else {
		client.Disconnect()
	}

	// Example 2: Resource error handling
	fmt.Println("\n=== Resource Error Handling ===")

	// Connect to local FabricManager
	client, err = fabricmanager.Connect("127.0.0.1:6666", 5000)
	if err != nil {
		log.Printf("Failed to connect to local FabricManager: %v", err)
		return
	}
	defer client.Disconnect()

	// Try to activate a non-existent partition
	err = client.ActivatePartition(99999)
	if err != nil {
		if fabricmanager.IsResourceError(err) {
			fmt.Printf("Resource error (expected): %v\n", err)
		} else if fabricmanager.IsPartitionError(err) {
			fmt.Printf("Partition error (expected): %v\n", err)
		} else {
			fmt.Printf("Unexpected error: %v\n", err)
		}
	}

	// Example 3: Partition error handling
	fmt.Println("\n=== Partition Error Handling ===")

	// Get current partitions
	partitions, err := client.GetSupportedPartitions()
	if err != nil {
		log.Printf("Failed to get partitions: %v", err)
		return
	}

	// Try to activate an already active partition
	for _, partition := range partitions {
		if partition.IsActive {
			fmt.Printf("Trying to activate already active partition %d...\n", partition.ID)
			err = client.ActivatePartition(partition.ID)
			if err != nil {
				if fabricmanager.IsPartitionError(err) {
					fmt.Printf("Partition error (expected): %v\n", err)
				} else {
					fmt.Printf("Unexpected error: %v\n", err)
				}
			}
			break
		}
	}

	// Example 4: Error type checking
	fmt.Println("\n=== Error Type Checking ===")

	// Demonstrate different error types
	testErrors := []error{
		&fabricmanager.FMError{Code: fabricmanager.FM_ST_CONNECTION_NOT_VALID, Message: "Connection test"},
		&fabricmanager.FMError{Code: fabricmanager.FM_ST_RESOURCE_BAD, Message: "Resource test"},
		&fabricmanager.FMError{Code: fabricmanager.FM_ST_PARTITION_ID_NOT_IN_USE, Message: "Partition test"},
		fmt.Errorf("Generic Go error"),
	}

	for i, testErr := range testErrors {
		fmt.Printf("Error %d: %v\n", i+1, testErr)
		fmt.Printf("  IsConnectionError: %t\n", fabricmanager.IsConnectionError(testErr))
		fmt.Printf("  IsResourceError: %t\n", fabricmanager.IsResourceError(testErr))
		fmt.Printf("  IsPartitionError: %t\n", fabricmanager.IsPartitionError(testErr))
		fmt.Println()
	}

	// Example 5: Graceful error recovery
	fmt.Println("=== Error Recovery ===")

	// Try to get partitions with retry logic
	var partitions2 []fabricmanager.Partition
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		partitions2, err = client.GetSupportedPartitions()
		if err == nil {
			fmt.Printf("Successfully retrieved %d partitions on attempt %d\n", len(partitions2), attempt)
			break
		}

		if fabricmanager.IsConnectionError(err) {
			fmt.Printf("Connection error on attempt %d: %v\n", attempt, err)
			if attempt < maxRetries {
				fmt.Println("Retrying...")
				// In a real application, you might want to wait here
				continue
			}
		}

		fmt.Printf("Failed to get partitions after %d attempts: %v\n", attempt, err)
		break
	}
}
