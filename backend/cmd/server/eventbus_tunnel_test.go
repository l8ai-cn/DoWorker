package main

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/relay"
	runnerservice "github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConnectRunnerTunnelRetriesTransientReadinessFailure(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/tunnel.db"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE runners (id INTEGER PRIMARY KEY, organization_id INTEGER NOT NULL)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO runners (id, organization_id) VALUES (?, ?)`, 7, 3).Error)

	oldDelays := connectRunnerTunnelRetryDelays
	connectRunnerTunnelRetryDelays = []time.Duration{0, time.Millisecond}
	t.Cleanup(func() { connectRunnerTunnelRetryDelays = oldDelays })

	sender := &retryTunnelSender{
		NoOpCommandSender: runnerservice.NewNoOpCommandSender(slog.Default()),
		failures:          1,
	}
	connectRunnerTunnel(
		&config.Config{PrimaryDomain: "localhost:10000"},
		db,
		relay.NewTokenGenerator("secret", "agentcloud-relay"),
		sender,
		7,
	)

	require.Equal(t, 2, sender.calls)
	require.Equal(t, "ws://localhost:10000/runner/tunnel", sender.gatewayURL)
	require.NotEmpty(t, sender.token)
}

type retryTunnelSender struct {
	*runnerservice.NoOpCommandSender
	mu         sync.Mutex
	calls      int
	failures   int
	gatewayURL string
	token      string
}

func (s *retryTunnelSender) SendConnectTunnel(
	_ context.Context,
	_ int64,
	gatewayURL,
	tunnelToken string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.gatewayURL = gatewayURL
	s.token = tunnelToken
	if s.calls <= s.failures {
		return errors.New("websocket: bad handshake")
	}
	return nil
}
