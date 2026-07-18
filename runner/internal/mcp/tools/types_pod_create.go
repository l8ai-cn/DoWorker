package tools

// PodCreateRequest represents a request to create a new pod.
type PodCreateRequest struct {
	PlanID string `json:"plan_id"`
}

// PodCreateResponse represents the response from creating a pod.
type PodCreateResponse struct {
	PodKey      string                  `json:"pod_key"`
	Status      string                  `json:"status"`
	TerminalURL string                  `json:"terminal_url,omitempty"`
	Resource    *AppliedResourceSummary `json:"resource,omitempty"`
}
