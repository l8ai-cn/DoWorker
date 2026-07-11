package workercreation

import (
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	runtimedomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

type Deps struct {
	Catalog      runtimedomain.Catalog
	Agents       AgentProvider
	Models       ModelResourceResolver
	Repositories RepositoryLookup
	Skills       SkillLookup
	Knowledge    KnowledgeLookup
	EnvBundles   EnvBundleLookup
}

type Draft struct {
	OptionsRevision string
	WorkerSpec      specservice.Draft
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
