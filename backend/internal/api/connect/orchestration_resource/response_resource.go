package orchestrationresourceconnect

import (
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/pkg/protoconv"
	resourcev1 "github.com/anthropics/agentsmesh/proto/gen/go/orchestration_resource/v1"
)

func resourceToProto(head control.ResourceHead) *resourcev1.Resource {
	return &resourcev1.Resource{
		Id:              head.ID,
		Identity:        identityToProto(head.Identity),
		DisplayName:     head.DisplayName,
		Labels:          cloneLabels(head.Labels),
		StatusJson:      append([]byte(nil), head.Status...),
		Revision:        head.Revision,
		Generation:      head.Generation,
		ResourceVersion: head.ResourceVersion,
		CreatedById:     head.CreatedByID,
		UpdatedById:     head.UpdatedByID,
		CreatedAt:       protoconv.RFC3339(head.CreatedAt),
		UpdatedAt:       protoconv.RFC3339(head.UpdatedAt),
	}
}

func targetToProto(target control.ResourceTarget) *resourcev1.ResourceTarget {
	return &resourcev1.ResourceTarget{
		TypeMeta:  typeMetaToProto(target.TypeMeta),
		Namespace: target.Namespace.String(),
		Name:      target.Name.String(),
	}
}

func identityToProto(identity control.ResourceIdentity) *resourcev1.ResourceIdentity {
	return &resourcev1.ResourceIdentity{
		Target: targetToProto(identity.ResourceTarget),
		Uid:    identity.UID,
	}
}

func typeMetaToProto(meta resource.TypeMeta) *resourcev1.TypeMeta {
	return &resourcev1.TypeMeta{
		ApiVersion: meta.APIVersion,
		Kind:       meta.Kind,
	}
}

func cloneLabels(source map[string]string) map[string]string {
	if source == nil {
		return nil
	}
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
