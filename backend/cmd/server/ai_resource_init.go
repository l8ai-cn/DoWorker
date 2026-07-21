package main

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	airesourceservice "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	organizationservice "github.com/l8ai-cn/agentcloud/backend/internal/service/organization"
	"github.com/l8ai-cn/agentcloud/backend/pkg/crypto"
	"gorm.io/gorm"
)

func initializeAIResourceService(db *gorm.DB, orgs *organizationservice.Service, cipher *crypto.Encryptor) *airesourceservice.Service {
	policy := airesourceservice.NewEndpointPolicy(false, nil)
	prober, err := airesourceservice.NewHTTPConnectionProber(airesourceservice.NewSafeHTTPClient(policy, nil))
	if err != nil {
		panic(err)
	}
	service, err := airesourceservice.NewService(airesourceservice.Dependencies{
		Repository: infra.NewAIResourceRepository(db), Cipher: cipher, Members: orgs, Prober: prober,
		Mutations: infra.NewAIResourceMutationRunner(db), Endpoints: policy,
	})
	if err != nil {
		panic(err)
	}
	return service
}
