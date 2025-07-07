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
