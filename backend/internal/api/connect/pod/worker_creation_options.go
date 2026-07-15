package podconnect

import (
	"encoding/json"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
)

type workerTypeSchemaJSON struct {
	Version uint32                               `json:"version"`
	Fields  map[string]workerTypeFieldSchemaJSON `json:"fields"`
}

type workerTypeFieldSchemaJSON struct {
	Kind    specdomain.TypeFieldKind `json:"kind"`
	Options []string                 `json:"options,omitempty"`
}

func workerCreateOptionsToProto(
	options workercreation.CreateOptions,
) (*podv1.ListWorkerCreateOptionsResponse, error) {
	response := &podv1.ListWorkerCreateOptionsResponse{Revision: options.Revision}
	for _, option := range options.WorkerTypes {
		schema, err := encodeWorkerTypeSchema(option.Schema)
		if err != nil {
			return nil, err
		}
		response.WorkerTypes = append(response.WorkerTypes, &podv1.WorkerTypeOption{
			Slug:                  option.Slug,
			Name:                  option.Name,
			Description:           option.Description,
			SchemaVersion:         option.Schema.Version,
			ConfigSchemaJson:      schema,
			RequiresModelResource: option.RequiresModelResource,
			ModelProtocolAdapters: workerSlugsToStrings(
				option.ModelProtocolAdapters,
			),
			ToolModelRequirements: workerToolModelRequirementsToProto(
				option.ToolModelRequirements,
			),
			Selectable:     option.Selectable,
			BlockingReason: option.BlockingReason,
		})
	}
	for _, option := range options.RuntimeImages {
		response.RuntimeImages = append(response.RuntimeImages, &podv1.WorkerRuntimeImageOption{
			Id:              option.Image.ID,
			Slug:            option.Image.Slug,
			Name:            option.Image.Name,
			Reference:       option.Image.Reference,
			Digest:          option.Image.Digest,
			WorkerTypeSlugs: append([]string{}, option.Image.WorkerTypeSlugs...),
			Selectable:      option.Selectable,
			BlockingReason:  option.BlockingReason,
		})
	}
	for _, option := range options.ComputeTargets {
		response.ComputeTargets = append(response.ComputeTargets, &podv1.WorkerComputeTargetOption{
			Id:                option.Target.ID,
			Slug:              option.Target.Slug,
			Name:              option.Target.Name,
			Kind:              string(option.Target.Kind),
			SupportsPooled:    option.Target.SupportsPooled,
			SupportsDedicated: option.Target.SupportsDedicated,
			Selectable:        option.Selectable,
			BlockingReason:    option.BlockingReason,
		})
	}
	for _, option := range options.DeploymentModes {
		response.DeploymentModes = append(response.DeploymentModes, &podv1.WorkerDeploymentModeOption{
			Value:          string(option.Value),
			Name:           option.Name,
			Selectable:     option.Selectable,
			BlockingReason: option.BlockingReason,
		})
	}
	for _, option := range options.ResourceProfiles {
		resources := option.Profile.Resources
		response.ResourceProfiles = append(response.ResourceProfiles, &podv1.WorkerResourceProfileOption{
			Id:                 option.Profile.ID,
			Slug:               option.Profile.Slug,
			Name:               option.Profile.Name,
			CpuRequestMillicpu: resources.CPURequestMilliCPU,
			CpuLimitMillicpu:   resources.CPULimitMilliCPU,
			MemoryRequestBytes: resources.MemoryRequestBytes,
			MemoryLimitBytes:   resources.MemoryLimitBytes,
			GpuRequest:         cloneUint32Pointer(resources.GPURequest),
			GpuLimit:           cloneUint32Pointer(resources.GPULimit),
			Selectable:         option.Selectable,
			BlockingReason:     option.BlockingReason,
		})
	}
	return response, nil
}

func workerSlugsToStrings(values []slugkit.Slug) []string {
	items := make([]string, len(values))
	for index, value := range values {
		items[index] = value.String()
	}
	return items
}

func workerToolModelRequirementsToProto(
	requirements []specdomain.ToolModelRequirement,
) []*podv1.WorkerToolModelRequirement {
	items := make([]*podv1.WorkerToolModelRequirement, 0, len(requirements))
	for _, requirement := range requirements {
		providers := make([]string, len(requirement.ProviderKeys))
		for index, provider := range requirement.ProviderKeys {
			providers[index] = provider.String()
		}
		adapters := make([]string, len(requirement.ProtocolAdapters))
		for index, adapter := range requirement.ProtocolAdapters {
			adapters[index] = adapter.String()
		}
		items = append(items, &podv1.WorkerToolModelRequirement{
			Role:             requirement.Role.String(),
			ProviderKeys:     providers,
			ProtocolAdapters: adapters,
			Modality:         string(requirement.Modality),
			Capability:       string(requirement.Capability),
		})
	}
	return items
}

func encodeWorkerTypeSchema(schema specdomain.TypeSchema) (string, error) {
	fields := make(map[string]workerTypeFieldSchemaJSON, len(schema.Fields))
	for name, field := range schema.Fields {
		fields[name] = workerTypeFieldSchemaJSON{
			Kind:    field.Kind,
			Options: append([]string{}, field.Options...),
		}
	}
	data, err := json.Marshal(workerTypeSchemaJSON{
		Version: schema.Version,
		Fields:  fields,
	})
	return string(data), err
}

func workerIssuesToProto(groups ...[]workercreation.Issue) []*podv1.WorkerPreflightIssue {
	var issues []*podv1.WorkerPreflightIssue
	for _, group := range groups {
		for _, issue := range group {
			issues = append(issues, &podv1.WorkerPreflightIssue{
				Code:     issue.Code,
				Field:    issue.Field,
				Message:  issue.Message,
				Severity: issue.Severity,
			})
		}
	}
	return issues
}

func cloneUint32Pointer(value *uint32) *uint32 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
