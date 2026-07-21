package workercreation

import (
	"context"
	"errors"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/gitprovider"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdependencyartifact"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func (service *Service) PrepareSnapshotWithDependencies(
	ctx context.Context,
	scope specservice.Scope,
	snapshot specdomain.Snapshot,
	document workerdependency.Document,
) (PreparedSnapshot, error) {
	if service == nil || service.workerTypes == nil || scope.OrgID <= 0 ||
		scope.UserID <= 0 || snapshot.ID <= 0 ||
		snapshot.OrganizationID != scope.OrgID {
		return PreparedSnapshot{}, specservice.ErrResolverUnavailable
	}
	spec, err := specdomain.NormalizeAndValidate(snapshot.Spec)
	if err != nil {
		return PreparedSnapshot{}, err
	}
	document, err = workerdependency.NormalizeAndValidate(document)
	if err != nil {
		return PreparedSnapshot{}, err
	}
	if err := validateSnapshotArtifactBinding(scope, spec, document); err != nil {
		return PreparedSnapshot{}, err
	}
	resolver := newArtifactCompilationResolver(document)
	layer, err := newCompiler(resolver).Compile(ctx, scope, spec)
	if err != nil {
		return PreparedSnapshot{}, err
	}
	return PreparedSnapshot{
		Spec: spec, AgentfileLayer: layer,
		Repository:   artifactRepository(document.Repository, scope.OrgID),
		Dependencies: &document,
	}, nil
}

func validateSnapshotArtifactBinding(
	scope specservice.Scope,
	spec specdomain.Spec,
	document workerdependency.Document,
) error {
	if document.OrganizationID != scope.OrgID {
		return specservice.ErrInvalidScope
	}
	digest, err := canonicalWorkerSpecDigest(spec)
	if err != nil {
		return err
	}
	if document.Worker.SpecDigest != digest {
		return fmt.Errorf("%w: artifact spec digest mismatch", specservice.ErrInvalidDraft)
	}
	if err := workerdependencyartifact.ValidateWorkerSpecConsistency(spec, document); err != nil {
		return err
	}
	return validateArtifactSecretOwners(scope, document.SecretReferences)
}

func canonicalWorkerSpecDigest(spec specdomain.Spec) (string, error) {
	encoded, err := specdomain.EncodeSpec(spec)
	if err != nil {
		return "", err
	}
	canonical, err := control.CanonicalJSONObject(encoded)
	if err != nil {
		return "", err
	}
	return control.DigestCanonicalJSON(canonical)
}

func validateArtifactSecretOwners(
	scope specservice.Scope,
	secrets []workerdependency.SecretReference,
) error {
	for _, secret := range secrets {
		switch secret.OwnerScope {
		case envbundle.OwnerScopeUser:
			if secret.OwnerID != scope.UserID {
				return errors.New("user secret owner does not match plan actor")
			}
		case envbundle.OwnerScopeOrg:
			if secret.OwnerID != scope.OrgID {
				return errors.New("organization secret owner does not match scope")
			}
		default:
			return errors.New("secret owner scope is invalid")
		}
	}
	return nil
}

func artifactRepository(
	repository *workerdependency.Repository,
	orgID int64,
) *gitprovider.Repository {
	if repository == nil {
		return nil
	}
	timeout := int(repository.PreparationTimeoutSeconds)
	script := repository.PreparationScript
	return &gitprovider.Repository{
		ID: repository.Pin.DomainID, OrganizationID: orgID,
		Slug:         repository.Pin.Reference.Name.String(),
		Name:         repository.Pin.Reference.Name.String(),
		HttpCloneURL: repository.HTTPCloneURL, SshCloneURL: repository.SSHCloneURL,
		DefaultBranch: repository.Branch, PreparationScript: &script,
		PreparationTimeout: &timeout, IsActive: true,
	}
}

type artifactCompilationResolver struct {
	document workerdependency.Document
}

func newArtifactCompilationResolver(
	document workerdependency.Document,
) *artifactCompilationResolver {
	return &artifactCompilationResolver{document: document}
}

func (resolver *artifactCompilationResolver) ResolveCompilationReferences(
	_ context.Context,
	scope specservice.Scope,
	_ slugkit.Slug,
	_ specdomain.Workspace,
	secretRefs map[string]specdomain.SecretReference,
) (compilationReferences, error) {
	if err := validateArtifactSecretOwners(
		scope,
		resolver.document.SecretReferences,
	); err != nil {
		return compilationReferences{}, err
	}
	return resolver.compilationReferences(secretRefs)
}
