package podconnect

import (
	"encoding/json"
	"fmt"
	"sort"

	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
)

func workerDraftToProto(draft workercreation.Draft) (*podv1.WorkerSpecDraft, error) {
	values, err := json.Marshal(draft.WorkerSpec.TypeConfig.Values)
	if err != nil {
		return nil, fmt.Errorf("encode worker type config values: %w", err)
	}
	secretFields := make([]string, 0, len(draft.WorkerSpec.TypeConfig.SecretRefs))
	for field := range draft.WorkerSpec.TypeConfig.SecretRefs {
		secretFields = append(secretFields, field)
	}
	sort.Strings(secretFields)
	secretRefs := make([]*podv1.WorkerSecretReference, 0, len(secretFields))
	for _, field := range secretFields {
		reference := draft.WorkerSpec.TypeConfig.SecretRefs[field]
		secretRefs = append(secretRefs, &podv1.WorkerSecretReference{
			Field: field,
			Kind:  reference.Kind.String(),
			Id:    reference.ID,
		})
	}
	knowledge := make(
		[]*podv1.WorkerKnowledgeMount,
		0,
		len(draft.WorkerSpec.Workspace.KnowledgeMounts),
	)
	for _, mount := range draft.WorkerSpec.Workspace.KnowledgeMounts {
		knowledge = append(knowledge, &podv1.WorkerKnowledgeMount{
			KnowledgeBaseId: mount.KnowledgeBaseID,
			Mode:            string(mount.Mode),
		})
	}
	envBundleIDs := make([]int64, len(draft.WorkerSpec.Workspace.EnvBundleIDs))
	for index, id := range draft.WorkerSpec.Workspace.EnvBundleIDs {
		envBundleIDs[index] = int64(id)
	}
	return &podv1.WorkerSpecDraft{
		ModelResourceId:      draft.WorkerSpec.ModelResourceID,
		ToolModelResourceIds: cloneToolModelResourceIDs(draft.WorkerSpec.ToolModelResourceIDs),
		WorkerTypeSlug:       draft.WorkerSpec.WorkerTypeSlug.String(),
		RuntimeImageId:       draft.WorkerSpec.Runtime.RuntimeImageID,
		PlacementPolicy:      string(draft.WorkerSpec.Runtime.PlacementPolicy),
		ComputeTargetId:      draft.WorkerSpec.Runtime.ComputeTargetID,
		DeploymentMode:       string(draft.WorkerSpec.Runtime.DeploymentMode),
		ResourceProfileId:    draft.WorkerSpec.Runtime.ResourceProfileID,
		TypeSchemaVersion:    draft.WorkerSpec.TypeConfig.SchemaVersion,
		TypeConfigValuesJson: string(values),
		SecretRefs:           secretRefs,
		InteractionMode:      string(draft.WorkerSpec.TypeConfig.InteractionMode),
		AutomationLevel:      string(draft.WorkerSpec.TypeConfig.AutomationLevel),
		RepositoryId:         cloneInt64Pointer(draft.WorkerSpec.Workspace.RepositoryID),
		Branch:               draft.WorkerSpec.Workspace.Branch,
		SkillIds:             append([]int64{}, draft.WorkerSpec.Workspace.SkillIDs...),
		KnowledgeMounts:      knowledge,
		EnvBundleIds:         envBundleIDs,
		Instructions:         draft.WorkerSpec.Workspace.Instructions,
		InitialTask:          draft.WorkerSpec.Workspace.InitialTask,
		TerminationPolicy:    string(draft.WorkerSpec.Lifecycle.TerminationPolicy),
		IdleTimeoutMinutes:   draft.WorkerSpec.Lifecycle.IdleTimeoutMinutes,
		Alias:                draft.WorkerSpec.Metadata.Alias,
		SourceExpertId:       cloneInt64Pointer(draft.WorkerSpec.Metadata.SourceExpertID),
		OptionsRevision:      draft.OptionsRevision,
	}, nil
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
