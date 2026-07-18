package tools

// ConfigFieldSummary is a simplified config field for LLM consumption.
// Removes validation and show_when fields that are only used by frontend.
type ConfigFieldSummary struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Default  interface{} `json:"default,omitempty"`
	Options  []string    `json:"options,omitempty"`
	Required bool        `json:"required,omitempty"`
}

// AgentSummary is a simplified Agent for LLM consumption.
type AgentSummary struct {
	Slug        string                 `json:"slug"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Config      []ConfigFieldSummary   `json:"config,omitempty"`
	UserConfig  map[string]interface{} `json:"user_config,omitempty"`
}

// RunnerSummary is a simplified Runner with nested Agent details.
// Optimized for LLM token efficiency - removes host_info, timestamps, etc.
type RunnerSummary struct {
	ID                int64          `json:"id"`
	NodeID            string         `json:"node_id"`
	Description       string         `json:"description,omitempty"`
	Status            string         `json:"status"`
	CurrentPods       int            `json:"current_pods"`
	MaxConcurrentPods int            `json:"max_concurrent_pods"`
	AvailableAgents   []AgentSummary `json:"available_agents"`
}

// Repository represents a Git repository configuration.
type Repository struct {
	ID              int64  `json:"id"`
	ProviderType    string `json:"provider_type"`
	ProviderBaseURL string `json:"provider_base_url"`
	HttpCloneURL    string `json:"http_clone_url,omitempty"`
	ExternalID      string `json:"external_id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	DefaultBranch   string `json:"default_branch"`
	TicketPrefix    string `json:"ticket_prefix,omitempty"`
	Visibility      string `json:"visibility"`
	IsActive        bool   `json:"is_active"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}
