package orchestrationcontrol

import (
	"context"
	"fmt"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
)

type RepositoryReferenceResolver struct {
	repository Repository
	authorizer Authorizer
}

func NewRepositoryReferenceResolver(
	repository Repository,
	authorizer Authorizer,
) (*RepositoryReferenceResolver, error) {
	if repository == nil || authorizer == nil {
		return nil, fmt.Errorf("%w: incomplete reference resolver", ErrUnavailable)
	}
	return &RepositoryReferenceResolver{
		repository: repository,
		authorizer: authorizer,
	}, nil
}

func (resolver *RepositoryReferenceResolver) Resolve(
	ctx context.Context,
	scope control.Scope,
	request DraftReference,
) (control.ResolvedReference, error) {
	if resolver == nil || resolver.repository == nil || resolver.authorizer == nil {
		return control.ResolvedReference{}, ErrUnavailable
	}
	if err := validateDraftReference(scope, request); err != nil {
		return control.ResolvedReference{}, err
	}
	target := referenceTarget(scope, request.Reference)
	head, err := resolver.repository.GetResource(ctx, scope, target)
	if err != nil {
		return control.ResolvedReference{}, err
	}
	if head.Identity.ResourceTarget != target {
		return control.ResolvedReference{}, control.ErrCorrupt
	}
	if err := resolver.authorizer.AuthorizeReference(ctx, scope, head); err != nil {
		return control.ResolvedReference{}, err
	}
	revisionNumber := request.Reference.Revision
	if revisionNumber == 0 {
		revisionNumber = head.Revision
	}
	revision, err := resolver.repository.GetRevision(
		ctx,
		scope,
		head.ID,
		revisionNumber,
	)
	if err != nil {
		return control.ResolvedReference{}, err
	}
	if err := validateReferenceRevision(scope, head, revision, revisionNumber); err != nil {
		return control.ResolvedReference{}, err
	}
	return control.ResolvedReference{
		TypeMeta:  head.Identity.TypeMeta,
		Namespace: head.Identity.Namespace,
		Name:      head.Identity.Name,
		UID:       head.Identity.UID,
		Revision:  revision.Revision,
		Digest:    revision.Digest,
	}, nil
}

func referenceTarget(
	scope control.Scope,
	reference orchestrationresource.Reference,
) control.ResourceTarget {
	apiVersion := reference.APIVersion
	if apiVersion == "" {
		apiVersion = orchestrationresource.APIVersionV1Alpha1
	}
	namespace := reference.Namespace
	if namespace == "" {
		namespace = scope.OrganizationSlug
	}
	return control.ResourceTarget{
		TypeMeta: orchestrationresource.TypeMeta{
			APIVersion: apiVersion,
			Kind:       reference.Kind,
		},
		Namespace: namespace,
		Name:      reference.Name,
	}
}

func validateReferenceRevision(
	scope control.Scope,
	head control.ResourceHead,
	revision control.ResourceRevision,
	expectedRevision int64,
) error {
	if err := revision.Validate(scope); err != nil {
		return fmt.Errorf("%w: invalid referenced revision", control.ErrCorrupt)
	}
	if revision.OrganizationID != head.OrganizationID ||
		revision.ResourceID != head.ID ||
		revision.Identity != head.Identity ||
		revision.Revision != expectedRevision ||
		revision.Revision > head.Revision ||
		revision.ResourceVersion > head.ResourceVersion {
		return control.ErrCorrupt
	}
	return nil
}
