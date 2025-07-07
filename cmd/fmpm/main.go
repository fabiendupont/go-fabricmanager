package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/NVIDIA/go-fabricmanager"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	hostname         string
	unixDomainSocket string
	timeoutMs        int = 5000

	// Root command
	rootCmd = &cobra.Command{
		Use:   "fmpm",
		Short: "NVIDIA FabricManager Partition Manager",
		Long: `FM Partition Manager (fmpm) is a tool for managing GPU partitions 
for NVIDIA Fabric Manager's Shared NVSwitch feature.

Management operations include listing, activating, deactivating partitions, etc.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize FabricManager library
			if err := fabricmanager.Init(); err != nil {
				return fmt.Errorf("failed to initialize FabricManager: %v", err)
			}
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			// Shutdown FabricManager library
			if err := fabricmanager.Shutdown(); err != nil {
				log.Printf("Warning: failed to shutdown FabricManager: %v", err)
			}
		},
	}

	// List command
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all supported fabric partitions",
		Long:  "List all supported fabric partitions with their current status and GPU information",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := connectToFabricManager()
			if err != nil {
				return err
			}
			defer client.Disconnect()

			partitions, err := client.GetSupportedPartitions()
			if err != nil {
				return fmt.Errorf("failed to get partitions: %v", err)
			}

			if len(partitions) == 0 {
				fmt.Println("No partitions found")
				return nil
			}

			fmt.Printf("Found %d partition(s):\n\n", len(partitions))
			for _, partition := range partitions {
				status := "Inactive"
				if partition.IsActive {
					status = "Active"
				}
				fmt.Printf("Partition ID: %d\n", partition.ID)
				fmt.Printf("  Status: %s\n", status)
				fmt.Printf("  GPUs: %d\n", partition.NumGPUs)

				if len(partition.GPUs) > 0 {
					fmt.Printf("  GPU Details:\n")
					for _, gpu := range partition.GPUs {
						fmt.Printf("    Physical ID: %d\n", gpu.PhysicalID)
						fmt.Printf("    UUID: %s\n", gpu.UUID)
						fmt.Printf("    PCI Bus ID: %s\n", gpu.PCIBusID)
						fmt.Printf("    NVLinks Available: %d/%d\n", gpu.NumNvLinksAvailable, gpu.MaxNumNvLinks)
						fmt.Printf("    Line Rate: %d MB/s\n", gpu.NvlinkLineRateMBps)
						fmt.Println()
					}
				}
				fmt.Println()
			}

			return nil
		},
	}

	// Activate command
	activateCmd = &cobra.Command{
		Use:   "activate [partition-id]",
		Short: "Activate a fabric partition",
		Long:  "Activate a fabric partition by its ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			partitionID, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid partition ID: %v", err)
			}

			client, err := connectToFabricManager()
			if err != nil {
				return err
			}
			defer client.Disconnect()

			if err := client.ActivatePartition(uint32(partitionID)); err != nil {
				return fmt.Errorf("failed to activate partition %d: %v", partitionID, err)
			}

			fmt.Printf("Successfully activated partition %d\n", partitionID)
			return nil
		},
	}

	// Deactivate command
	deactivateCmd = &cobra.Command{
		Use:   "deactivate [partition-id]",
		Short: "Deactivate a fabric partition",
		Long:  "Deactivate a fabric partition by its ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			partitionID, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid partition ID: %v", err)
			}

			client, err := connectToFabricManager()
			if err != nil {
				return err
			}
			defer client.Disconnect()

			if err := client.DeactivatePartition(uint32(partitionID)); err != nil {
				return fmt.Errorf("failed to deactivate partition %d: %v", partitionID, err)
			}

			fmt.Printf("Successfully deactivated partition %d\n", partitionID)
			return nil
		},
	}

	// NVLink failed devices command
	nvlinkFailedCmd = &cobra.Command{
		Use:   "nvlink-failed",
		Short: "Query all NVLink failed devices",
		Long:  "Query all GPUs and NVSwitches with failed NVLinks",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := connectToFabricManager()
			if err != nil {
				return err
			}
			defer client.Disconnect()

			failedDevices, err := client.GetNvlinkFailedDevices()
			if err != nil {
				return fmt.Errorf("failed to get NVLink failed devices: %v", err)
			}

			fmt.Printf("NVLink Failed Devices Report:\n\n")
			fmt.Printf("GPUs with failed NVLinks: %d\n", failedDevices.NumGPUs)
			fmt.Printf("NVSwitches with failed NVLinks: %d\n\n", failedDevices.NumSwitches)

			if len(failedDevices.GPUInfo) > 0 {
				fmt.Println("Failed GPUs:")
				for i, gpu := range failedDevices.GPUInfo {
					fmt.Printf("  %d. UUID: %s\n", i+1, gpu.UUID)
					fmt.Printf("     PCI Bus ID: %s\n", gpu.PCIBusID)
					fmt.Printf("     Failed Ports: %d\n", gpu.NumPorts)
					if len(gpu.PortNums) > 0 {
						fmt.Printf("     Port Numbers: %v\n", gpu.PortNums)
					}
					fmt.Println()
				}
			}

			if len(failedDevices.SwitchInfo) > 0 {
				fmt.Println("Failed NVSwitches:")
				for i, switch_ := range failedDevices.SwitchInfo {
					fmt.Printf("  %d. UUID: %s\n", i+1, switch_.UUID)
					fmt.Printf("     PCI Bus ID: %s\n", switch_.PCIBusID)
					fmt.Printf("     Failed Ports: %d\n", switch_.NumPorts)
					if len(switch_.PortNums) > 0 {
						fmt.Printf("     Port Numbers: %v\n", switch_.PortNums)
					}
					fmt.Println()
				}
			}

			if len(failedDevices.GPUInfo) == 0 && len(failedDevices.SwitchInfo) == 0 {
				fmt.Println("No NVLink failures detected")
			}

			return nil
		},
	}

	// Unsupported partitions command
	unsupportedCmd = &cobra.Command{
		Use:   "unsupported",
		Short: "List unsupported fabric partitions",
		Long:  "Query all unsupported fabric partitions",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := connectToFabricManager()
			if err != nil {
				return err
			}
			defer client.Disconnect()

			partitions, err := client.GetUnsupportedPartitions()
			if err != nil {
				return fmt.Errorf("failed to get unsupported partitions: %v", err)
			}

			if len(partitions) == 0 {
				fmt.Println("No unsupported partitions found")
				return nil
			}

			fmt.Printf("Found %d unsupported partition(s):\n\n", len(partitions))
			for _, partition := range partitions {
				fmt.Printf("Partition ID: %d\n", partition.ID)
				fmt.Printf("  GPUs: %d\n", partition.NumGPUs)
				if len(partition.GPUPhysicalIDs) > 0 {
					fmt.Printf("  GPU Physical IDs: %v\n", partition.GPUPhysicalIDs)
				}
				fmt.Println()
			}

			return nil
		},
	}

	// Set activated partitions command
	setActivatedCmd = &cobra.Command{
		Use:   "set-activated [partition-ids]",
		Short: "Set activated partition list",
		Long:  "Set a list of currently activated fabric partitions (comma-separated, no spaces)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse comma-separated partition IDs
			idStrs := strings.Split(args[0], ",")
			partitionIDs := make([]uint32, 0, len(idStrs))

			for _, idStr := range idStrs {
				idStr = strings.TrimSpace(idStr)
				if idStr == "" {
					continue
				}
				id, err := strconv.ParseUint(idStr, 10, 32)
				if err != nil {
					return fmt.Errorf("invalid partition ID '%s': %v", idStr, err)
				}
				partitionIDs = append(partitionIDs, uint32(id))
			}

			client, err := connectToFabricManager()
			if err != nil {
				return err
			}
			defer client.Disconnect()

			if err := client.SetActivatedPartitions(partitionIDs); err != nil {
				return fmt.Errorf("failed to set activated partitions: %v", err)
			}

			fmt.Printf("Successfully set activated partitions: %v\n", partitionIDs)
			return nil
		},
	}

	// Version command
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("fmpm version %s\n", fabricmanager.Version)
		},
	}
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&hostname, "hostname", "127.0.0.1", "hostname or IP address (TCP socket) of Fabric Manager")
	rootCmd.PersistentFlags().StringVar(&unixDomainSocket, "unix-domain-socket", "", "UNIX domain socket path for Fabric Manager connection")
	rootCmd.PersistentFlags().IntVar(&timeoutMs, "timeout", 5000, "connection timeout in milliseconds")

	// Add commands
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(activateCmd)
	rootCmd.AddCommand(deactivateCmd)
	rootCmd.AddCommand(nvlinkFailedCmd)
	rootCmd.AddCommand(unsupportedCmd)
	rootCmd.AddCommand(setActivatedCmd)
	rootCmd.AddCommand(versionCmd)

	// Add legacy short flags for backward compatibility
	rootCmd.Flags().BoolP("list", "l", false, "List partitions (legacy flag)")
	rootCmd.Flags().UintP("activate", "a", 0, "Activate partition ID (legacy flag)")
	rootCmd.Flags().UintP("deactivate", "d", 0, "Deactivate partition ID (legacy flag)")
	rootCmd.Flags().Bool("get-nvlink-failed-devices", false, "Query all NVLink failed devices (legacy flag)")
	rootCmd.Flags().Bool("list-unsupported-partitions", false, "Query all unsupported fabric partitions (legacy flag)")
	rootCmd.Flags().String("set-activated-list", "", "Set activated partition list (legacy flag)")
	rootCmd.Flags().BoolP("version", "v", false, "Show version (legacy flag)")

	// Handle legacy flags
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Check for legacy flags
		if list, _ := cmd.Flags().GetBool("list"); list {
			return listCmd.RunE(cmd, args)
		}
		if activateID, _ := cmd.Flags().GetUint("activate"); activateID > 0 {
			return activateCmd.RunE(cmd, []string{strconv.FormatUint(uint64(activateID), 10)})
		}
		if deactivateID, _ := cmd.Flags().GetUint("deactivate"); deactivateID > 0 {
			return deactivateCmd.RunE(cmd, []string{strconv.FormatUint(uint64(deactivateID), 10)})
		}
		if nvlinkFailed, _ := cmd.Flags().GetBool("get-nvlink-failed-devices"); nvlinkFailed {
			return nvlinkFailedCmd.RunE(cmd, args)
		}
		if unsupported, _ := cmd.Flags().GetBool("list-unsupported-partitions"); unsupported {
			return unsupportedCmd.RunE(cmd, args)
		}
		if setActivated, _ := cmd.Flags().GetString("set-activated-list"); setActivated != "" {
			return setActivatedCmd.RunE(cmd, []string{setActivated})
		}
		if version, _ := cmd.Flags().GetBool("version"); version {
			versionCmd.Run(cmd, args)
			return nil
		}

		// If no legacy flags, show help
		return cmd.Help()
	}
}

func connectToFabricManager() (*fabricmanager.Client, error) {
	var address string

	if unixDomainSocket != "" {
		address = unixDomainSocket
	} else {
		// Check if hostname includes port
		if !strings.Contains(hostname, ":") {
			address = fmt.Sprintf("%s:%d", hostname, fabricmanager.FM_CMD_PORT_NUMBER)
		} else {
			address = hostname
		}
	}

	client, err := fabricmanager.Connect(address, timeoutMs)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to FabricManager at %s: %v", address, err)
	}

	return client, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
