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
