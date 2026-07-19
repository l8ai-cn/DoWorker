package client

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func selectEchoWorkerSpec(
	options workerCreateOptionsWire,
	interactionMode, alias string,
) (WorkerSpecDraft, error) {
	for _, mode := range options.DeploymentModes {
		if mode.Value == "pooled" && mode.Selectable {
			return selectEchoRuntime(options, interactionMode, alias)
		}
	}
	return WorkerSpecDraft{}, fmt.Errorf("pooled deployment mode is unavailable")
}

func selectEchoRuntime(
	options workerCreateOptionsWire,
	interactionMode, alias string,
) (WorkerSpecDraft, error) {
	for _, workerType := range options.WorkerTypes {
		if workerType.Slug != echoWorkerType || !workerType.Selectable {
			continue
		}
		if workerType.RequiresModel {
			return WorkerSpecDraft{}, fmt.Errorf(
				"e2e-echo unexpectedly requires a model",
			)
		}
		if !slices.Contains(workerType.InteractionModes, interactionMode) {
			return WorkerSpecDraft{}, fmt.Errorf(
				"e2e-echo does not support %s interaction",
				interactionMode,
			)
		}
		values, err := workerTypeDefaults(workerType.ConfigSchemaJSON)
		if err != nil {
			return WorkerSpecDraft{}, err
		}
		imageID, err := selectableEchoImage(options)
		if err != nil {
			return WorkerSpecDraft{}, err
		}
		targetID, profileID, err := selectablePlacement(options)
		if err != nil {
			return WorkerSpecDraft{}, err
		}
		return WorkerSpecDraft{
			WorkerTypeSlug: echoWorkerType, RuntimeImageID: imageID,
			ComputeTargetID: targetID, ResourceProfileID: profileID,
			TypeSchemaVersion: workerType.SchemaVersion,
			TypeConfigValues:  values, InteractionMode: interactionMode,
			Alias: alias, OptionsRevision: options.Revision,
		}, nil
	}
	return WorkerSpecDraft{}, fmt.Errorf(
		"selectable e2e-echo worker type is unavailable: %s",
		workerOptionsSummary(options),
	)
}

func selectableEchoImage(options workerCreateOptionsWire) (int64, error) {
	for _, image := range options.RuntimeImages {
		if image.Selectable &&
			slices.Contains(image.WorkerTypes, echoWorkerType) {
			return parsePositiveID("runtime image", image.ID)
		}
	}
	return 0, fmt.Errorf(
		"selectable e2e-echo runtime image is unavailable: %s",
		workerOptionsSummary(options),
	)
}

func selectablePlacement(
	options workerCreateOptionsWire,
) (int64, int64, error) {
	var targetID int64
	for _, target := range options.ComputeTargets {
		if target.Selectable && target.SupportsPooled {
			var err error
			targetID, err = parsePositiveID("compute target", target.ID)
			if err != nil {
				return 0, 0, err
			}
			break
		}
	}
	if targetID == 0 {
		return 0, 0, fmt.Errorf(
			"selectable pooled compute target is unavailable: %s",
			workerOptionsSummary(options),
		)
	}
	for _, profile := range options.ResourceProfiles {
		if profile.Selectable {
			profileID, err := parsePositiveID(
				"resource profile",
				profile.ID,
			)
			return targetID, profileID, err
		}
	}
	return 0, 0, fmt.Errorf(
		"selectable resource profile is unavailable: %s",
		workerOptionsSummary(options),
	)
}

func workerOptionsSummary(options workerCreateOptionsWire) string {
	var parts []string
	for _, workerType := range options.WorkerTypes {
		if workerType.Slug == echoWorkerType {
			parts = append(parts, fmt.Sprintf(
				"workerType selectable=%t reason=%q modes=%v",
				workerType.Selectable,
				workerType.BlockingReason,
				workerType.InteractionModes,
			))
			break
		}
	}
	if len(parts) == 0 {
		parts = append(parts, fmt.Sprintf(
			"workerType missing among %d type(s)",
			len(options.WorkerTypes),
		))
	}
	parts = append(parts, fmt.Sprintf(
		"runtimeImages=%d computeTargets=%d deploymentModes=%d resourceProfiles=%d",
		len(options.RuntimeImages),
		len(options.ComputeTargets),
		len(options.DeploymentModes),
		len(options.ResourceProfiles),
	))
	return strings.Join(parts, "; ")
}

func parsePositiveID(field, value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%s id %q is invalid", field, value)
	}
	return id, nil
}

func workerTypeDefaults(raw string) (map[string]any, error) {
	var schema struct {
		Fields map[string]struct {
			Default  json.RawMessage `json:"default"`
			Required bool            `json:"required"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		return nil, fmt.Errorf(
			"decode e2e-echo config schema: %w",
			err,
		)
	}
	values := make(map[string]any)
	for name, field := range schema.Fields {
		if len(field.Default) == 0 {
			if field.Required {
				return nil, fmt.Errorf(
					"e2e-echo config %q has no default",
					name,
				)
			}
			continue
		}
		var value any
		if err := json.Unmarshal(field.Default, &value); err != nil {
			return nil, fmt.Errorf(
				"decode e2e-echo config %q default: %w",
				name,
				err,
			)
		}
		values[name] = value
	}
	return values, nil
}
