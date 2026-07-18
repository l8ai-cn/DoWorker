package main

import (
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	workerruntime "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	airesourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/service/repository"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
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

func initializeWorkerCreationService(
	db *gorm.DB,
	definitions *workerdefinition.Catalog,
	agents *agent.AgentService,
	models *airesourceservice.Service,
	repositories *repository.Service,
	catalog workerruntime.Catalog,
	runners workercreation.RunnerAvailabilityResolver,
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
	})
}
