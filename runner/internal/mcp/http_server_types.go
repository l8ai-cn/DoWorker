package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/l8ai-cn/agentcloud/runner/internal/client"
	"github.com/l8ai-cn/agentcloud/runner/internal/mcp/tools"
)

type PodStatusProvider interface {
	GetPodStatus(
		podKey string,
	) (agentStatus string, podStatus string, shellPid int, found bool)
}

type LocalPodProvider interface {
	GetPodSnapshot(podKey string, lines int) (string, error)
	SendPodInput(podKey string, text string, keys []string) error
}

type WorkbenchArtifactPublisher interface {
	PublishWorkbenchArtifact(
		ctx context.Context,
		podKey string,
		executionID string,
		declaration json.RawMessage,
	) (interface{}, error)
}

type HTTPServer struct {
	rpcClient         *client.RPCClient
	port              int
	pods              map[string]*PodInfo
	mu                sync.RWMutex
	httpServer        *http.Server
	tools             []*MCPTool
	statusProvider    PodStatusProvider
	podProvider       LocalPodProvider
	artifactPublisher WorkbenchArtifactPublisher
}

type PodInfo struct {
	PodKey       string
	OrgSlug      string
	TicketID     *int
	ProjectID    *int
	Agent        string
	RegisteredAt time.Time
	Client       tools.CollaborationClient
}

type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     MCPToolHandler
	PodHandler  MCPPodToolHandler
}

type MCPToolHandler func(
	ctx context.Context,
	client tools.CollaborationClient,
	args map[string]interface{},
) (interface{}, error)

type MCPPodToolHandler func(
	ctx context.Context,
	pod *PodInfo,
	args map[string]interface{},
) (interface{}, error)
