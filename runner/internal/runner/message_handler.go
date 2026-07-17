package runner

import (
	"log/slog"
	"sync"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"github.com/anthropics/agentsmesh/runner/internal/relay"
	"golang.org/x/sync/singleflight"
)

// RunnerMessageHandler implements client.MessageHandler interface.
type RunnerMessageHandler struct {
	runner             MessageHandlerContext
	podStore           PodStore
	conn               client.Connection
	relayClientFactory func(url, podKey, token string, logger *slog.Logger) relay.RelayClient
	relaySubscriptions singleflight.Group
	verificationRuns   singleflight.Group
	promptDedup        map[string]*promptDedupRing
	promptDedupMu      sync.Mutex
	verificationMu     sync.Mutex
	verificationCache  map[string]*runnerv1.VerificationResultEvent
	verificationOrder  []string
	receipts           *commandReceiptStore

	// Per-runner outbound HTTP tunnel to the Gateway (see message_handler_tunnel.go).
	tunnelMu            sync.Mutex
	tunnelClient        tunnelRunner
	tunnelClientFactory func(gatewayURL, token string) tunnelRunner
}

// NewRunnerMessageHandler creates a new message handler.
func NewRunnerMessageHandler(runner MessageHandlerContext, store PodStore, conn client.Connection) *RunnerMessageHandler {
	logger.Runner().Debug("Creating message handler")
	h := &RunnerMessageHandler{
		runner:            runner,
		podStore:          store,
		conn:              conn,
		promptDedup:       make(map[string]*promptDedupRing),
		verificationCache: make(map[string]*runnerv1.VerificationResultEvent),
		receipts:          commandReceiptStoreForRunner(runner),
		relayClientFactory: func(url, podKey, token string, logger *slog.Logger) relay.RelayClient {
			return relay.NewClient(runner.GetRunContext(), url, podKey, token, logger)
		},
	}
	h.tunnelClientFactory = h.defaultTunnelClientFactory
	return h
}

var _ client.MessageHandler = (*RunnerMessageHandler)(nil)
