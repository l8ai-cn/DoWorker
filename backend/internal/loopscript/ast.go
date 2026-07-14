package loopscript

type Program struct {
	SchemaVersion int
	Loop          LoopNode
}

type LoopNode struct {
	NodeID        string
	LocalID       string
	Worker        WorkerNode
	Limits        Limits
	Repeat        RepeatNode
	FailurePolicy FailurePolicy
}

type WorkerNode struct {
	NodeID     string
	LocalID    string
	SnapshotID int64
}

type Limits struct {
	Iterations  int64
	Tokens      int64
	TimeoutMins int64
	NoProgress  int64
	SameError   int64
}

type RepeatNode struct {
	NodeID   string
	LocalID  string
	Max      int64
	Until    Reference
	Agent    AgentNode
	Verifier VerifierNode
}

type Reference struct {
	LocalID string
	Field   string
}

type AgentNode struct {
	NodeID  string
	LocalID string
	Using   string
	Prompt  string
}

type VerifierNode struct {
	NodeID  string
	LocalID string
	Command string
	Accept  string
}

type FailurePolicy string

const (
	FailurePause FailurePolicy = "pause"
	FailureFail  FailurePolicy = "fail"
)

type Diagnostic struct {
	Code    string
	Message string
	NodeID  string
	Line    int
	Column  int
}

type GoalLoopLaunchSpec struct {
	Name                string
	Slug                string
	WorkerSnapshotID    int64
	Objective           string
	AcceptanceCriteria  []string
	VerificationCommand string
	MaxIterations       int
	TokenBudget         int64
	TimeoutMinutes      int
	NoProgressLimit     int
	SameErrorLimit      int
	EscalationPolicy    string
}

type sourcePosition struct {
	line   int
	column int
}

type programPositions struct {
	loop     sourcePosition
	worker   sourcePosition
	limits   sourcePosition
	repeat   sourcePosition
	agent    sourcePosition
	verifier sourcePosition
	failure  sourcePosition
}
