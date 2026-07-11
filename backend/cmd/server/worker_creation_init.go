package main

import (
	workerruntime "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	airesourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/service/repository"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"gorm.io/gorm"
)

type workerServices struct {
	workerCreation    *workercreation.Service
	workerDraftFiller *workercreation.DraftFiller
	workerSpecs       specservice.SnapshotRepository
}

func initializeWorkerServices(
	db *gorm.DB,
	agents *agent.AgentService,
	models *airesourceservice.Service,
	repositories *repository.Service,
) workerServices {
	creation := initializeWorkerCreationService(db, agents, models, repositories)
	generator := workercreation.NewProviderDraftGenerator(
		airesourceservice.NewSafeHTTPClient(
			airesourceservice.NewEndpointPolicy(false, nil),
			nil,
		),
	)
	return workerServices{
		workerCreation:    creation,
		workerDraftFiller: workercreation.NewDraftFiller(creation, models, generator),
		workerSpecs:       infra.NewWorkerSpecSnapshotRepository(db),
	}
}

func initializeWorkerCreationService(
	db *gorm.DB,
	agents *agent.AgentService,
	models *airesourceservice.Service,
	repositories *repository.Service,
) *workercreation.Service {
	return workercreation.NewService(workercreation.Deps{
		Catalog:      workerruntime.DefaultCatalog(),
		Agents:       agents,
		Models:       models,
		Repositories: repositories,
		Skills:       infra.NewSkillCatalogRepository(db),
		Knowledge:    infra.NewKnowledgeBaseRepository(db),
		EnvBundles:   infra.NewEnvBundleRepository(db),
	})
}
