package fabricmanager

/*
#cgo CFLAGS: -I${SRCDIR}/headers
#cgo LDFLAGS: -lnvfm
#include "nv_fm_agent.h"
#include "nv_fm_types.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"
)

// Version information
const (
	Version = "575.57.08"
)

// Return codes from FabricManager API
const (
	FM_ST_SUCCESS                            = 0
	FM_ST_BADPARAM                           = -1
	FM_ST_GENERIC_ERROR                      = -2
	FM_ST_NOT_SUPPORTED                      = -3
	FM_ST_UNINITIALIZED                      = -4
	FM_ST_TIMEOUT                            = -5
	FM_ST_VERSION_MISMATCH                   = -6
	FM_ST_IN_USE                             = -7
	FM_ST_NOT_CONFIGURED                     = -8
	FM_ST_CONNECTION_NOT_VALID               = -9
	FM_ST_NVLINK_ERROR                       = -10
	FM_ST_RESOURCE_BAD                       = -11
	FM_ST_RESOURCE_IN_USE                    = -12
	FM_ST_RESOURCE_NOT_IN_USE                = -13
	FM_ST_RESOURCE_EXHAUSTED                 = -14
	FM_ST_RESOURCE_NOT_READY                 = -15
	FM_ST_PARTITION_EXISTS                   = -16
	FM_ST_PARTITION_ID_IN_USE                = -17
	FM_ST_PARTITION_ID_NOT_IN_USE            = -18
	FM_ST_PARTITION_NAME_IN_USE              = -19
	FM_ST_PARTITION_NAME_NOT_IN_USE          = -20
	FM_ST_PARTITION_ID_NAME_MISMATCH         = -21
	FM_ST_NOT_READY                          = -22
	FM_ST_RESOURCE_USED_IN_THIS_PARTITION    = -23
	FM_ST_RESOURCE_USED_IN_ANOTHER_PARTITION = -24
)

// Default values
const (
	FM_CMD_PORT_NUMBER       = 6666
	FM_MAX_STR_LENGTH        = 256
	FM_MAX_NUM_GPUS          = 16
	FM_MAX_FABRIC_PARTITIONS = 64
)

// Error types
type FMError struct {
	Code    int
	Message string
}

func (e *FMError) Error() string {
	return fmt.Sprintf("FabricManager error %d: %s", e.Code, e.Message)
}

// Error helper functions
func IsConnectionError(err error) bool {
	if fmErr, ok := err.(*FMError); ok {
		return fmErr.Code == FM_ST_CONNECTION_NOT_VALID ||
			fmErr.Code == FM_ST_UNINITIALIZED ||
			fmErr.Code == FM_ST_TIMEOUT
	}
	return false
}

func IsResourceError(err error) bool {
	if fmErr, ok := err.(*FMError); ok {
		return fmErr.Code == FM_ST_RESOURCE_BAD ||
			fmErr.Code == FM_ST_RESOURCE_IN_USE ||
			fmErr.Code == FM_ST_RESOURCE_NOT_IN_USE ||
			fmErr.Code == FM_ST_RESOURCE_EXHAUSTED ||
			fmErr.Code == FM_ST_RESOURCE_NOT_READY
	}
	return false
}

func IsPartitionError(err error) bool {
	if fmErr, ok := err.(*FMError); ok {
		return fmErr.Code == FM_ST_PARTITION_EXISTS ||
			fmErr.Code == FM_ST_PARTITION_ID_IN_USE ||
			fmErr.Code == FM_ST_PARTITION_ID_NOT_IN_USE ||
			fmErr.Code == FM_ST_PARTITION_NAME_IN_USE ||
			fmErr.Code == FM_ST_PARTITION_NAME_NOT_IN_USE ||
			fmErr.Code == FM_ST_PARTITION_ID_NAME_MISMATCH
	}
	return false
}

// PCI Device information
type PCIDevice struct {
	Domain   uint32
	Bus      uint32
	Device   uint32
	Function uint32
}

// GPU information within a partition
type PartitionGPUInfo struct {
	PhysicalID          uint32
	UUID                string
	PCIBusID            string
	NumNvLinksAvailable uint32
	MaxNumNvLinks       uint32
	NvlinkLineRateMBps  uint32
}

// Fabric partition information
type Partition struct {
	ID       uint32
	IsActive bool
	NumGPUs  uint32
	GPUs     []PartitionGPUInfo
}

// NVLink failed device information
type NvlinkFailedDeviceInfo struct {
	UUID     string
	PCIBusID string
	NumPorts uint32
	PortNums []uint32
}

// NVLink failed devices
type NvlinkFailedDevices struct {
	NumGPUs     uint32
	NumSwitches uint32
	GPUInfo     []NvlinkFailedDeviceInfo
	SwitchInfo  []NvlinkFailedDeviceInfo
}

// Unsupported partition information
type UnsupportedPartition struct {
	ID             uint32
	NumGPUs        uint32
	GPUPhysicalIDs []uint32
}

// Client represents a connection to FabricManager
type Client struct {
	handle C.fmHandle_t
}

// convertReturnCode converts C return code to Go error
func convertReturnCode(code C.fmReturn_t) error {
	if code == C.FM_ST_SUCCESS {
		return nil
	}

	// Convert C string to Go string for error message
	var message string
	switch code {
	case C.FM_ST_BADPARAM:
		message = "Bad parameter"
	case C.FM_ST_GENERIC_ERROR:
		message = "Generic error"
	case C.FM_ST_NOT_SUPPORTED:
		message = "Not supported"
	case C.FM_ST_UNINITIALIZED:
		message = "Uninitialized"
	case C.FM_ST_TIMEOUT:
		message = "Timeout"
	case C.FM_ST_VERSION_MISMATCH:
		message = "Version mismatch"
	case C.FM_ST_IN_USE:
		message = "Resource in use"
	case C.FM_ST_NOT_CONFIGURED:
		message = "Not configured"
	case C.FM_ST_CONNECTION_NOT_VALID:
		message = "Connection not valid"
	case C.FM_ST_NVLINK_ERROR:
		message = "NVLink error"
	case C.FM_ST_RESOURCE_BAD:
		message = "Bad resource"
	case C.FM_ST_RESOURCE_IN_USE:
		message = "Resource in use"
	case C.FM_ST_RESOURCE_NOT_IN_USE:
		message = "Resource not in use"
	case C.FM_ST_RESOURCE_EXHAUSTED:
		message = "Resource exhausted"
	case C.FM_ST_RESOURCE_NOT_READY:
		message = "Resource not ready"
	case C.FM_ST_PARTITION_EXISTS:
		message = "Partition exists"
	case C.FM_ST_PARTITION_ID_IN_USE:
		message = "Partition ID in use"
	case C.FM_ST_PARTITION_ID_NOT_IN_USE:
		message = "Partition ID not in use"
	case C.FM_ST_PARTITION_NAME_IN_USE:
		message = "Partition name in use"
	case C.FM_ST_PARTITION_NAME_NOT_IN_USE:
		message = "Partition name not in use"
	case C.FM_ST_PARTITION_ID_NAME_MISMATCH:
		message = "Partition ID name mismatch"
	case C.FM_ST_NOT_READY:
		message = "Not ready"
	case C.FM_ST_RESOURCE_USED_IN_THIS_PARTITION:
		message = "Resource used in this partition"
	case C.FM_ST_RESOURCE_USED_IN_ANOTHER_PARTITION:
		message = "Resource used in another partition"
	default:
		message = "Unknown error"
	}

	return &FMError{Code: int(code), Message: message}
}

// Init initializes the FabricManager library
func Init() error {
	ret := C.fmLibInit()
	return convertReturnCode(ret)
}

// Shutdown shuts down the FabricManager library
func Shutdown() error {
	ret := C.fmLibShutdown()
	return convertReturnCode(ret)
}

// Connect connects to a FabricManager instance
func Connect(address string, timeoutMs int) (*Client, error) {
	// Parse address to determine if it's a Unix socket or TCP
	isUnixSocket := strings.HasPrefix(address, "/") || strings.Contains(address, ".sock")

	// Create connection parameters
	params := C.fmConnectParams_t{
		version:             C.fmConnectParams_version,
		timeoutMs:           C.uint(timeoutMs),
		addressIsUnixSocket: C.uint(0),
	}

	if isUnixSocket {
		params.addressIsUnixSocket = C.uint(1)
	}

	// Copy address string to C buffer
	addrCStr := C.CString(address)
	defer C.free(unsafe.Pointer(addrCStr))
	C.strncpy(&params.addressInfo[0], addrCStr, C.FM_MAX_STR_LENGTH-1)
	params.addressInfo[C.FM_MAX_STR_LENGTH-1] = 0

	// Connect
	var handle C.fmHandle_t
	ret := C.fmConnect(&params, &handle)
	if ret != C.FM_ST_SUCCESS {
		return nil, convertReturnCode(ret)
	}

	return &Client{handle: handle}, nil
}

// Disconnect disconnects from the FabricManager instance
func (c *Client) Disconnect() error {
	ret := C.fmDisconnect(c.handle)
	return convertReturnCode(ret)
}

// GetSupportedPartitions gets the list of supported fabric partitions
func (c *Client) GetSupportedPartitions() ([]Partition, error) {
	var partitionList C.fmFabricPartitionList_t
	partitionList.version = C.fmFabricPartitionList_version

	ret := C.fmGetSupportedFabricPartitions(c.handle, &partitionList)
	if ret != C.FM_ST_SUCCESS {
		return nil, convertReturnCode(ret)
	}

	partitions := make([]Partition, partitionList.numPartitions)
	for i := 0; i < int(partitionList.numPartitions); i++ {
		cPartition := partitionList.partitionInfo[i]

		partition := Partition{
			ID:       uint32(cPartition.partitionId),
			IsActive: cPartition.isActive != 0,
			NumGPUs:  uint32(cPartition.numGpus),
			GPUs:     make([]PartitionGPUInfo, cPartition.numGpus),
		}

		for j := 0; j < int(cPartition.numGpus); j++ {
			cGPU := cPartition.gpuInfo[j]
			partition.GPUs[j] = PartitionGPUInfo{
				PhysicalID:          uint32(cGPU.physicalId),
				UUID:                C.GoString(&cGPU.uuid[0]),
				PCIBusID:            C.GoString(&cGPU.pciBusId[0]),
				NumNvLinksAvailable: uint32(cGPU.numNvLinksAvailable),
				MaxNumNvLinks:       uint32(cGPU.maxNumNvLinks),
				NvlinkLineRateMBps:  uint32(cGPU.nvlinkLineRateMBps),
			}
		}

		partitions[i] = partition
	}

	return partitions, nil
}

// ActivatePartition activates a fabric partition
func (c *Client) ActivatePartition(id uint32) error {
	ret := C.fmActivateFabricPartition(c.handle, C.fmFabricPartitionId_t(id))
	return convertReturnCode(ret)
}

// DeactivatePartition deactivates a fabric partition
func (c *Client) DeactivatePartition(id uint32) error {
	ret := C.fmDeactivateFabricPartition(c.handle, C.fmFabricPartitionId_t(id))
	return convertReturnCode(ret)
}

// GetNvlinkFailedDevices gets information about NVLink failed devices
func (c *Client) GetNvlinkFailedDevices() (*NvlinkFailedDevices, error) {
	var failedDevices C.fmNvlinkFailedDevices_v1
	failedDevices.version = C.fmNvlinkFailedDevices_version

	ret := C.fmGetNvlinkFailedDevices(c.handle, (*C.fmNvlinkFailedDevices_t)(unsafe.Pointer(&failedDevices)))
	if ret != C.FM_ST_SUCCESS {
		return nil, convertReturnCode(ret)
	}

	result := &NvlinkFailedDevices{
		NumGPUs:     uint32(failedDevices.numGpus),
		NumSwitches: uint32(failedDevices.numSwitches),
		GPUInfo:     make([]NvlinkFailedDeviceInfo, failedDevices.numGpus),
		SwitchInfo:  make([]NvlinkFailedDeviceInfo, failedDevices.numSwitches),
	}

	// Convert GPU info
	for i := 0; i < int(failedDevices.numGpus); i++ {
		cGPU := failedDevices.gpuInfo[i]
		result.GPUInfo[i] = NvlinkFailedDeviceInfo{
			UUID:     C.GoString(&cGPU.uuid[0]),
			PCIBusID: C.GoString(&cGPU.pciBusId[0]),
			NumPorts: uint32(cGPU.numPorts),
			PortNums: make([]uint32, cGPU.numPorts),
		}
		for j := 0; j < int(cGPU.numPorts); j++ {
			result.GPUInfo[i].PortNums[j] = uint32(cGPU.portNum[j])
		}
	}

	// Convert switch info
	for i := 0; i < int(failedDevices.numSwitches); i++ {
		cSwitch := failedDevices.switchInfo[i]
		result.SwitchInfo[i] = NvlinkFailedDeviceInfo{
			UUID:     C.GoString(&cSwitch.uuid[0]),
			PCIBusID: C.GoString(&cSwitch.pciBusId[0]),
			NumPorts: uint32(cSwitch.numPorts),
			PortNums: make([]uint32, cSwitch.numPorts),
		}
		for j := 0; j < int(cSwitch.numPorts); j++ {
			result.SwitchInfo[i].PortNums[j] = uint32(cSwitch.portNum[j])
		}
	}

	return result, nil
}

// GetUnsupportedPartitions gets the list of unsupported fabric partitions
func (c *Client) GetUnsupportedPartitions() ([]UnsupportedPartition, error) {
	var unsupportedList C.fmUnsupportedFabricPartitionList_v1
	unsupportedList.version = C.fmUnsupportedFabricPartitionList_version

	ret := C.fmGetUnsupportedFabricPartitions(c.handle, (*C.fmUnsupportedFabricPartitionList_t)(unsafe.Pointer(&unsupportedList)))
	if ret != C.FM_ST_SUCCESS {
		return nil, convertReturnCode(ret)
	}

	partitions := make([]UnsupportedPartition, unsupportedList.numPartitions)
	for i := 0; i < int(unsupportedList.numPartitions); i++ {
		cPartition := unsupportedList.partitionInfo[i]

		partition := UnsupportedPartition{
			ID:             uint32(cPartition.partitionId),
			NumGPUs:        uint32(cPartition.numGpus),
			GPUPhysicalIDs: make([]uint32, cPartition.numGpus),
		}

		for j := 0; j < int(cPartition.numGpus); j++ {
			partition.GPUPhysicalIDs[j] = uint32(cPartition.gpuPhysicalIds[j])
		}

		partitions[i] = partition
	}

	return partitions, nil
}

// SetActivatedPartitions sets the list of currently activated fabric partitions
func (c *Client) SetActivatedPartitions(ids []uint32) error {
	var activatedList C.fmActivatedFabricPartitionList_v1
	activatedList.version = C.fmActivatedFabricPartitionList_version
	activatedList.numPartitions = C.uint(len(ids))

	for i, id := range ids {
		activatedList.partitionIds[i] = C.fmFabricPartitionId_t(id)
	}

	ret := C.fmSetActivatedFabricPartitions(c.handle, (*C.fmActivatedFabricPartitionList_t)(unsafe.Pointer(&activatedList)))
	return convertReturnCode(ret)
}
