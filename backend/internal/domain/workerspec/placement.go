package workerspec

type PlacementPolicy string

const (
	PlacementPolicyExplicit  PlacementPolicy = "explicit"
	PlacementPolicyAutomatic PlacementPolicy = "automatic"
)

type ComputeTargetKind string

const (
	ComputeTargetKindRunnerPool ComputeTargetKind = "runner-pool"
	ComputeTargetKindKubernetes ComputeTargetKind = "kubernetes"
)

type ComputeTarget struct {
	ID   int64             `json:"id"`
	Kind ComputeTargetKind `json:"kind"`
}

type DeploymentMode string

const (
	DeploymentModePooled    DeploymentMode = "pooled"
	DeploymentModeDedicated DeploymentMode = "dedicated"
)

type ResourceRequestsLimits struct {
	CPURequestMilliCPU  uint32  `json:"cpu_request_millicpu"`
	CPULimitMilliCPU    uint32  `json:"cpu_limit_millicpu"`
	MemoryRequestBytes  uint64  `json:"memory_request_bytes"`
	MemoryLimitBytes    uint64  `json:"memory_limit_bytes"`
	StorageRequestBytes uint64  `json:"storage_request_bytes,omitempty"`
	StorageLimitBytes   uint64  `json:"storage_limit_bytes,omitempty"`
	GPURequest          *uint32 `json:"gpu_request,omitempty"`
	GPULimit            *uint32 `json:"gpu_limit,omitempty"`
}

type ResourceProfile struct {
	ID        int64                  `json:"id"`
	Custom    bool                   `json:"custom,omitempty"`
	Resources ResourceRequestsLimits `json:"resources"`
}

type Placement struct {
	Policy          PlacementPolicy `json:"policy"`
	ComputeTarget   ComputeTarget   `json:"compute_target"`
	DeploymentMode  DeploymentMode  `json:"deployment_mode"`
	ResourceProfile ResourceProfile `json:"resource_profile"`
}
