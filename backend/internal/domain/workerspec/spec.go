package workerspec

type Spec struct {
	Version    Version    `json:"version"`
	Runtime    Runtime    `json:"runtime"`
	Placement  Placement  `json:"placement"`
	TypeConfig TypeConfig `json:"type_config"`
	Workspace  Workspace  `json:"workspace"`
	Lifecycle  Lifecycle  `json:"lifecycle"`
	Metadata   Metadata   `json:"metadata"`
}

func NewV1(
	runtime Runtime,
	typeConfig TypeConfig,
	workspace Workspace,
	lifecycle Lifecycle,
	metadata Metadata,
) Spec {
	return Spec{
		Version:    VersionV1,
		Runtime:    runtime,
		TypeConfig: typeConfig,
		Workspace:  workspace,
		Lifecycle:  lifecycle,
		Metadata:   metadata,
	}
}
