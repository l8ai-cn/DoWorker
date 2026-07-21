package runner

import (
	"context"
	"sync"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/executioncluster"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/billing"
)

type GrantQuerier interface {
	GetGrantedResourceIDs(ctx context.Context, resourceType string, userID int64, orgID int64) ([]string, error)
}

type Service struct {
	repo           runner.RunnerRepository
	billingService *billing.Service
	clusterRepo    executioncluster.Repository
	grantQuerier   GrantQuerier
	activeMu       sync.Mutex
	activeRunners  sync.Map
}

type ActiveRunner struct {
	Runner   *runner.Runner
	LastPing time.Time
	PodCount int
}

// NewService accepts optional billingService — nil skips quota checks (test use).
func NewService(repo runner.RunnerRepository, billingService ...*billing.Service) *Service {
	s := &Service{
		repo: repo,
	}
	if len(billingService) > 0 {
		s.billingService = billingService[0]
	}
	return s
}

func (s *Service) SetGrantQuerier(gq GrantQuerier) {
	s.grantQuerier = gq
}

func (s *Service) SetExecutionClusterRepository(repo executioncluster.Repository) {
	s.clusterRepo = repo
}
