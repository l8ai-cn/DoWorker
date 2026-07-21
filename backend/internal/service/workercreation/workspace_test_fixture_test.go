package workercreation

import specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"

func validWorkspaceDraft() specdomain.Workspace {
	repositoryID := int64(22)
	return specdomain.Workspace{
		RepositoryID: &repositoryID,
		Branch:       "main",
		SkillIDs:     []int64{3},
		KnowledgeMounts: []specdomain.KnowledgeMount{
			{KnowledgeBaseID: 4, Mode: specdomain.KnowledgeMountReadWrite},
		},
		EnvBundleIDs: []specdomain.RuntimeEnvBundleID{5},
		Instructions: "Review before editing.",
		InitialTask:  "Fix the failing test.",
	}
}
