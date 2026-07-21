package agentpod

import (
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func artifactKnowledgeMounts(
	document *workerdependency.Document,
) ([]*runnerv1.KnowledgeMount, error) {
	if document == nil || len(document.KnowledgeBases) == 0 {
		return nil, nil
	}
	out := make([]*runnerv1.KnowledgeMount, 0, len(document.KnowledgeBases))
	for _, kb := range document.KnowledgeBases {
		if strings.TrimSpace(kb.HTTPCloneURL) == "" ||
			strings.TrimSpace(kb.CommitSHA) == "" {
			return nil, ErrWorkerSpecDependencyUnavailable
		}
		out = append(out, &runnerv1.KnowledgeMount{
			Slug:         kb.Slug.String(),
			HttpCloneUrl: kb.HTTPCloneURL,
			Branch:       kb.Branch,
			CommitSha:    kb.CommitSHA,
			MountPath:    "kb/" + kb.Slug.String(),
			Mode:         string(kb.Mode),
		})
	}
	return out, nil
}
