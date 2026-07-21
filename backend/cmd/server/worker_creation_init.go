package main

import (
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	workerruntime "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerruntime"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agent"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	airesourceservice "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/repository"
	workercreation "github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"gorm.io/gorm"
)

type workerServices struct {
	workerCreation       *workercreation.Service
	workerDraftFiller    *workercreation.DraftFiller
	workerDraftGenerator *workercreation.ProviderDraftGenerator
	workerSpecs          specservice.SnapshotRepository
	workerDependencies   agentpod.WorkerSpecDependencyArtifactLoader
	workerDefinitions    *workerdefinition.Catalog
	workerRuntimeCatalog workerruntime.Catalog
}

func initializeWorkerServices(
	cfg *config.Config,
	db *gorm.DB,
	agents *agent.AgentService,
	models *airesourceservice.Service,
	repositories *repository.Service,
	runners workercreation.RunnerAvailabilityResolver,
	providerTokens workercreation.ProviderTokenLookup,
) (workerServices, error) {
	definitions, err := workerdefinition.Load(cfg.WorkerDefinitionsDir)
	if err != nil {
		return workerServices{}, fmt.Errorf("load Worker definition catalog: %w", err)
	}
	catalog, err := workerruntime.LoadCatalog(cfg.WorkerRuntimeCatalogFile)
	if err != nil {
		return workerServices{}, fmt.Errorf("load Worker runtime catalog: %w", err)
	}
	creation := initializeWorkerCreationService(
		db,
		definitions,
		agents,
		models,
		repositories,
		catalog,
		runners,
		workercreation.NewProductionWorkspaceCommitResolver(
			providerTokens,
			newGiteaClientForNamespace(cfg, cfg.KnowledgeBase.GiteaOrg),
			workerRepositoryBaseURLs(cfg)...,
		),
	)
	generator := workercreation.NewProviderDraftGenerator(
		airesourceservice.NewSafeHTTPClient(
			airesourceservice.NewEndpointPolicy(false, nil),
			nil,
		),
	)
	return workerServices{
		workerCreation:       creation,
		workerDraftFiller:    workercreation.NewDraftFiller(creation, models, generator),
		workerDraftGenerator: generator,
		workerSpecs:          infra.NewWorkerSpecSnapshotRepository(db),
		workerDependencies:   infra.NewWorkerSpecDependencyArtifactRepository(db),
		workerDefinitions:    definitions,
		workerRuntimeCatalog: catalog,
	}, nil
}

func workerRepositoryBaseURLs(cfg *config.Config) []string {
	origins := []string{cfg.KnowledgeBase.GiteaURL, cfg.KnowledgeBase.CloneBaseURL}
	return append(origins, cfg.KnowledgeBase.RepositoryBaseURLs...)
}

func initializeWorkerCreationService(
	db *gorm.DB,
	definitions *workerdefinition.Catalog,
	agents *agent.AgentService,
	models *airesourceservice.Service,
	repositories *repository.Service,
	catalog workerruntime.Catalog,
	runners workercreation.RunnerAvailabilityResolver,
	commits workercreation.WorkspaceCommitResolver,
) *workercreation.Service {
	return workercreation.NewService(workercreation.Deps{
		Catalog:      catalog,
		Definitions:  definitions,
		Agents:       agents,
		Models:       models,
		Runners:      runners,
		Repositories: repositories,
		Skills:       infra.NewSkillCatalogRepository(db),
		Knowledge:    infra.NewKnowledgeBaseRepository(db),
		EnvBundles:   infra.NewEnvBundleRepository(db),
		Commits:      commits,
	})
}
