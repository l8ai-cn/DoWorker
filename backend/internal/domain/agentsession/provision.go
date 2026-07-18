package agentsession

type ProvisionSpec struct {
	ID              string
	Title           *string
	ParentSessionID *string
	ExpectedPodKey  string
	UpdateExisting  bool
}

type ProvisionReceipt struct {
	Session           *Session
	Created           bool
	PreviousPodKey    string
	PreviousAgentSlug string
}
