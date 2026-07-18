package agentworkbench

import "time"

type SessionState struct {
	SessionID      string
	StreamEpoch    string
	Revision       uint64
	LatestSequence uint64
	Projection     []byte
	Digest         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Event struct {
	SessionID          string
	StreamEpoch        string
	Revision           uint64
	Sequence           uint64
	Payload            []byte
	Digest             string
	CausationCommandID *string
	CreatedAt          time.Time
}

type SourceEvent struct {
	SessionID          string
	StableEventID      string
	RunnerSessionEpoch string
	SourceSequence     uint64
	PayloadDigest      string
}

type ReceiptState int16

const (
	ReceiptStateReceived  ReceiptState = 1
	ReceiptStateAccepted  ReceiptState = 2
	ReceiptStateRunning   ReceiptState = 3
	ReceiptStateSucceeded ReceiptState = 4
	ReceiptStateFailed    ReceiptState = 5
	ReceiptStateRejected  ReceiptState = 6
	ReceiptStateCancelled ReceiptState = 7
)

type CommandReceipt struct {
	SessionID     string
	CommandID     string
	PayloadDigest string
	State         ReceiptState
	Receipt       []byte
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
