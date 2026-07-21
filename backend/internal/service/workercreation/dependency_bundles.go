package workercreation

import (
	"fmt"
	"sort"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdependencyartifact"
)

func buildBundleResolutions(
	scope control.Scope,
	refs ArtifactReferences,
	spec workerspec.Spec,
	definition workerdefinition.Definition,
	workspace *workspaceResolver,
) ([]workerdependencyartifact.RuntimeBundleResolution, error) {
	bundles := make([]workerdependencyartifact.RuntimeBundleResolution, 0)
	for _, id := range spec.Workspace.EnvBundleIDs {
		bundle, err := runtimeBundle(scope, refs.RuntimeBundles[int64(id)], int64(id), nil, workspace)
		if err != nil {
			return nil, err
		}
		bundles = append(bundles, bundle)
	}
	for _, binding := range spec.Workspace.ConfigDocumentBindings {
		document, err := configDocument(definition, binding.DocumentID)
		if err != nil {
			return nil, err
		}
		bundle, err := runtimeBundle(
			scope,
			refs.ConfigBundles[binding.ConfigBundleID],
			binding.ConfigBundleID,
			document,
			workspace,
		)
		if err != nil {
			return nil, err
		}
		bundles = append(bundles, bundle)
	}
	return bundles, nil
}

func runtimeBundle(
	scope control.Scope,
	reference control.ResolvedReference,
	id int64,
	document *workerdependencyartifact.ConfigDocumentResolution,
	workspace *workspaceResolver,
) (workerdependencyartifact.RuntimeBundleResolution, error) {
	bundle := workspace.envBundles[id]
	if bundle == nil {
		return workerdependencyartifact.RuntimeBundleResolution{}, fmt.Errorf(
			"WorkerTemplate artifact environment bundle %d was not resolved",
			id,
		)
	}
	pin, err := referencePin(scope, reference, id)
	if err != nil {
		return workerdependencyartifact.RuntimeBundleResolution{}, err
	}
	values := runtimeValues(bundle.Data)
	digest, err := workerdependency.DigestRuntimeValues(materializedRuntimeValues(values))
	if err != nil {
		return workerdependencyartifact.RuntimeBundleResolution{}, err
	}
	return workerdependencyartifact.RuntimeBundleResolution{
		ResourceResolution: pin, Kind: bundle.Kind, ContentDigest: digest,
		Values: values, ConfigDocument: document,
	}, nil
}

func runtimeValues(data envbundle.BundleData) []workerdependencyartifact.RuntimeValueResolution {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	values := make([]workerdependencyartifact.RuntimeValueResolution, len(keys))
	for index, key := range keys {
		values[index] = workerdependencyartifact.RuntimeValueResolution{
			Name: key, Value: data[key],
		}
	}
	return values
}
