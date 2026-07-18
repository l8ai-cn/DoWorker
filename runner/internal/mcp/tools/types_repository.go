package tools

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
