package workercreation

import (
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/google/uuid"
)

func freshResolvedReference(
	scope specservice.Scope,
	namespace slugkit.Slug,
	kind string,
	name slugkit.Slug,
	revision int64,
	payload map[string]any,
) (control.ResolvedReference, error) {
	if revision <= 0 {
		return control.ResolvedReference{}, invalidFreshReference(kind, "revision is missing")
	}
	if err := slugkit.Validate(namespace.String()); err != nil {
		return control.ResolvedReference{}, err
	}
	payload["kind"] = kind
	payload["name"] = name.String()
	digest, err := digestFreshReference(payload)
	if err != nil {
		return control.ResolvedReference{}, err
	}
	uid := uuid.NewSHA1(uuid.NameSpaceOID, []byte(kind+"/"+name.String()+"/"+digest))
	reference := control.ResolvedReference{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       kind,
		},
		Namespace: namespace, Name: name, UID: uid.String(),
		Revision: revision, Digest: digest,
	}
	if err := reference.Validate(control.Scope{
		OrganizationID: scope.OrgID, OrganizationSlug: namespace,
		ActorID: scope.UserID,
	}); err != nil {
		return control.ResolvedReference{}, err
	}
	return reference, nil
}

func digestFreshReference(payload map[string]any) (string, error) {
	canonical, err := control.CanonicalJSONObject(payload)
	if err != nil {
		return "", err
	}
	return control.DigestCanonicalJSON(canonical)
}

func positiveDigestRevision(value string) int64 {
	digest, err := digestFreshReference(map[string]any{"revision": value})
	if err != nil || len(digest) < 18 {
		return 1
	}
	var out int64 = 1
	for _, char := range digest[len(digest)-16:] {
		out = out*31 + int64(char)
		if out < 0 {
			out = -out
		}
	}
	if out == 0 {
		return 1
	}
	return out
}

func invalidFreshReference(kind, reason string) error {
	return &specservice.InvalidDraftFieldError{Field: kind, Reason: reason}
}
