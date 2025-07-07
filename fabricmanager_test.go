package fabricmanager

import (
	"testing"
)

func TestErrorTypes(t *testing.T) {
	// Test connection errors
	connErr := &FMError{Code: FM_ST_CONNECTION_NOT_VALID, Message: "test"}
	if !IsConnectionError(connErr) {
		t.Error("Expected connection error to be detected")
	}

	// Test resource errors
	resourceErr := &FMError{Code: FM_ST_RESOURCE_BAD, Message: "test"}
	if !IsResourceError(resourceErr) {
		t.Error("Expected resource error to be detected")
	}

	// Test partition errors
	partitionErr := &FMError{Code: FM_ST_PARTITION_ID_NOT_IN_USE, Message: "test"}
	if !IsPartitionError(partitionErr) {
		t.Error("Expected partition error to be detected")
	}

	// Test non-FM errors
	genericErr := &FMError{Code: FM_ST_GENERIC_ERROR, Message: "test"}
	if IsConnectionError(genericErr) || IsResourceError(genericErr) || IsPartitionError(genericErr) {
		t.Error("Generic error should not be classified as specific error type")
	}
}

func TestErrorString(t *testing.T) {
	err := &FMError{Code: FM_ST_BADPARAM, Message: "Bad parameter"}
	expected := "FabricManager error -1: Bad parameter"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	if FM_ST_SUCCESS != 0 {
		t.Errorf("Expected FM_ST_SUCCESS to be 0, got %d", FM_ST_SUCCESS)
	}

	if FM_ST_BADPARAM != -1 {
		t.Errorf("Expected FM_ST_BADPARAM to be -1, got %d", FM_ST_BADPARAM)
	}

	if FM_CMD_PORT_NUMBER != 6666 {
		t.Errorf("Expected FM_CMD_PORT_NUMBER to be 6666, got %d", FM_CMD_PORT_NUMBER)
	}

	// Test new constants we added for 100% coverage
	if FM_MAX_STR_LENGTH != 256 {
		t.Errorf("Expected FM_MAX_STR_LENGTH to be 256, got %d", FM_MAX_STR_LENGTH)
	}

	if FM_MAX_NUM_GPUS != 16 {
		t.Errorf("Expected FM_MAX_NUM_GPUS to be 16, got %d", FM_MAX_NUM_GPUS)
	}

	if FM_MAX_FABRIC_PARTITIONS != 64 {
		t.Errorf("Expected FM_MAX_FABRIC_PARTITIONS to be 64, got %d", FM_MAX_FABRIC_PARTITIONS)
	}

	if FM_MAX_NUM_NVLINK_PORTS != 64 {
		t.Errorf("Expected FM_MAX_NUM_NVLINK_PORTS to be 64, got %d", FM_MAX_NUM_NVLINK_PORTS)
	}

	if FM_MAX_NUM_NVSWITCHES != 12 {
		t.Errorf("Expected FM_MAX_NUM_NVSWITCHES to be 12, got %d", FM_MAX_NUM_NVSWITCHES)
	}

	if FM_DEVICE_PCI_BUS_ID_BUFFER_SIZE != 32 {
		t.Errorf("Expected FM_DEVICE_PCI_BUS_ID_BUFFER_SIZE to be 32, got %d", FM_DEVICE_PCI_BUS_ID_BUFFER_SIZE)
	}

	if FM_UUID_BUFFER_SIZE != 80 {
		t.Errorf("Expected FM_UUID_BUFFER_SIZE to be 80, got %d", FM_UUID_BUFFER_SIZE)
	}

	// Test version constants
	if FM_ACTIVATED_FABRIC_PARTITION_LIST_VERSION1 != 1 {
		t.Errorf("Expected FM_ACTIVATED_FABRIC_PARTITION_LIST_VERSION1 to be 1, got %d", FM_ACTIVATED_FABRIC_PARTITION_LIST_VERSION1)
	}

	if FM_ACTIVATED_FABRIC_PARTITION_LIST_VERSION != FM_ACTIVATED_FABRIC_PARTITION_LIST_VERSION1 {
		t.Errorf("Expected FM_ACTIVATED_FABRIC_PARTITION_LIST_VERSION to equal FM_ACTIVATED_FABRIC_PARTITION_LIST_VERSION1")
	}

	if FM_CONNECT_PARAMS_VERSION1 != 1 {
		t.Errorf("Expected FM_CONNECT_PARAMS_VERSION1 to be 1, got %d", FM_CONNECT_PARAMS_VERSION1)
	}

	if FM_FABRIC_PARTITION_LIST_VERSION2 != 1 {
		t.Errorf("Expected FM_FABRIC_PARTITION_LIST_VERSION2 to be 1, got %d", FM_FABRIC_PARTITION_LIST_VERSION2)
	}

	if FM_NVLINK_FAILED_DEVICES_VERSION1 != 1 {
		t.Errorf("Expected FM_NVLINK_FAILED_DEVICES_VERSION1 to be 1, got %d", FM_NVLINK_FAILED_DEVICES_VERSION1)
	}

	if FM_UNSUPPORTED_FABRIC_PARTITION_LIST_VERSION1 != 1 {
		t.Errorf("Expected FM_UNSUPPORTED_FABRIC_PARTITION_LIST_VERSION1 to be 1, got %d", FM_UNSUPPORTED_FABRIC_PARTITION_LIST_VERSION1)
	}

	// Test header guard constants
	if NV_FM_AGENT_H != 1 {
		t.Errorf("Expected NV_FM_AGENT_H to be 1, got %d", NV_FM_AGENT_H)
	}

	if NV_FM_TYPES_H != 1 {
		t.Errorf("Expected NV_FM_TYPES_H to be 1, got %d", NV_FM_TYPES_H)
	}
}

func TestNewTypes(t *testing.T) {
	// Test that new types can be created
	var partitionID FabricPartitionID = 123
	if partitionID != 123 {
		t.Errorf("Expected FabricPartitionID to work as uint32, got %d", partitionID)
	}

	// Test that placeholder types can be instantiated
	activatedList := ActivatedFabricPartitionList{}
	_ = activatedList // Just verify it can be instantiated

	connectParams := ConnectParams{}
	_ = connectParams // Just verify it can be instantiated

	fabricList := FabricPartitionList{}
	_ = fabricList // Just verify it can be instantiated

	nvlinkList := NvlinkFailedDevicesList{}
	_ = nvlinkList // Just verify it can be instantiated

	unsupportedList := UnsupportedFabricPartitionList{}
	_ = unsupportedList // Just verify it can be instantiated
}

func TestMakeFMParamVersion(t *testing.T) {
	// Test the helper function for creating version numbers
	result := makeFMParamVersion(16, 1)
	expected := uint32(16) | (1 << 24)
	if result != expected {
		t.Errorf("Expected makeFMParamVersion(16, 1) to be %d, got %d", expected, result)
	}

	// Test with different values
	result2 := makeFMParamVersion(32, 2)
	expected2 := uint32(32) | (2 << 24)
	if result2 != expected2 {
		t.Errorf("Expected makeFMParamVersion(32, 2) to be %d, got %d", expected2, result2)
	}
}

// Note: convertReturnCode tests require CGO and are tested in integration tests
func TestConvertReturnCode(t *testing.T) {
	t.Skip("Skipping CGO-dependent test")
}

// Note: These tests require a running FabricManager instance
// They are commented out to avoid failures in CI/CD environments
/*
func TestInitShutdown(t *testing.T) {
	// Test initialization
	if err := Init(); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test shutdown
	if err := Shutdown(); err != nil {
		t.Fatalf("Failed to shutdown: %v", err)
	}
}

func TestConnect(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer Shutdown()

	// Test connection to local FabricManager
	client, err := Connect("127.0.0.1:6666", 5000)
	if err != nil {
		t.Skipf("Skipping test - no local FabricManager running: %v", err)
	}
	defer client.Disconnect()

	if client == nil {
		t.Error("Expected client to be non-nil")
	}
}
*/
