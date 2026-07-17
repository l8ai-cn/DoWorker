package runner

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"github.com/anthropics/agentsmesh/runner/internal/relay"
)

// OnSubscribePod handles subscribe PTY command from server.
// The channel is identified by PodKey (not session ID).
// If already connected to the same Relay URL, just update the token without reconnecting.
// This allows multiple clients (Web + Mobile) to share the same connection.
//
// Lock strategy: relayMu is held ONLY for the pointer check/swap to avoid
// blocking on network I/O or cross-module locks (vt.mu via GetSnapshot).
func (h *RunnerMessageHandler) OnSubscribePod(req client.SubscribePodRequest) error {
	relayURL := h.runner.GetConfig().RewriteRelayURL(req.RelayURL)
	req.RelayURL = relayURL

	_, err, _ := h.relaySubscriptions.Do(req.PodKey+"\x00"+relayURL, func() (any, error) {
		return nil, h.subscribePod(req)
	})
	if err != nil {
		return err
	}
	return h.refreshRelayToken(req.PodKey, relayURL, req.RunnerToken)
}

func (h *RunnerMessageHandler) subscribePod(req client.SubscribePodRequest) error {
	log := logger.Pod()
	relayURL := req.RelayURL

	log.Info("Subscribing to pod via Relay",
		"pod_key", req.PodKey,
		"relay_url", relayURL)

	pod, ok := h.podStore.Get(req.PodKey)
	if !ok {
		return fmt.Errorf("pod not found: %s", req.PodKey)
	}

	// Reject subscribe for pods in terminal states. Initializing pods are
	// allowed — backend may send subscribe_pod before the process starts.
	if status := pod.GetStatus(); status == PodStatusStopped || status == PodStatusFailed {
		log.Info("Pod is not active, rejecting subscribe",
			"pod_key", req.PodKey, "status", status)
		return fmt.Errorf("pod is not active: %s", status)
	}

	log.Debug("Pod interaction mode", "pod_key", req.PodKey, "mode", pod.InteractionMode)

	// Phase 1: Under lock — check existing client and extract/clear if needed.
	// Keep lock scope minimal to avoid blocking on network I/O or cross-module locks.
	var oldClient relay.RelayClient
	pod.LockRelay()
	existingClient := pod.RelayClient
	if existingClient != nil {
		if existingClient.IsConnected() && existingClient.GetRelayURL() == relayURL {
			log.Info("Already connected to same relay",
				"pod_key", req.PodKey,
				"relay_url", relayURL)
			pod.UnlockRelay()
			return nil
		}
		// Connected to different Relay or disconnected, need to reconnect
		log.Info("Disconnecting existing relay connection",
			"pod_key", req.PodKey,
			"old_relay_url", existingClient.GetRelayURL(),
			"new_relay_url", relayURL,
			"was_connected", existingClient.IsConnected())
		pod.RelayClient = nil
		oldClient = existingClient
	}
	pod.UnlockRelay()

	// Stop old client outside the lock — it has its own internal state
	if oldClient != nil {
		oldClient.Stop()
	}

	// Phase 2: Outside lock — network I/O (Connect, Start) cannot deadlock.
	relayClient := h.relayClientFactory(
		relayURL,
		req.PodKey,
		req.RunnerToken,
		slog.Default().With("pod_key", req.PodKey),
	)

	h.setupRelayClientHandlers(relayClient, pod, req)

	if err := relayClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to relay: %w", err)
	}

	if !relayClient.Start() {
		relayClient.Stop()
		return fmt.Errorf("failed to start relay client: client already stopped")
	}

	// Phase 3: Under lock — swap the pointer atomically.
	// Check for a race: another goroutine may have set a different client while we were connecting.
	pod.LockRelay()
	if pod.RelayClient != nil {
		winner := pod.RelayClient
		pod.UnlockRelay()
		relayClient.Stop()
		if winner.IsConnected() && winner.GetRelayURL() == relayURL {
			return nil
		}
		return fmt.Errorf("relay subscription superseded by another connection")
	}
	pod.RelayClient = relayClient
	pod.UnlockRelay()

	// Phase 4: Outside lock — set up relay output and send snapshot.
	// These operations may acquire other locks (vt.mu) but relayMu is NOT held.
	if pod.Relay != nil {
		pod.Relay.OnRelayConnected(relayClient)
		pod.Relay.SendSnapshot(relayClient)
	}

	log.Info("Successfully subscribed to pod via Relay", "pod_key", req.PodKey, "mode", pod.InteractionMode)
	return nil
}

// setupRelayClientHandlers sets up all handlers for a relay client.
// Mode-specific behavior is delegated to PodRelay; shared handlers are wired directly.
func (h *RunnerMessageHandler) setupRelayClientHandlers(relayClient relay.RelayClient, pod *Pod, req client.SubscribePodRequest) {
	log := logger.Pod()
	podKey := req.PodKey

	// Mode-specific handlers — delegated to PodRelay
	if pod.Relay != nil {
		pod.Relay.SetupHandlers(relayClient)
	}

	// Shared: CloseHandler
	relayClient.SetCloseHandler(func() {
		log.Info("Relay connection closed permanently", "pod_key", podKey)
		if pod.GetRelayClient() == relayClient {
			pod.SetRelayClient(nil)
			if pod.Relay != nil {
				pod.Relay.OnRelayDisconnected()
			}
		} else {
			log.Debug("Relay close handler skipped: client already replaced", "pod_key", podKey)
		}
	})

	// Shared: TokenExpiredHandler
	relayClient.SetTokenExpiredHandler(func() string {
		log.Info("Relay token expired, requesting new token", "pod_key", podKey)
		if err := h.conn.SendRequestRelayToken(podKey, relayClient.GetRelayURL()); err != nil {
			log.Error("Failed to send token refresh request", "pod_key", podKey, "error", err)
			return ""
		}
		newToken := pod.WaitForNewToken(30 * time.Second)
		if newToken == "" {
			log.Warn("Timeout waiting for new token", "pod_key", podKey)
		}
		return newToken
	})

	// Shared: ReconnectHandler
	relayClient.SetReconnectHandler(func() {
		log.Info("Relay reconnected, sending snapshot", "pod_key", podKey)
		if pod.Relay != nil {
			pod.Relay.SendSnapshot(relayClient)
		}
	})
}

// OnUnsubscribePod handles unsubscribe PTY command from server.
func (h *RunnerMessageHandler) OnUnsubscribePod(req client.UnsubscribePodRequest) error {
	log := logger.Pod()
	log.Info("Unsubscribing from terminal relay", "pod_key", req.PodKey)

	pod, ok := h.podStore.Get(req.PodKey)
	if !ok {
		return nil
	}

	pod.DisconnectRelay()
	log.Info("Successfully unsubscribed from terminal relay", "pod_key", req.PodKey)
	return nil
}
