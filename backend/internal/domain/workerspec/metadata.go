package workerspec

type Metadata struct {
	Alias          string `json:"alias"`
	SourceExpertID *int64 `json:"source_expert_id,omitempty"`
}
