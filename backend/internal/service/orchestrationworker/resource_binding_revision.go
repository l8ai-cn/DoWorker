package orchestrationworker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
)

type loadedBindingRevision struct {
	revision control.ResourceRevision
	spec     any
}

func (resolver *ResourceBindingResolver) loadBindingRevision(
	ctx context.Context,
	scope control.Scope,
	reference control.ResolvedReference,
) (loadedBindingRevision, error) {
	if resolver == nil || resolver.registry == nil ||
		resolver.repository == nil || resolver.authorizer == nil {
		return loadedBindingRevision{}, control.ErrCorrupt
	}
	if err := scope.Validate(); err != nil {
		return loadedBindingRevision{}, err
	}
	if err := reference.Validate(scope); err != nil {
		return loadedBindingRevision{}, err
	}
	target := control.ResourceTarget{
		TypeMeta:  reference.TypeMeta,
		Namespace: reference.Namespace,
		Name:      reference.Name,
	}
	head, err := resolver.repository.GetResource(ctx, scope, target)
	if err != nil {
		return loadedBindingRevision{}, err
	}
	if err := validateBindingHead(scope, target, reference, head); err != nil {
		return loadedBindingRevision{}, err
	}
	if err := resolver.authorizer.AuthorizeReference(ctx, scope, head); err != nil {
		return loadedBindingRevision{}, err
	}
	revision, err := resolver.repository.GetRevision(
		ctx,
		scope,
		head.ID,
		reference.Revision,
	)
	if errors.Is(err, control.ErrNotFound) {
		return loadedBindingRevision{}, control.ErrCorrupt
	}
	if err != nil {
		return loadedBindingRevision{}, err
	}
	if err := validateBindingRevision(scope, head, reference, revision); err != nil {
		return loadedBindingRevision{}, err
	}
	manifest, err := decodeStoredBindingManifest(revision.CanonicalManifest)
	if err != nil {
		return loadedBindingRevision{}, control.ErrCorrupt
	}
	spec, err := resolver.registry.DecodeAndValidate(manifest)
	if err != nil {
		return loadedBindingRevision{}, control.ErrCorrupt
	}
	return loadedBindingRevision{revision: revision, spec: spec}, nil
}

func validateBindingHead(
	scope control.Scope,
	target control.ResourceTarget,
	reference control.ResolvedReference,
	head control.ResourceHead,
) error {
	if err := head.Validate(scope); err != nil {
		return control.ErrCorrupt
	}
	if head.Identity.ResourceTarget != target ||
		head.Identity.UID != reference.UID ||
		reference.Revision > head.Revision {
		return control.ErrCorrupt
	}
	return nil
}

func validateBindingRevision(
	scope control.Scope,
	head control.ResourceHead,
	reference control.ResolvedReference,
	revision control.ResourceRevision,
) error {
	if err := revision.Validate(scope); err != nil {
		return control.ErrCorrupt
	}
	if revision.OrganizationID != head.OrganizationID ||
		revision.ResourceID != head.ID ||
		revision.Identity != head.Identity ||
		revision.Revision != reference.Revision ||
		revision.Digest != reference.Digest ||
		revision.Revision > head.Revision ||
		revision.Generation > head.Generation ||
		revision.ResourceVersion > head.ResourceVersion {
		return control.ErrCorrupt
	}
	return nil
}

func decodeStoredBindingManifest(
	source json.RawMessage,
) (resource.Manifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var manifest resource.Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return resource.Manifest{}, fmt.Errorf("decode stored manifest: %w", err)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return resource.Manifest{}, fmt.Errorf("stored manifest has trailing data")
	}
	return manifest, nil
}
