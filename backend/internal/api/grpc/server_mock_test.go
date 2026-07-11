package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
)

// newTestLogger creates a test logger
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// mockRunnerService implements RunnerServiceInterface for testing
type mockRunnerService struct {
	runners               map[string]RunnerInfo
	revokedCerts          map[string]bool
	err                   error
	revocationCheckErr    error // Separate error for IsCertificateRevoked
	markConnectedErr      error
	markDisconnectedErr   error
	mu                    sync.Mutex
	connected             map[int64]bool
	disconnected          map[int64]bool
	markConnectedCalls    int
	markDisconnectedCalls int
	refreshHeartbeatCalls int
	lastHeartbeatPods     int
	updateLastSeenCalls   int
}

func newMockRunnerService() *mockRunnerService {
	return &mockRunnerService{
		runners:      make(map[string]RunnerInfo),
		revokedCerts: make(map[string]bool),
		connected:    make(map[int64]bool),
		disconnected: make(map[int64]bool),
	}
}

func (m *mockRunnerService) GetByNodeID(ctx context.Context, nodeID string) (RunnerInfo, error) {
	if m.err != nil {
		return RunnerInfo{}, m.err
	}
	if runner, ok := m.runners[nodeID]; ok {
		return runner, nil
	}
	return RunnerInfo{}, context.DeadlineExceeded
}

func (m *mockRunnerService) UpdateLastSeen(ctx context.Context, runnerID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateLastSeenCalls++
	return m.err
}

func (m *mockRunnerService) MarkConnected(ctx context.Context, runnerID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.markConnectedCalls++
	m.connected[runnerID] = true
	if m.markConnectedErr != nil {
		return m.markConnectedErr
	}
	return m.err
}

func (m *mockRunnerService) MarkDisconnected(ctx context.Context, runnerID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.markDisconnectedCalls++
	m.disconnected[runnerID] = true
	if m.markDisconnectedErr != nil {
		return m.markDisconnectedErr
	}
	return m.err
}

func (m *mockRunnerService) WasMarkedConnected(runnerID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected[runnerID]
}

func (m *mockRunnerService) WasMarkedDisconnected(runnerID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.disconnected[runnerID]
}

func (m *mockRunnerService) RefreshActiveHeartbeat(runnerID int64, currentPods int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refreshHeartbeatCalls++
	m.lastHeartbeatPods = currentPods
}

func (m *mockRunnerService) UpdateAvailableAgents(ctx context.Context, runnerID int64, agents []string) error {
	return m.err
}

func (m *mockRunnerService) UpdateAgentVersions(ctx context.Context, runnerID int64, versions []runnerDomain.AgentVersion) error {
	return m.err
}

func (m *mockRunnerService) IsCertificateRevoked(ctx context.Context, serialNumber string) (bool, error) {
	// Use separate error for revocation check to allow testing different error scenarios
	if m.revocationCheckErr != nil {
		return false, m.revocationCheckErr
	}
	// Check if serial is in revoked set
	if revoked, ok := m.revokedCerts[serialNumber]; ok {
		return revoked, nil
	}
	return false, nil
}

func (m *mockRunnerService) UpdateRunnerVersionAndHostInfo(ctx context.Context, runnerID int64, version string, hostInfo map[string]interface{}) error {
	return m.err
}

func (m *mockRunnerService) MergeAgentVersions(ctx context.Context, runnerID int64, changes map[string]runnerDomain.AgentVersion) error {
	return m.err
}

func (m *mockRunnerService) GetByNodeIDAndOrgID(ctx context.Context, nodeID string, orgID int64) (RunnerInfo, error) {
	if m.err != nil {
		return RunnerInfo{}, m.err
	}
	key := fmt.Sprintf("%s:%d", nodeID, orgID)
	if runner, ok := m.runners[key]; ok {
		return runner, nil
	}
	return RunnerInfo{}, context.DeadlineExceeded
}

func (m *mockRunnerService) AddRunner(nodeID string, runner RunnerInfo) {
	m.runners[nodeID] = runner
	// Also register with composite key for GetByNodeIDAndOrgID
	key := fmt.Sprintf("%s:%d", nodeID, runner.OrganizationID)
	m.runners[key] = runner
}

func (m *mockRunnerService) SetCertificateRevoked(serialNumber string, revoked bool) {
	if m.revokedCerts == nil {
		m.revokedCerts = make(map[string]bool)
	}
	m.revokedCerts[serialNumber] = revoked
}

func (m *mockRunnerService) SetRevocationCheckError(err error) {
	m.revocationCheckErr = err
}

func (m *mockRunnerService) SetMarkConnectedError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.markConnectedErr = err
}

// mockOrgService implements OrganizationServiceInterface for testing
type mockOrgService struct {
	orgs map[string]OrganizationInfo
	err  error
}

func newMockOrgService() *mockOrgService {
	return &mockOrgService{
		orgs: make(map[string]OrganizationInfo),
	}
}

func (m *mockOrgService) GetBySlug(ctx context.Context, slug string) (OrganizationInfo, error) {
	if m.err != nil {
		return OrganizationInfo{}, m.err
	}
	if org, ok := m.orgs[slug]; ok {
		return org, nil
	}
	return OrganizationInfo{}, context.DeadlineExceeded
}

func (m *mockOrgService) AddOrg(slug string, org OrganizationInfo) {
	m.orgs[slug] = org
}

// NOTE: ConnectionEventHandler interface has been removed in favor of RunnerConnectionManager callbacks
