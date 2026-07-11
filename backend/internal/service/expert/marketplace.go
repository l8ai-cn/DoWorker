package expert

import (
	"context"
	"errors"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
)

var ErrMarketApplicationNotFound = errors.New("market application not found")

type MarketApplication struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Icon        string   `json:"icon"`
	AgentSlug   string   `json:"agent_slug"`
	SkillSlugs  []string `json:"skill_slugs"`
	Tags        []string `json:"tags"`
	Outcomes    []string `json:"outcomes"`
	Version     int      `json:"version"`
	Featured    bool     `json:"featured"`
	Prompt      string   `json:"-"`
}

var marketApplications = []MarketApplication{
	{
		Slug:        "software-delivery-expert",
		Name:        "软件交付专家",
		Summary:     "从需求拆解、隔离开发到测试与合并，完成一条可验证的软件交付链路。",
		Description: "适用于功能开发、缺陷修复和版本交付。专家会建立隔离工作区，执行代码修改与测试，并根据仓库类型完成 GitHub PR 或 GitLab MR 交付。",
		Category:    "研发交付",
		Icon:        "rocket",
		AgentSlug:   "codex-cli",
		SkillSlugs:  []string{"worktree", "e2e", "gh-merge", "merge"},
		Tags:        []string{"开发", "测试", "代码合并"},
		Outcomes:    []string{"隔离完成代码修改", "执行关键路径验证", "提交并推动代码合并"},
		Version:     1,
		Featured:    true,
		Prompt:      "你是软件交付专家。先确认验收目标和仓库状态，再使用隔离工作区完成实现。必须执行与改动风险匹配的测试，审查差异，并通过适用的 GitHub PR 或 GitLab MR 流程完成交付。不得用跳过测试或静默降级掩盖问题。",
	},
	{
		Slug:        "multi-worker-orchestrator",
		Name:        "多 Worker 协作专家",
		Summary:     "把复杂目标拆成可并行任务，调度多个 Worker，并汇总为一个可验收结果。",
		Description: "适用于跨模块开发、批量排查和多项独立任务。专家负责范围划分、依赖识别、任务去重、执行跟踪和最终整合。",
		Category:    "团队协作",
		Icon:        "network",
		AgentSlug:   "codex-cli",
		SkillSlugs:  []string{"worker-create", "am-delegate", "worktree"},
		Tags:        []string{"多智能体", "任务拆解", "并行执行"},
		Outcomes:    []string{"形成清晰任务边界", "并行推进独立工作", "统一审查并汇总结果"},
		Version:     1,
		Featured:    true,
		Prompt:      "你是多 Worker 协作专家。先识别目标、依赖和可并行边界，再创建职责互斥的 Worker。持续检查重复工作、阻塞和验证证据。最终由你完成整合审查，并给出一个一致、可验证的结果。",
	},
	{
		Slug:        "dual-repo-sync-expert",
		Name:        "双仓同步专家",
		Summary:     "安全同步 GitLab 与 GitHub 仓库，识别快进、分叉和需要人工决策的差异。",
		Description: "适用于内部仓与开源仓的双向同步。专家会检查双方提交关系，快进场景直接同步，分叉场景进入 PR 或 MR 流程并验证结果。",
		Category:    "代码管理",
		Icon:        "git-compare",
		AgentSlug:   "codex-cli",
		SkillSlugs:  []string{"gl-gh-sync", "e2e"},
		Tags:        []string{"GitLab", "GitHub", "仓库同步"},
		Outcomes:    []string{"识别双方最新提交", "选择正确同步策略", "验证远端分支可见性"},
		Version:     1,
		Featured:    false,
		Prompt:      "你是双仓同步专家。检查 GitLab 与 GitHub 的远端状态和提交包含关系，只在快进条件成立时直接推送。遇到分叉必须通过 PR 或 MR 处理并说明差异。同步后验证远端分支和目标提交均可见。",
	},
}

func (s *Service) ListMarketApplications() []MarketApplication {
	items := make([]MarketApplication, len(marketApplications))
	copy(items, marketApplications)
	return items
}

func (s *Service) InstallMarketApplication(
	ctx context.Context,
	orgID, userID int64,
	marketSlug string,
) (*expertdom.Expert, bool, error) {
	app, ok := findMarketApplication(marketSlug)
	if !ok {
		return nil, false, ErrMarketApplicationNotFound
	}
	existing, err := s.store.GetBySlug(ctx, orgID, app.Slug)
	if err == nil {
		return existing, true, nil
	}
	if !errors.Is(err, expertdom.ErrNotFound) {
		return nil, false, err
	}
	description := app.Description
	prompt := app.Prompt
	expertType := app.Category
	row, err := s.Create(ctx, &CreateExpertRequest{
		OrganizationID:  orgID,
		UserID:          userID,
		Name:            app.Name,
		Slug:            app.Slug,
		Description:     &description,
		AgentSlug:       app.AgentSlug,
		Prompt:          &prompt,
		InteractionMode: expertdom.InteractionModePTY,
		AutomationLevel: expertdom.AutomationLevelAutonomous,
		SkillSlugs:      app.SkillSlugs,
		ExpertType:      &expertType,
	})
	return row, false, err
}

func findMarketApplication(slug string) (MarketApplication, bool) {
	for _, app := range marketApplications {
		if app.Slug == slug {
			return app, true
		}
	}
	return MarketApplication{}, false
}
