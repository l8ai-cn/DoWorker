package orchestrationresourceconnect

import (
	service "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	resourcev1 "github.com/anthropics/agentsmesh/proto/gen/go/orchestration_resource/v1"
)

func modelBindingFilterFromProto(
	filter *resourcev1.ModelBindingReferenceFilter,
) (*service.ModelBindingReferenceFilter, error) {
	if filter == nil || len(filter.GetProtocolAdapters()) != 0 {
		return nil, invalidRequest()
	}
	workerType := slugkit.Slug(filter.GetWorkerType())
	if err := slugkit.Validate(workerType.String()); err != nil {
		return nil, invalidRequest()
	}
	return &service.ModelBindingReferenceFilter{WorkerType: workerType}, nil
}

func modelBindingFilterToProto(
	filter *service.ModelBindingReferenceFilter,
) *resourcev1.ModelBindingReferenceFilter {
	if filter == nil {
		return nil
	}
	return &resourcev1.ModelBindingReferenceFilter{
		WorkerType:       filter.WorkerType.String(),
		ProtocolAdapters: append([]string{}, filter.ProtocolAdapters...),
	}
}
