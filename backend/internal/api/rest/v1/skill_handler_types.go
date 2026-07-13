package v1

import (
	"context"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	skillSvc "github.com/anthropics/agentsmesh/backend/internal/service/skill"
)

type createSkillRequest struct {
	Slug         string   `json:"slug"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	License      string   `json:"license"`
	Instructions string   `json:"instructions"`
	Tags         []string `json:"tags"`
}

type updateSkillRequest struct {
	Name         *string   `json:"name"`
	Description  *string   `json:"description"`
	License      *string   `json:"license"`
	Instructions *string   `json:"instructions"`
	Tags         *[]string `json:"tags"`
}

type importSkillsRequest struct {
	URL            string   `json:"url" binding:"required"`
	Branch         string   `json:"branch"`
	Subdir         string   `json:"subdir"`
	AgentFilter    []string `json:"agent_filter"`
	AuthType       string   `json:"auth_type"`
	AuthCredential string   `json:"auth_credential"`
}

type skillHandlerService interface {
	List(context.Context, int64, int, int) ([]skilldom.Skill, int64, error)
	Get(context.Context, int64, string) (*skilldom.Skill, error)
	Create(context.Context, *skillSvc.CreateSkillRequest) (*skilldom.Skill, error)
	Update(context.Context, *skillSvc.UpdateSkillRequest) (*skilldom.Skill, error)
	Delete(context.Context, int64, int64) error
	ReadSkillFile(context.Context, int64, string, string) ([]byte, *gitops.Entry, error)
	ListSkillTree(context.Context, int64, string) ([]gitops.Entry, error)
	ImportFromGit(context.Context, *skillSvc.ImportFromGitRequest) ([]*skilldom.Skill, error)
	SyncFromUpstream(context.Context, int64, string) (*skilldom.Skill, error)
}
