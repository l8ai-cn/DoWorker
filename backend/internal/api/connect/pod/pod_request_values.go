package podconnect

import (
	"errors"
	"strings"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
)

func normalizeAlias(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func validateAlias(value *string) error {
	if value != nil && len(*value) > 100 {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("alias must be 100 characters or less"))
	}
	return nil
}

func optionalString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func optionalInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func optionalBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func knowledgeMountsFromProto(mounts []*podv1.PodKnowledgeMount) []agentpod.KnowledgeMountRequest {
	if len(mounts) == 0 {
		return nil
	}
	requests := make([]agentpod.KnowledgeMountRequest, 0, len(mounts))
	for _, mount := range mounts {
		requests = append(requests, agentpod.KnowledgeMountRequest{
			Slug: mount.GetSlug(),
			Mode: mount.GetMode(),
		})
	}
	return requests
}
