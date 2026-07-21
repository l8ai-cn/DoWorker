package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/relay"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	"gorm.io/gorm"
)

var connectRunnerTunnelRetryDelays = []time.Duration{
	0,
	time.Second,
	2 * time.Second,
	4 * time.Second,
	8 * time.Second,
	15 * time.Second,
	30 * time.Second,
}

// setupTunnelConnectCallback wires the runner "initialized" callback so that the
// backend instructs each runner to establish its outbound HTTP data-plane tunnel
// to the Gateway. The tunnel token is NOT bound to a pod (token_type=tunnel);
// routing later relies purely on the runner_id claim. Readiness confirmation runs
// outside the receive callback because the result arrives on the same gRPC stream.
func setupTunnelConnectCallback(
	cfg *config.Config,
	db *gorm.DB,
	runnerConnMgr *runner.RunnerConnectionManager,
	tokenGenerator *relay.TokenGenerator,
	commandSender runner.RunnerCommandSender,
) {
	origInit := runnerConnMgr.GetInitializedCallback()
	runnerConnMgr.SetInitializedCallback(func(runnerID int64, agents []string) {
		if origInit != nil {
			origInit(runnerID, agents)
		}
		go connectRunnerTunnel(cfg, db, tokenGenerator, commandSender, runnerID)
	})
}

func connectRunnerTunnel(
	cfg *config.Config,
	db *gorm.DB,
	tokenGenerator *relay.TokenGenerator,
	commandSender runner.RunnerCommandSender,
	runnerID int64,
) {
	for attempt, delay := range connectRunnerTunnelRetryDelays {
		if delay > 0 {
			time.Sleep(delay)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err := connectRunnerTunnelOnce(ctx, cfg, db, tokenGenerator, commandSender, runnerID)
		cancel()
		if err == nil {
			slog.Info("Runner tunnel ready", "runner_id", runnerID, "gateway_url", cfg.TunnelURL())
			return
		}
		slog.Warn("connect_tunnel: attempt failed",
			"runner_id", runnerID,
			"attempt", attempt+1,
			"attempts", len(connectRunnerTunnelRetryDelays),
			"error", err,
		)
	}
}

func connectRunnerTunnelOnce(
	ctx context.Context,
	cfg *config.Config,
	db *gorm.DB,
	tokenGenerator *relay.TokenGenerator,
	commandSender runner.RunnerCommandSender,
	runnerID int64,
) error {
	var r struct {
		OrganizationID int64 `gorm:"column:organization_id"`
	}
	if err := db.WithContext(ctx).Table("runners").Where("id = ?", runnerID).First(&r).Error; err != nil {
		return fmt.Errorf("load runner org: %w", err)
	}

	token, err := tokenGenerator.GenerateTypedToken(
		"",
		runnerID,
		0,
		r.OrganizationID,
		"tunnel",
		"",
		24*time.Hour,
	)
	if err != nil {
		return fmt.Errorf("generate tunnel token: %w", err)
	}

	if err := commandSender.SendConnectTunnel(ctx, runnerID, cfg.TunnelURL(), token); err != nil {
		return fmt.Errorf("runner readiness: %w", err)
	}
	return nil
}
