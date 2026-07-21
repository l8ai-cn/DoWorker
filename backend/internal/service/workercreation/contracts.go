package workercreation

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/gitprovider"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	runtimedomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerruntime"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdependencyartifact"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type WorkerDefinitionProvider interface {
	Get(string) (workerdefinition.Definition, bool)
}

type RunnerAvailabilityResolver interface {
	HasAvailableRunnerForAgent(context.Context, int64, int64, string) (bool, error)
}

type Deps struct {
	Catalog      runtimedomain.Catalog
	Definitions  WorkerDefinitionProvider
	Agents       AgentProvider
	Models       ModelResourceResolver
	Runners      RunnerAvailabilityResolver
	Repositories RepositoryLookup
	Skills       SkillLookup
	Knowledge    KnowledgeLookup
	EnvBundles   EnvBundleLookup
	Commits      WorkspaceCommitResolver
}

type Draft struct {
	OptionsRevision  string
	OrganizationSlug slugkit.Slug
	WorkerSpec       specservice.Draft
	ArtifactRefs     ArtifactReferences
}

type ArtifactReferences struct {
	PrimaryModel      *control.ResolvedReference
	ToolBindings      map[string]control.ResolvedReference
	ToolModels        map[string]control.ResolvedReference
	Repository        *control.ResolvedReference
	Skills            map[int64]control.ResolvedReference
	KnowledgeBases    map[int64]control.ResolvedReference
	RuntimeBundles    map[int64]control.ResolvedReference
	SecretBundles     map[string]control.ResolvedReference
	ConfigBundles     map[int64]control.ResolvedReference
	ComputeTarget     *control.ResolvedReference
	ResourceProfile   *control.ResolvedReference
	AllPlanReferences []control.ResolvedReference
}

type Issue struct {
	Code     string
	Field    string
	Message  string
	Severity string
}

type Prepared struct {
	Snapshot       specservice.ResolvedSnapshot
	Spec           specdomain.Spec
	AgentfileLayer string
	Repository     *gitprovider.Repository `json:"-"`
	Artifact       *workerdependencyartifact.Artifact
	Dependencies   *workerdependency.Document
}

type PreparedSnapshot struct {
	Spec           specdomain.Spec
	AgentfileLayer string
	Repository     *gitprovider.Repository `json:"-"`
	Dependencies   *workerdependency.Document
}

type PreflightResult struct {
	BlockingErrors  []Issue
	Warnings        []Issue
	Resolved        *Prepared
	OptionsRevision string
}

type FillResult struct {
	Draft  Draft
	Issues []Issue
}
