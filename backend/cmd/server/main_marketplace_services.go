package main

import (
	"log/slog"

	v1 "github.com/anthropics/agentsmesh/backend/internal/api/rest/v1"
	"github.com/anthropics/agentsmesh/backend/internal/config"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	skillsvc "github.com/anthropics/agentsmesh/backend/internal/service/skill"
	"gorm.io/gorm"
)

func wireExpertAndSkillServices(
	cfg *config.Config,
	db *gorm.DB,
	services *serviceContainer,
	apiServices *v1.Services,
	podOrchestrator *agentpod.PodOrchestrator,
	logger *slog.Logger,
) {
	expertService, skillService := newExpertAndSkillServices(
		cfg,
		db,
		services,
		podOrchestrator,
		logger,
	)
	apiServices.Expert = expertService
	apiServices.Skill = skillService
	apiServices.WorkerSpecs = services.workerSpecs
}

func newExpertAndSkillServices(
	cfg *config.Config,
	db *gorm.DB,
	services *serviceContainer,
	podOrchestrator *agentpod.PodOrchestrator,
	logger *slog.Logger,
) (*expertsvc.Service, *skillsvc.Service) {
	expertGitops := gitops.NewService(
		newGiteaClientForNamespace(cfg, "am-experts"),
		logger,
	)
	skillCatalog := infra.NewSkillCatalogRepository(db)
	expertService := expertsvc.NewService(expertsvc.Deps{
		Store:             infra.NewExpertRepository(db),
		Pods:              services.pod,
		Dispatch:          podOrchestrator,
		Repos:             services.repository,
		WorkerSpecs:       services.workerSpecs,
		WorkerSpecWriter:  services.workerSpecs,
		MarketWorkerSpecs: services.workerCreation,
		MarketInstallLock: infra.NewExpertMarketInstallationLocker(db),
		Market:            infra.NewExpertMarketRepository(db),
		Skills:            skillCatalog,
		MarketSkills:      skillCatalog,
		Gitops:            expertGitops,
		Logger:            logger,
	})

	var packager skillsvc.SkillPackagerBridge
	if services.extension != nil {
		packager = services.extension.SkillPackager()
	}
	skillService := skillsvc.NewService(skillsvc.Deps{
		Store: skillCatalog,
		Gitops: gitops.NewService(
			newGiteaClientForNamespace(cfg, "am-skills"),
			logger,
		),
		Packager: packager,
		Logger:   logger,
	})
	return expertService, skillService
}
