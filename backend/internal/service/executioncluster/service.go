package executioncluster

import (
	"context"
	"errors"
	"time"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/executioncluster"
	runnerdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
	runnerservice "github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
)

var ErrClusterNotFound = errors.New("execution cluster not found")

type RegistrationTokenIssuer interface {
	GenerateGRPCRegistrationToken(
		context.Context,
		int64,
		int64,
		*runnerservice.GenerateGRPCRegistrationTokenRequest,
		string,
	) (*runnerservice.GenerateGRPCRegistrationTokenResponse, error)
}

type View struct {
	Cluster              *domain.Cluster
	RunnerCount          int
	OnlineRunnerCount    int
	AvailableRunnerCount int
	TunnelStatus         string
	TunnelLastSeenAt     *time.Time
	TunnelLastError      *string
}

type RegistrationCommand struct {
	Command   string
	ExpiresAt time.Time
}

type Service struct {
	clusters  domain.Repository
	runners   runnerdomain.RunnerRepository
	issuer    RegistrationTokenIssuer
	serverURL string
}

func NewService(
	clusters domain.Repository,
	runners runnerdomain.RunnerRepository,
	issuer RegistrationTokenIssuer,
	serverURL string,
) *Service {
	return &Service{clusters: clusters, runners: runners, issuer: issuer, serverURL: serverURL}
}

func (s *Service) List(ctx context.Context, organizationID int64) ([]View, error) {
	clusters, err := s.clusters.EnsureDefaults(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	runners, err := s.runners.ListForClusterStatus(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	views := make([]View, len(clusters))
	index := make(map[int64]int, len(clusters))
	for i, cluster := range clusters {
		views[i] = View{Cluster: cluster, TunnelStatus: "disconnected"}
		index[cluster.ID] = i
	}
	for _, runner := range runners {
		i, ok := index[runner.ClusterID]
		if !ok {
			continue
		}
		view := &views[i]
		view.RunnerCount++
		if runner.IsOnline() {
			view.OnlineRunnerCount++
		}
		if runner.CanAcceptPod() {
			view.AvailableRunnerCount++
		}
		updateTunnelView(view, runner)
	}
	return views, nil
}

func (s *Service) CreateRegistrationCommand(
	ctx context.Context,
	organizationID, userID, clusterID int64,
	nodeName string,
) (RegistrationCommand, error) {
	cluster, err := s.clusters.GetByIDAndOrganization(ctx, clusterID, organizationID)
	if err != nil {
		return RegistrationCommand{}, err
	}
	if cluster == nil {
		return RegistrationCommand{}, ErrClusterNotFound
	}
	token, err := s.issuer.GenerateGRPCRegistrationToken(
		ctx,
		organizationID,
		userID,
		&runnerservice.GenerateGRPCRegistrationTokenRequest{
			Name:      nodeName,
			ClusterID: cluster.ID,
			SingleUse: true,
			MaxUses:   1,
			ExpiresIn: 900,
		},
		s.serverURL,
	)
	if err != nil {
		return RegistrationCommand{}, err
	}
	return RegistrationCommand{Command: token.Command, ExpiresAt: token.ExpiresAt}, nil
}

func updateTunnelView(view *View, runner *runnerdomain.Runner) {
	if runner.TunnelState == "connected" {
		view.TunnelStatus = "connected"
	}
	if runner.TunnelLastSeenAt != nil &&
		(view.TunnelLastSeenAt == nil || runner.TunnelLastSeenAt.After(*view.TunnelLastSeenAt)) {
		view.TunnelLastSeenAt = runner.TunnelLastSeenAt
		view.TunnelLastError = runner.TunnelLastError
	}
}
