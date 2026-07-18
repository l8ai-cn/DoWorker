package orchestrationworker

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

func workerTemplateReferences(
	spec resource.WorkerTemplateSpec,
) []controlservice.DraftReference {
	references := make([]controlservice.DraftReference, 0)
	appendReference := func(path string, reference resource.Reference) {
		references = append(references, controlservice.DraftReference{
			Path: path, Reference: reference,
		})
	}
	if spec.ModelRef != nil {
		appendReference("/spec/modelRef", *spec.ModelRef)
	}
	appendReference("/spec/runtime/computeTargetRef", spec.Runtime.ComputeTargetRef)
	if spec.Runtime.ResourceProfileRef != nil {
		appendReference("/spec/runtime/resourceProfileRef", *spec.Runtime.ResourceProfileRef)
	}
	for role, reference := range spec.ToolRefs {
		appendReference("/spec/toolRefs/"+role, reference)
	}
	for field, reference := range spec.TypeConfig.SecretRefs {
		appendReference("/spec/typeConfig/secretRefs/"+field, reference)
	}
	if spec.Workspace.RepositoryRef != nil {
		appendReference("/spec/workspace/repositoryRef", *spec.Workspace.RepositoryRef)
	}
	for index, reference := range spec.Workspace.SkillRefs {
		appendReference(fmt.Sprintf("/spec/workspace/skillRefs/%d", index), reference)
	}
	for index, mount := range spec.Workspace.KnowledgeMounts {
		appendReference(fmt.Sprintf("/spec/workspace/knowledgeMounts/%d/ref", index), mount.Ref)
	}
	for index, reference := range spec.Workspace.EnvironmentBundleRefs {
		appendReference(fmt.Sprintf(
			"/spec/workspace/environmentBundleRefs/%d",
			index,
		), reference)
	}
	for index, binding := range spec.Workspace.ConfigDocumentBindings {
		appendReference(fmt.Sprintf(
			"/spec/workspace/configDocumentBindings/%d/configBundleRef",
			index,
		), binding.ConfigBundleRef)
	}
	sort.Slice(references, func(left, right int) bool {
		return references[left].Path < references[right].Path
	})
	return references
}

type pinnedReferenceIndex struct {
	scope      control.Scope
	references map[string]control.ResolvedReference
	revisions  map[string][]control.ResolvedReference
}

func newPinnedReferenceIndex(
	scope control.Scope,
	references []control.ResolvedReference,
) (pinnedReferenceIndex, error) {
	index := pinnedReferenceIndex{
		scope:      scope,
		references: make(map[string]control.ResolvedReference, len(references)),
		revisions:  make(map[string][]control.ResolvedReference, len(references)),
	}
	for _, reference := range references {
		if err := reference.Validate(scope); err != nil {
			return pinnedReferenceIndex{}, err
		}
		identity := resolvedReferenceIdentity(reference)
		key := resolvedReferenceRevisionIdentity(reference)
		if _, exists := index.references[key]; exists {
			return pinnedReferenceIndex{}, control.ErrCorrupt
		}
		index.references[key] = reference
		index.revisions[identity] = append(index.revisions[identity], reference)
	}
	return index, nil
}

func (index pinnedReferenceIndex) resolve(
	reference resource.Reference,
) (control.ResolvedReference, error) {
	target := normalizedDraftReference(index.scope, reference)
	identity := draftReferenceIdentity(target)
	if target.Revision == 0 {
		revisions := index.revisions[identity]
		if len(revisions) != 1 {
			return control.ResolvedReference{}, control.ErrCorrupt
		}
		return revisions[0], nil
	}
	resolved, exists := index.references[draftReferenceRevisionIdentity(target)]
	if !exists {
		return control.ResolvedReference{}, control.ErrCorrupt
	}
	return resolved, nil
}

func normalizedDraftReference(
	scope control.Scope,
	reference resource.Reference,
) resource.Reference {
	if reference.APIVersion == "" {
		reference.APIVersion = resource.APIVersionV1Alpha1
	}
	if reference.Namespace == "" {
		reference.Namespace = scope.OrganizationSlug
	}
	return reference
}

func draftReferenceIdentity(reference resource.Reference) string {
	return strings.Join([]string{
		reference.APIVersion, reference.Kind, reference.Namespace.String(),
		reference.Name.String(),
	}, "\x00")
}

func resolvedReferenceIdentity(reference control.ResolvedReference) string {
	return strings.Join([]string{
		reference.APIVersion, reference.Kind, reference.Namespace.String(),
		reference.Name.String(),
	}, "\x00")
}

func draftReferenceRevisionIdentity(reference resource.Reference) string {
	return draftReferenceIdentity(reference) + "\x00" +
		strconv.FormatInt(reference.Revision, 10)
}

func resolvedReferenceRevisionIdentity(reference control.ResolvedReference) string {
	return resolvedReferenceIdentity(reference) + "\x00" +
		strconv.FormatInt(reference.Revision, 10)
}
