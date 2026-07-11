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
	workerCreation *workercreation.Service
	workerSpecs    specservice.SnapshotRepository
}

func initializeWorkerServices(
	db *gorm.DB,
	agents *agent.AgentService,
	models *airesourceservice.Service,
	repositories *repository.Service,
) workerServices {
	return workerServices{
		workerCreation: initializeWorkerCreationService(db, agents, models, repositories),
		workerSpecs:    infra.NewWorkerSpecSnapshotRepository(db),
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
