package knowledgebaseconnect

import (
	"errors"
	"time"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
	kbservice "github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase"
	kbv1 "github.com/anthropics/agentsmesh/proto/gen/go/knowledgebase/v1"
)

func toProtoKnowledgeBase(kb *knowledgebase.KnowledgeBase) *kbv1.KnowledgeBase {
	out := &kbv1.KnowledgeBase{
		Id:               kb.ID,
		Slug:             kb.Slug,
		Name:             kb.Name,
		Description:      kb.Description,
		GitRepoPath:      kb.GitRepoPath,
		HttpCloneUrl:     kb.HTTPCloneURL,
		DefaultBranch:    kb.DefaultBranch,
		SourceType:       kb.SourceType,
		SourceConfigJson: kbservice.RedactedSourceConfigJSON(kb.SourceConfig),
		SyncStatus:       kb.SyncStatus,
		CreatedByUserId:  kb.CreatedByUserID,
		CreatedAt:        kb.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        kb.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if kb.SyncError != nil {
		out.SyncError = kb.SyncError
	}
	if kb.LastSyncedAt != nil {
		ts := kb.LastSyncedAt.UTC().Format(time.RFC3339)
		out.LastSyncedAt = &ts
	}
	return out
}

func mapKBError(err error) error {
	switch {
	case errors.Is(err, kbservice.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, kbservice.ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, knowledgebase.ErrSlugExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
