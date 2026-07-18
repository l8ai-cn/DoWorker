package tools

type PodCreateRequest struct {
	PlanID string `json:"plan_id"`
}

type PodCreateResponse struct {
	PodKey      string `json:"pod_key"`
	Status      string `json:"status"`
	TerminalURL string `json:"terminal_url,omitempty"`
}
