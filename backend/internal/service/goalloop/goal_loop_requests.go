package goalloop

type CreateRequest struct {
	OrganizationID       int64
	CreatedByID          int64
	Name                 string
	Slug                 string
	Description          *string
	WorkerSpecSnapshotID int64
	Objective            string
	AcceptanceCriteria   []string
	VerificationCommand  string
	MaxIterations        int
	TokenBudget          *int64
	TimeoutMinutes       int
	NoProgressLimit      int
	SameErrorLimit       int
	EscalationPolicy     string
}
