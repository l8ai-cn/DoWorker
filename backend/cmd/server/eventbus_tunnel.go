package main

import (
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
// routing later relies purely on the runner_id claim. Failures are logged as
// warnings only: runners reconnect and re-request on their own, so a transient
// send failure must not block initialization.
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

		var r struct {
			OrganizationID int64 `gorm:"column:organization_id"`
		}
		if err := db.Table("runners").Where("id = ?", runnerID).First(&r).Error; err != nil {
			slog.Warn("connect_tunnel: failed to load runner org, skipping (will retry on reconnect)",
				"runner_id", runnerID, "error", err)
			return
		}

		token, err := tokenGenerator.GenerateTypedToken(
			"",             // podKey: tunnel tokens are not pod-bound
			runnerID,       // runner_id claim drives tunnel routing
			0,              // userID=0 for runner-class tokens
			r.OrganizationID,
			"tunnel",       // token_type
			"",             // previewTarget: not applicable for tunnel tokens
			24*time.Hour,
		)
		if err != nil {
			slog.Warn("connect_tunnel: failed to generate tunnel token (will retry on reconnect)",
				"runner_id", runnerID, "error", err)
			return
		}

		if err := commandSender.SendConnectTunnel(runnerID, cfg.TunnelURL(), token); err != nil {
			slog.Warn("connect_tunnel: failed to send command (relying on reconnect)",
				"runner_id", runnerID, "error", err)
			return
		}

		slog.Info("Sent connect_tunnel to runner", "runner_id", runnerID, "gateway_url", cfg.TunnelURL())
	})
}
