package v1

type createSkillRequest struct {
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	License      string `json:"license"`
	Instructions string `json:"instructions"`
}

type updateSkillRequest struct {
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	License      *string `json:"license"`
	Instructions *string `json:"instructions"`
}

type importSkillsRequest struct {
	URL            string   `json:"url" binding:"required"`
	Branch         string   `json:"branch"`
	Subdir         string   `json:"subdir"`
	AgentFilter    []string `json:"agent_filter"`
	AuthType       string   `json:"auth_type"`
	AuthCredential string   `json:"auth_credential"`
}
