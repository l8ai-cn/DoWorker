package workerspec

type TerminationPolicy string

const (
	TerminationPolicyManual      TerminationPolicy = "manual"
	TerminationPolicyOnIdle      TerminationPolicy = "idle"
	TerminationPolicyOnCompleted TerminationPolicy = "completed"
)

type Lifecycle struct {
	TerminationPolicy  TerminationPolicy `json:"termination_policy"`
	IdleTimeoutMinutes uint32            `json:"idle_timeout_minutes"`
}
