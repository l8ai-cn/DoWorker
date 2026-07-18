package client

import (
	"context"
	"encoding/json"
	"strconv"
)

const echoWorkerType = "e2e-echo"

type WorkerSpecDraft struct {
	WorkerTypeSlug    string
	RuntimeImageID    int64
	ComputeTargetID   int64
	ResourceProfileID int64
	TypeSchemaVersion uint32
	TypeConfigValues  map[string]any
	InteractionMode   string
	Alias             string
	OptionsRevision   string
}

type workerSpecWire struct {
	ModelResourceID      string            `json:"modelResourceId"`
	WorkerTypeSlug       string            `json:"workerTypeSlug"`
	RuntimeImageID       string            `json:"runtimeImageId"`
	PlacementPolicy      string            `json:"placementPolicy"`
	ComputeTargetID      string            `json:"computeTargetId"`
	DeploymentMode       string            `json:"deploymentMode"`
	ResourceProfileID    string            `json:"resourceProfileId"`
	TypeSchemaVersion    uint32            `json:"typeSchemaVersion"`
	TypeConfigValuesJSON string            `json:"typeConfigValuesJson"`
	SecretRefs           []any             `json:"secretRefs"`
	InteractionMode      string            `json:"interactionMode"`
	AutomationLevel      string            `json:"automationLevel"`
	SkillIDs             []string          `json:"skillIds"`
	KnowledgeMounts      []any             `json:"knowledgeMounts"`
	EnvBundleIDs         []string          `json:"envBundleIds"`
	TerminationPolicy    string            `json:"terminationPolicy"`
	Alias                string            `json:"alias"`
	OptionsRevision      string            `json:"optionsRevision"`
	ToolModelResourceIDs map[string]string `json:"toolModelResourceIds"`
	ConfigBindings       []any             `json:"configDocumentBindings"`
}

func (draft WorkerSpecDraft) wire() workerSpecWire {
	values, _ := json.Marshal(draft.TypeConfigValues)
	return workerSpecWire{
		ModelResourceID: "0", WorkerTypeSlug: draft.WorkerTypeSlug,
		RuntimeImageID:       strconv.FormatInt(draft.RuntimeImageID, 10),
		PlacementPolicy:      "automatic",
		ComputeTargetID:      strconv.FormatInt(draft.ComputeTargetID, 10),
		DeploymentMode:       "pooled",
		ResourceProfileID:    strconv.FormatInt(draft.ResourceProfileID, 10),
		TypeSchemaVersion:    draft.TypeSchemaVersion,
		TypeConfigValuesJSON: string(values),
		SecretRefs:           []any{}, InteractionMode: draft.InteractionMode,
		AutomationLevel: "interactive", SkillIDs: []string{},
		KnowledgeMounts: []any{}, EnvBundleIDs: []string{},
		TerminationPolicy: "manual", Alias: draft.Alias,
		OptionsRevision:      draft.OptionsRevision,
		ToolModelResourceIDs: map[string]string{},
		ConfigBindings:       []any{},
	}
}

type workerCreateOptionsWire struct {
	Revision    string `json:"revision"`
	WorkerTypes []struct {
		Slug             string   `json:"slug"`
		SchemaVersion    uint32   `json:"schemaVersion"`
		ConfigSchemaJSON string   `json:"configSchemaJson"`
		Selectable       bool     `json:"selectable"`
		RequiresModel    bool     `json:"requiresModelResource"`
		InteractionModes []string `json:"supportedInteractionModes"`
	} `json:"workerTypes"`
	RuntimeImages []struct {
		ID          string   `json:"id"`
		WorkerTypes []string `json:"workerTypeSlugs"`
		Selectable  bool     `json:"selectable"`
	} `json:"runtimeImages"`
	ComputeTargets []struct {
		ID             string `json:"id"`
		SupportsPooled bool   `json:"supportsPooled"`
		Selectable     bool   `json:"selectable"`
	} `json:"computeTargets"`
	DeploymentModes []struct {
		Value      string `json:"value"`
		Selectable bool   `json:"selectable"`
	} `json:"deploymentModes"`
	ResourceProfiles []struct {
		ID         string `json:"id"`
		Selectable bool   `json:"selectable"`
	} `json:"resourceProfiles"`
}

func (r *REST) BuildEchoWorkerSpec(
	ctx context.Context,
	orgSlug, interactionMode, alias string,
) (WorkerSpecDraft, error) {
	var options workerCreateOptionsWire
	request := map[string]string{
		"orgSlug": orgSlug, "workerTypeSlug": echoWorkerType,
	}
	if err := r.connectCall(
		ctx,
		"/proto.pod.v1.PodService/ListWorkerCreateOptions",
		request,
		&options,
	); err != nil {
		return WorkerSpecDraft{}, err
	}
	return selectEchoWorkerSpec(options, interactionMode, alias)
}
