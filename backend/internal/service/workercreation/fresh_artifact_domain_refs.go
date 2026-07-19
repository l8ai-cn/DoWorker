package workercreation

import (
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func freshNamedDomainReference(
	scope specservice.Scope,
	namespace slugkit.Slug,
	kind, name string,
	domainID int64,
	payload map[string]any,
) (control.ResolvedReference, error) {
	slug, err := slugkit.NewFromTrusted(name)
	if err != nil {
		return control.ResolvedReference{}, err
	}
	payload["domain_id"] = domainID
	return freshResolvedReference(
		scope,
		namespace,
		kind,
		slug,
		positiveDigestRevision(fmt.Sprintf("%s:%d", kind, domainID)),
		payload,
	)
}

func freshDomainReference(
	scope specservice.Scope,
	namespace slugkit.Slug,
	kind string,
	prefix string,
	domainID int64,
	payload map[string]any,
) (control.ResolvedReference, error) {
	name, err := slugkit.NewFromTrusted(fmt.Sprintf("%s-%d", prefix, domainID))
	if err != nil {
		return control.ResolvedReference{}, err
	}
	payload["domain_id"] = domainID
	return freshResolvedReference(
		scope,
		namespace,
		kind,
		name,
		positiveDigestRevision(fmt.Sprintf("%s:%d", kind, domainID)),
		payload,
	)
}
