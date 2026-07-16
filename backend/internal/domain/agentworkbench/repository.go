package agentworkbench

import (
	"context"
	"errors"
)

var (
	ErrInvalidArgument     = errors.New("invalid agent workbench repository argument")
	ErrRevisionConflict    = errors.New("agent workbench revision conflict")
	ErrStreamConflict      = errors.New("agent workbench stream epoch conflict")
	ErrEventConflict       = errors.New("agent workbench event conflict")
	ErrCommandIDConflict   = errors.New("agent workbench command ID conflict")
	ErrReceiptConflict     = errors.New("agent workbench receipt conflict")
	ErrSourceEventConflict = errors.New("agent workbench source event conflict")
)

type AppendRequest struct {
	SessionID        string
	ExpectedRevision uint64
	Sources          []SourceEvent
	Receipts         []CommandReceipt
	Events           []Event
	Projection       SessionState
}

type AppendResult struct {
	Applied bool
}

type PersistenceRepository interface {
	Repository
	EnsureSnapshot(
		ctx context.Context,
		initial SessionState,
	) (*SessionState, error)
}

type Repository interface {
	Append(ctx context.Context, request AppendRequest) (AppendResult, error)
	GetSnapshot(ctx context.Context, sessionID string) (*SessionState, error)
	ListAfter(
		ctx context.Context,
		sessionID string,
		streamEpoch string,
		sequence uint64,
		limit int,
	) ([]Event, error)
	PutCommandReceipt(
		ctx context.Context,
		receipt CommandReceipt,
	) (*CommandReceipt, error)
	GetCommandReceipt(
		ctx context.Context,
		sessionID string,
		commandID string,
	) (*CommandReceipt, error)
}
