package workercreation

import (
	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func computeTargetOptions(catalog runtimedomain.Catalog) []ComputeTargetOption {
	targets := catalog.Targets()
	options := make([]ComputeTargetOption, 0, len(targets))
	for _, target := range targets {
		option := ComputeTargetOption{Target: target, Selectable: target.Enabled}
		if !target.Enabled {
			option.BlockingReason = target.DisabledReason
			if option.BlockingReason == "" {
				option.BlockingReason = "Compute target is disabled"
			}
		}
		options = append(options, option)
	}
	return options
}

func deploymentModeOptions(
	catalog runtimedomain.Catalog,
	computeTargetID *int64,
) []DeploymentModeOption {
	return []DeploymentModeOption{
		deploymentModeOption(
			catalog,
			computeTargetID,
			specdomain.DeploymentModePooled,
			"Pooled",
		),
		deploymentModeOption(
			catalog,
			computeTargetID,
			specdomain.DeploymentModeDedicated,
			"Dedicated",
		),
	}
}

func deploymentModeOption(
	catalog runtimedomain.Catalog,
	computeTargetID *int64,
	mode specdomain.DeploymentMode,
	name string,
) DeploymentModeOption {
	option := DeploymentModeOption{Value: mode, Name: name}
	if computeTargetID == nil {
		for _, target := range catalog.Targets() {
			if target.Enabled && supportsDeploymentMode(target, mode) {
				option.Selectable = true
				return option
			}
		}
		option.BlockingReason = "No enabled compute target supports this deployment mode"
		return option
	}
	target := catalog.Target(*computeTargetID)
	if target == nil || !target.Enabled {
		option.BlockingReason = "Selected compute target is unavailable"
		return option
	}
	option.Selectable = supportsDeploymentMode(*target, mode)
	if !option.Selectable {
		option.BlockingReason = "Selected compute target does not support this deployment mode"
	}
	return option
}

func resourceProfileOptions(catalog runtimedomain.Catalog) []ResourceProfileOption {
	profiles := catalog.Profiles()
	options := make([]ResourceProfileOption, 0, len(profiles))
	for _, profile := range profiles {
		option := ResourceProfileOption{Profile: profile, Selectable: profile.Enabled}
		if !profile.Enabled {
			option.BlockingReason = "Resource profile is disabled"
		}
		options = append(options, option)
	}
	return options
}
