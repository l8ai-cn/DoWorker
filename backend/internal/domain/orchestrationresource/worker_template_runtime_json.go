package orchestrationresource

import (
	"bytes"
	"encoding/json"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

type workerTemplateRuntimeJSON struct {
	RuntimeImageID     int64                             `json:"runtimeImageId"`
	PlacementPolicy    workerspec.PlacementPolicy        `json:"placementPolicy"`
	ComputeTargetRef   Reference                         `json:"computeTargetRef"`
	DeploymentMode     workerspec.DeploymentMode         `json:"deploymentMode"`
	ResourceProfileRef *Reference                        `json:"resourceProfileRef,omitempty"`
	CustomResources    *workerTemplateResourceLimitsJSON `json:"customResources,omitempty"`
}

type workerTemplateResourceLimitsJSON struct {
	CPURequestMilliCPU  uint32  `json:"cpuRequestMilliCPU"`
	CPULimitMilliCPU    uint32  `json:"cpuLimitMilliCPU"`
	MemoryRequestBytes  uint64  `json:"memoryRequestBytes"`
	MemoryLimitBytes    uint64  `json:"memoryLimitBytes"`
	StorageRequestBytes uint64  `json:"storageRequestBytes,omitempty"`
	StorageLimitBytes   uint64  `json:"storageLimitBytes,omitempty"`
	GPURequest          *uint32 `json:"gpuRequest,omitempty"`
	GPULimit            *uint32 `json:"gpuLimit,omitempty"`
}

func (spec WorkerTemplateRuntimeSpec) MarshalJSON() ([]byte, error) {
	wire := workerTemplateRuntimeJSON{
		RuntimeImageID:     spec.RuntimeImageID,
		PlacementPolicy:    spec.PlacementPolicy,
		ComputeTargetRef:   spec.ComputeTargetRef,
		DeploymentMode:     spec.DeploymentMode,
		ResourceProfileRef: spec.ResourceProfileRef,
	}
	if spec.CustomResources != nil {
		wire.CustomResources = resourceLimitsToJSON(*spec.CustomResources)
	}
	return json.Marshal(wire)
}

func (spec *WorkerTemplateRuntimeSpec) UnmarshalJSON(source []byte) error {
	if err := rejectWorkerTemplateRuntimeNulls(source); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.DisallowUnknownFields()
	var wire workerTemplateRuntimeJSON
	if err := decoder.Decode(&wire); err != nil {
		return err
	}
	if err := requireJSONEOF(decoder); err != nil {
		return err
	}

	spec.RuntimeImageID = wire.RuntimeImageID
	spec.PlacementPolicy = wire.PlacementPolicy
	spec.ComputeTargetRef = wire.ComputeTargetRef
	spec.DeploymentMode = wire.DeploymentMode
	spec.ResourceProfileRef = wire.ResourceProfileRef
	spec.CustomResources = nil
	if wire.CustomResources != nil {
		spec.CustomResources = resourceLimitsFromJSON(*wire.CustomResources)
	}
	return nil
}

func resourceLimitsToJSON(
	resources workerspec.ResourceRequestsLimits,
) *workerTemplateResourceLimitsJSON {
	return &workerTemplateResourceLimitsJSON{
		CPURequestMilliCPU:  resources.CPURequestMilliCPU,
		CPULimitMilliCPU:    resources.CPULimitMilliCPU,
		MemoryRequestBytes:  resources.MemoryRequestBytes,
		MemoryLimitBytes:    resources.MemoryLimitBytes,
		StorageRequestBytes: resources.StorageRequestBytes,
		StorageLimitBytes:   resources.StorageLimitBytes,
		GPURequest:          resources.GPURequest,
		GPULimit:            resources.GPULimit,
	}
}

func resourceLimitsFromJSON(
	resources workerTemplateResourceLimitsJSON,
) *workerspec.ResourceRequestsLimits {
	return &workerspec.ResourceRequestsLimits{
		CPURequestMilliCPU:  resources.CPURequestMilliCPU,
		CPULimitMilliCPU:    resources.CPULimitMilliCPU,
		MemoryRequestBytes:  resources.MemoryRequestBytes,
		MemoryLimitBytes:    resources.MemoryLimitBytes,
		StorageRequestBytes: resources.StorageRequestBytes,
		StorageLimitBytes:   resources.StorageLimitBytes,
		GPURequest:          resources.GPURequest,
		GPULimit:            resources.GPULimit,
	}
}
