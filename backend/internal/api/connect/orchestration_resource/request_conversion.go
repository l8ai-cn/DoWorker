package orchestrationresourceconnect

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	service "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	resourcev1 "github.com/anthropics/agentsmesh/proto/gen/go/orchestration_resource/v1"
)

const (
	defaultListLimit = 50
	maxListLimit     = 100
)

func sourceFromProto(source *resourcev1.ResourceSource) (service.ResourceSource, error) {
	if source == nil {
		return service.ResourceSource{}, invalidRequest()
	}
	format, err := sourceFormatFromProto(source.GetFormat())
	if err != nil {
		return service.ResourceSource{}, err
	}
	return service.ResourceSource{
		Format:  format,
		Content: append([]byte(nil), source.GetContent()...),
	}, nil
}

func sourceFormatFromProto(format resourcev1.SourceFormat) (service.SourceFormat, error) {
	switch format {
	case resourcev1.SourceFormat_SOURCE_FORMAT_JSON:
		return service.SourceFormatJSON, nil
	case resourcev1.SourceFormat_SOURCE_FORMAT_YAML:
		return service.SourceFormatYAML, nil
	default:
		return "", invalidRequest()
	}
}

func targetFromProto(
	scope control.Scope,
	target *resourcev1.ResourceTarget,
) (control.ResourceTarget, error) {
	if target == nil || target.GetTypeMeta() == nil {
		return control.ResourceTarget{}, invalidRequest()
	}
	converted := control.ResourceTarget{
		TypeMeta: resource.TypeMeta{
			APIVersion: target.GetTypeMeta().GetApiVersion(),
			Kind:       target.GetTypeMeta().GetKind(),
		},
		Namespace: slugkit.Slug(target.GetNamespace()),
		Name:      slugkit.Slug(target.GetName()),
	}
	if err := converted.Validate(scope); err != nil {
		return control.ResourceTarget{}, invalidRequest()
	}
	return converted, nil
}

func listFilterFromProto(
	request *resourcev1.ListResourcesRequest,
) (service.ResourceListFilter, error) {
	filter := service.ResourceListFilter{Limit: defaultListLimit}
	if request.Offset != nil {
		if request.GetOffset() < 0 {
			return service.ResourceListFilter{}, invalidRequest()
		}
		filter.Offset = int(request.GetOffset())
	}
	if request.Limit != nil && request.GetLimit() != 0 {
		if request.GetLimit() < 0 || request.GetLimit() > maxListLimit {
			return service.ResourceListFilter{}, invalidRequest()
		}
		filter.Limit = int(request.GetLimit())
	}
	if request.Kind != nil {
		meta := resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       request.GetKind(),
		}
		if err := meta.Validate(); err != nil {
			return service.ResourceListFilter{}, invalidRequest()
		}
		filter.Kind = request.GetKind()
	}
	return filter, nil
}

func revisionFromProto(revision *int64) (int64, error) {
	if revision == nil {
		return 0, nil
	}
	if *revision < 0 {
		return 0, invalidRequest()
	}
	return *revision, nil
}

func planIDFromProto(value string) (string, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		return "", invalidRequest()
	}
	return value, nil
}

func invalidRequest() error {
	return connect.NewError(
		connect.CodeInvalidArgument,
		errors.New("invalid orchestration resource request"),
	)
}
