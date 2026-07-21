package orchestrationcontrol

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
)

func (service *Service) resolveReferences(
	ctx context.Context,
	scope control.Scope,
	draft validatedDraft,
) ([]control.ResolvedReference, error) {
	requests, err := draft.planner.References(draft.typedSpec)
	if err != nil {
		return nil, err
	}
	sort.Slice(requests, func(left, right int) bool {
		return requests[left].Path < requests[right].Path
	})
	resolvedByKey := make(map[string]control.ResolvedReference, len(requests))
	for _, request := range requests {
		if err := validateDraftReference(scope, request); err != nil {
			return nil, err
		}
		resolved, err := service.references.Resolve(ctx, scope, request)
		if err != nil {
			return nil, err
		}
		if err := validateResolvedReference(scope, request.Reference, resolved); err != nil {
			return nil, err
		}
		resolvedByKey[resolvedReferenceKey(resolved)] = resolved
	}
	resolved := make([]control.ResolvedReference, 0, len(resolvedByKey))
	for _, reference := range resolvedByKey {
		resolved = append(resolved, reference)
	}
	sort.Slice(resolved, func(left, right int) bool {
		return resolvedReferenceKey(resolved[left]) <
			resolvedReferenceKey(resolved[right])
	})
	return resolved, nil
}

func validateDraftReference(
	scope control.Scope,
	request DraftReference,
) error {
	pathCheck := control.PlanIssue{
		Severity: control.PlanIssueWarning,
		Path:     request.Path, Code: "reference-path", Message: "Reference path.",
	}
	if err := pathCheck.Validate(); err != nil {
		return err
	}
	return request.Reference.ValidateDraft(scope.OrganizationSlug.String())
}

func validateResolvedReference(
	scope control.Scope,
	requested orchestrationresource.Reference,
	resolved control.ResolvedReference,
) error {
	if err := resolved.Validate(scope); err != nil {
		return err
	}
	apiVersion := requested.APIVersion
	if apiVersion == "" {
		apiVersion = orchestrationresource.APIVersionV1Alpha1
	}
	namespace := requested.Namespace
	if namespace == "" {
		namespace = scope.OrganizationSlug
	}
	if resolved.APIVersion != apiVersion || resolved.Kind != requested.Kind ||
		resolved.Namespace != namespace || resolved.Name != requested.Name ||
		(requested.Revision > 0 && resolved.Revision != requested.Revision) {
		return fmt.Errorf("%w: reference resolver substituted identity", control.ErrCorrupt)
	}
	return nil
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
	}, "|")
}
