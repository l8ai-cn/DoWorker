package workerdependencyartifact

import (
	"fmt"
	"strconv"
	"strings"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
)

func validateReferenceClosure(
	scope control.Scope,
	planned []control.ResolvedReference,
	toolModels []ToolModelResolution,
	document workerdependency.Document,
) error {
	if err := validateDirectReferenceClosure(scope, planned, document); err != nil {
		return err
	}
	return validateToolModelResolutions(scope, toolModels, document)
}

func validateDirectReferenceClosure(
	scope control.Scope,
	planned []control.ResolvedReference,
	document workerdependency.Document,
) error {
	planIndex, err := indexPlanReferences(scope, planned)
	if err != nil {
		return err
	}
	direct := directDocumentReferences(document)
	for key := range planIndex {
		if _, exists := direct[key]; !exists {
			return fmt.Errorf("planned reference is absent from worker dependency artifact")
		}
	}
	for key := range direct {
		if _, exists := planIndex[key]; !exists {
			return fmt.Errorf("worker dependency artifact contains an unplanned reference")
		}
	}
	return nil
}

func indexPlanReferences(
	scope control.Scope,
	references []control.ResolvedReference,
) (map[string]struct{}, error) {
	index := make(map[string]struct{}, len(references))
	identities := make(map[string]struct{}, len(references))
	for _, reference := range references {
		if err := reference.Validate(scope); err != nil {
			return nil, fmt.Errorf("validate planned reference: %w", err)
		}
		identity := resolvedReferenceIdentity(reference)
		if _, exists := identities[identity]; exists {
			return nil, fmt.Errorf("duplicate planned reference identity revision")
		}
		identities[identity] = struct{}{}
		key := resolvedReferenceKey(reference)
		index[key] = struct{}{}
	}
	return index, nil
}

func directDocumentReferences(
	document workerdependency.Document,
) map[string]struct{} {
	references := make(map[string]struct{})
	add := func(reference resource.Reference) {
		references[referenceKey(reference)] = struct{}{}
	}
	if document.Models.Primary != nil {
		add(document.Models.Primary.Pin.Reference)
	}
	for _, tool := range document.Models.Tools {
		add(tool.Binding)
		add(tool.Model.Pin.Reference)
	}
	if document.Repository != nil {
		add(document.Repository.Pin.Reference)
	}
	for _, skill := range document.Skills {
		add(skill.Pin.Reference)
	}
	for _, knowledge := range document.KnowledgeBases {
		add(knowledge.Pin.Reference)
	}
	for _, bundle := range document.RuntimeBundles {
		add(bundle.Pin.Reference)
	}
	for _, secret := range document.SecretReferences {
		add(secret.Pin.Reference)
	}
	add(document.Placement.ComputeTarget.Reference)
	if document.Placement.ResourceProfile != nil {
		add(document.Placement.ResourceProfile.Reference)
	}
	return references
}

func resolvedReferenceKey(reference control.ResolvedReference) string {
	return strings.Join([]string{
		reference.APIVersion,
		reference.Kind,
		reference.Namespace.String(),
		reference.Name.String(),
		reference.UID,
		strconv.FormatInt(reference.Revision, 10),
		reference.Digest,
	}, "\x00")
}

func resolvedReferenceIdentity(reference control.ResolvedReference) string {
	return strings.Join([]string{
		reference.APIVersion,
		reference.Kind,
		reference.Namespace.String(),
		reference.Name.String(),
		reference.UID,
		strconv.FormatInt(reference.Revision, 10),
	}, "\x00")
}

func referenceKey(reference resource.Reference) string {
	return strings.Join([]string{
		reference.APIVersion,
		reference.Kind,
		reference.Namespace.String(),
		reference.Name.String(),
		reference.UID,
		strconv.FormatInt(reference.Revision, 10),
		reference.Digest,
	}, "\x00")
}
