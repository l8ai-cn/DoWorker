package runner

import (
	"fmt"

	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"github.com/anthropics/agentsmesh/runner/internal/tunnel"
)

// tunnelRunner is the minimal surface of the outbound tunnel client the handler
// drives. It is satisfied by *tunnel.Client and faked in tests.
type tunnelRunner interface {
	Connect() error
	Start()
	Stop()
	UpdateToken(string)
	GatewayURL() string
	IsConnected() bool
}

// OnConnectTunnel handles the connect_tunnel command. It establishes (or
// refreshes) the single per-runner outbound HTTP tunnel to the Gateway.
//
// Lock strategy mirrors OnSubscribePod: tunnelMu is held only for the pointer
// check/swap, never across network I/O (Connect/Start).
func (h *RunnerMessageHandler) OnConnectTunnel(req client.ConnectTunnelRequest) error {
	log := logger.Runner()

	gatewayURL := h.runner.GetConfig().RewriteRelayURL(req.GatewayURL)
	if gatewayURL != req.GatewayURL {
		log.Info("Gateway URL rewritten", "original", req.GatewayURL, "rewritten", gatewayURL)
	}

	// Phase 1: under lock — reuse existing tunnel if same gateway, else detach.
	var old tunnelRunner
	h.tunnelMu.Lock()
	existing := h.tunnelClient
	if existing != nil {
		if existing.IsConnected() && existing.GatewayURL() == gatewayURL {
			existing.UpdateToken(req.TunnelToken)
			h.tunnelMu.Unlock()
			log.Debug("Tunnel already connected to same gateway, token updated")
			return nil
		}
		h.tunnelClient = nil
		old = existing
	}
	h.tunnelMu.Unlock()

	if old != nil {
		old.Stop()
	}

	// Phase 2: outside lock — build/connect.
	cl := h.tunnelClientFactory(gatewayURL, req.TunnelToken)
	if err := cl.Connect(); err != nil {
		return fmt.Errorf("failed to connect tunnel: %w", err)
	}
	cl.Start()

	// Phase 3: under lock — swap pointer, guarding against a racing connect.
	h.tunnelMu.Lock()
	if h.tunnelClient != nil {
		h.tunnelMu.Unlock()
		cl.Stop()
		log.Info("Another tunnel client was set while connecting, discarding ours")
		return nil
	}
	h.tunnelClient = cl
	h.tunnelMu.Unlock()

	log.Info("Tunnel connected to gateway", "gateway_url", gatewayURL)
	return nil
}

// defaultTunnelClientFactory builds a real tunnel client. runnerID/orgID are
// left 0 (unknown): the Gateway authoritatively derives them from the verified
// tunnel token, so the HELLO carries no numeric id.
func (h *RunnerMessageHandler) defaultTunnelClientFactory(gatewayURL, token string) tunnelRunner {
	ctx := h.runner.GetRunContext()
	dispatcher := tunnel.NewDispatcher(ctx, 0)
	return tunnel.NewClient(ctx, gatewayURL, token, 0, 0, dispatcher)
}
