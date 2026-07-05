package agent

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

// AgentConfigProvider provides agent lookups for ConfigBuilder. Credential
// resolution moved out: EnvBundles are loaded directly through
// ConfigBuilder.envBundleSvc in buildEnvBundleContext, mirroring how MCP
// servers are exposed to AgentFile eval.
type AgentConfigProvider interface {
	GetAgent(ctx context.Context, slug string) (*agent.Agent, error)
}

// ConfigBuildRequest contains all the information needed to build a pod config
type ConfigBuildRequest struct {
	AgentSlug      string
	OrganizationID int64
	UserID         int64

	// RepositoryID is the repository this pod belongs to (for loading installed extensions)
	RepositoryID *int64

	// Repository configuration
	HttpCloneURL string // HTTPS clone URL
	SshCloneURL  string // SSH clone URL
	SourceBranch string // Branch to checkout

	// Git authentication
	// CredentialType determines how to authenticate:
	// - "runner_local": Use Runner's local git config, no credentials needed
	// - "oauth" or "pat": Use GitToken
	// - "ssh_key": Use SSHPrivateKey
	CredentialType string
	GitToken       string // For oauth/pat types
	SSHPrivateKey  string // For ssh_key type (private key content)

	// Ticket association
	TicketSlug string

	// Preparation script (from AgentFile SETUP or Repository fallback)
	PreparationScript  string
	PreparationTimeout int

	// Local path mode (resume from existing sandbox)
	LocalPath string

	// Prompt (from AgentFile PROMPT declaration)
	Prompt string

	// Runtime info (provided by Runner during handshake)
	MCPPort int
	PodKey  string

	// Terminal size (from browser)
	Cols int32
	Rows int32

	// RunnerAgentVersions maps agent slug to version string.
	// Populated from Runner.AgentVersions during pod creation.
	// Empty map or nil means Runner did not report version info (old Runner).
	RunnerAgentVersions map[string]string

	// MergedAgentfileSource is the merged AgentFile source (base + user layer, serialized).
	// Populated by orchestrator's extractFromAgentfileLayer when AgentfileLayer is provided.
	// When empty (resume mode or no layer): buildFromAgentfile falls back to agent's base AgentFile.
	MergedAgentfileSource string

	// KnowledgeMounts are pre-resolved KB mounts (orchestrator merges agent
	// defaults + Agentfile KNOWLEDGE + request selections and issues tokens).
	KnowledgeMounts []*runnerv1.KnowledgeMount
}

// ConfigSchemaResponse is the config schema returned to frontend.
// CredentialFields are derived from AgentFile ENV SECRET/TEXT declarations;
// frontend merges them with per-agent UX overrides (oneof groups, i18n labels).
type ConfigSchemaResponse struct {
	Fields            []ConfigFieldResponse     `json:"fields"`
	CredentialFields  []CredentialFieldResponse `json:"credential_fields,omitempty"`
	ConfigFiles       []ConfigFileResponse      `json:"config_files,omitempty"`
}

// CredentialFieldResponse describes one user-editable credential ENV key.
type CredentialFieldResponse struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Optional bool   `json:"optional,omitempty"`
}

// ConfigFileResponse describes a JSON config file the agent reads at runtime.
type ConfigFileResponse struct {
	ID       string `json:"id"`
	PathEnv  string `json:"path_env,omitempty"`
	Format   string `json:"format"`
	PathHint string `json:"path_hint,omitempty"`
}

// ConfigFieldResponse is a config field returned to frontend
type ConfigFieldResponse struct {
	Name    string                `json:"name"`
	Type    string                `json:"type"`
	Default interface{}           `json:"default,omitempty"`
	Options []FieldOptionResponse `json:"options,omitempty"`
}

// FieldOptionResponse is a field option returned to frontend
type FieldOptionResponse struct {
	Value string `json:"value"`
}
