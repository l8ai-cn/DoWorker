package podconnect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
)

func workerDraftFromProto(message *podv1.WorkerSpecDraft) (workercreation.Draft, error) {
	if message == nil {
		return workercreation.Draft{}, invalidWorkerDraft("draft", "is required")
	}
	workerType, err := slugkit.NewFromTrusted(message.GetWorkerTypeSlug())
	if err != nil {
		return workercreation.Draft{}, invalidWorkerDraft("worker_type_slug", err.Error())
	}
	values, err := decodeTypeConfigValues(message.GetTypeConfigValuesJson())
	if err != nil {
		return workercreation.Draft{}, invalidWorkerDraft("type_config_values_json", err.Error())
	}
	secretRefs, err := workerSecretRefsFromProto(message.GetSecretRefs())
	if err != nil {
		return workercreation.Draft{}, err
	}
	knowledge := make([]specdomain.KnowledgeMount, 0, len(message.GetKnowledgeMounts()))
	for _, mount := range message.GetKnowledgeMounts() {
		if mount == nil {
			return workercreation.Draft{}, invalidWorkerDraft("knowledge_mounts", "contains an empty item")
		}
		knowledge = append(knowledge, specdomain.KnowledgeMount{
			KnowledgeBaseID: mount.GetKnowledgeBaseId(),
			Mode:            specdomain.KnowledgeMountMode(mount.GetMode()),
		})
	}
	envBundleIDs := make(
		[]specdomain.RuntimeEnvBundleID,
		len(message.GetEnvBundleIds()),
	)
	for index, id := range message.GetEnvBundleIds() {
		envBundleIDs[index] = specdomain.RuntimeEnvBundleID(id)
	}
	var customResources *specdomain.ResourceRequestsLimits
	if resource := message.GetCustomResources(); resource != nil {
		customResources = &specdomain.ResourceRequestsLimits{
			CPURequestMilliCPU:  resource.GetCpuRequestMillicpu(),
			CPULimitMilliCPU:    resource.GetCpuLimitMillicpu(),
			MemoryRequestBytes:  resource.GetMemoryRequestBytes(),
			MemoryLimitBytes:    resource.GetMemoryLimitBytes(),
			StorageRequestBytes: resource.GetStorageRequestBytes(),
			StorageLimitBytes:   resource.GetStorageLimitBytes(),
		}
	}
	return workercreation.Draft{
		OptionsRevision: message.GetOptionsRevision(),
		WorkerSpec: specservice.Draft{
			ModelResourceID:      message.GetModelResourceId(),
			ToolModelResourceIDs: cloneToolModelResourceIDs(message.GetToolModelResourceIds()),
			WorkerTypeSlug:       workerType,
			Runtime: specservice.RuntimeSelection{
				RuntimeImageID:    message.GetRuntimeImageId(),
				PlacementPolicy:   specdomain.PlacementPolicy(message.GetPlacementPolicy()),
				ComputeTargetID:   message.GetComputeTargetId(),
				DeploymentMode:    workerDeploymentMode(message.GetDeploymentMode()),
				ResourceProfileID: message.GetResourceProfileId(),
				CustomResources:   customResources,
			},
			TypeConfig: specdomain.TypeConfig{
				SchemaVersion:   message.GetTypeSchemaVersion(),
				Values:          values,
				SecretRefs:      secretRefs,
				InteractionMode: specdomain.InteractionMode(message.GetInteractionMode()),
				AutomationLevel: specdomain.AutomationLevel(message.GetAutomationLevel()),
			},
			Workspace: specdomain.Workspace{
				RepositoryID:    optionalInt64(message.RepositoryId),
				Branch:          message.GetBranch(),
				SkillIDs:        append([]int64{}, message.GetSkillIds()...),
				KnowledgeMounts: knowledge,
				EnvBundleIDs:    envBundleIDs,
				ConfigBundleIDs: append([]int64{}, message.GetConfigBundleIds()...),
				Instructions:    message.GetInstructions(),
				InitialTask:     message.GetInitialTask(),
			},
			Lifecycle: specdomain.Lifecycle{
				TerminationPolicy:  specdomain.TerminationPolicy(message.GetTerminationPolicy()),
				IdleTimeoutMinutes: message.GetIdleTimeoutMinutes(),
			},
			Metadata: specdomain.Metadata{
				Alias:          message.GetAlias(),
				SourceExpertID: optionalInt64(message.SourceExpertId),
			},
		},
	}, nil
}

func cloneToolModelResourceIDs(values map[string]int64) map[string]int64 {
	if values == nil {
		return nil
	}
	cloned := make(map[string]int64, len(values))
	for role, id := range values {
		cloned[role] = id
	}
	return cloned
}

func decodeTypeConfigValues(raw string) (map[string]any, error) {
	if raw == "" {
		raw = "{}"
	}
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.UseNumber()
	var values map[string]any
	if err := decoder.Decode(&values); err != nil {
		return nil, err
	}
	if values == nil {
		return nil, errors.New("must be a JSON object")
	}
	var trailing any
	switch err := decoder.Decode(&trailing); {
	case errors.Is(err, io.EOF):
		return values, nil
	case err == nil:
		return nil, errors.New("contains trailing JSON data")
	default:
		return nil, err
	}
}

func workerSecretRefsFromProto(
	items []*podv1.WorkerSecretReference,
) (map[string]specdomain.SecretReference, error) {
	references := make(map[string]specdomain.SecretReference, len(items))
	for _, item := range items {
		if item == nil {
			return nil, invalidWorkerDraft("secret_refs", "contains an empty item")
		}
		if _, exists := references[item.GetField()]; exists {
			return nil, invalidWorkerDraft(
				"secret_refs",
				fmt.Sprintf("contains duplicate field %q", item.GetField()),
			)
		}
		kind, err := slugkit.NewFromTrusted(item.GetKind())
		if err != nil {
			return nil, invalidWorkerDraft("secret_refs", err.Error())
		}
		references[item.GetField()] = specdomain.SecretReference{
			Kind: kind,
			ID:   item.GetId(),
		}
	}
	return references, nil
}

func invalidWorkerDraft(field, reason string) error {
	return &specservice.InvalidDraftFieldError{Field: field, Reason: reason}
}

func workerDeploymentMode(value string) specdomain.DeploymentMode {
	return specdomain.DeploymentMode(value)
}
