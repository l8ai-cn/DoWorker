package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	"github.com/anthropics/agentsmesh/backend/internal/service/relay"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"gorm.io/gorm"
)

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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var r struct {
		OrganizationID int64 `gorm:"column:organization_id"`
	}
	if err := db.WithContext(ctx).Table("runners").Where("id = ?", runnerID).First(&r).Error; err != nil {
		slog.Warn("connect_tunnel: failed to load runner org",
			"runner_id", runnerID, "error", err)
		return
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
		slog.Warn("connect_tunnel: failed to generate tunnel token",
			"runner_id", runnerID, "error", err)
		return
	}

	if err := commandSender.SendConnectTunnel(ctx, runnerID, cfg.TunnelURL(), token); err != nil {
		slog.Warn("connect_tunnel: runner did not confirm readiness",
			"runner_id", runnerID, "error", err)
		return
	}

	slog.Info("Runner tunnel ready", "runner_id", runnerID, "gateway_url", cfg.TunnelURL())
}
