package workercreation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/gitprovider"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/knowledgebase"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
)

var gitCommitPattern = regexp.MustCompile(`^([0-9a-f]{40}|[0-9a-f]{64})$`)

type WorkspaceCommitResolver interface {
	ResolveRepositoryCommit(
		context.Context,
		specservice.Scope,
		*gitprovider.Repository,
		string,
	) (string, error)
	ResolveKnowledgeBaseCommit(
		context.Context,
		specservice.Scope,
		*knowledgebase.KnowledgeBase,
		string,
	) (string, error)
}

func validateResolvedCommit(label, value string) (string, error) {
	commit := strings.TrimSpace(value)
	if commit == "" || commit != strings.ToLower(commit) ||
		!gitCommitPattern.MatchString(commit) {
		return "", fmt.Errorf("%s resolver returned invalid immutable commit", label)
	}
	return commit, nil
}
