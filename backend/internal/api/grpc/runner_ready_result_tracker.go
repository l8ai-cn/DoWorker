package grpc

import "sync"

type runnerReadyResultKind uint8

const (
	relaySubscriptionReady runnerReadyResultKind = iota + 1
	tunnelConnectionReady
)

type runnerReadyResultKey struct {
	runnerID  int64
	commandID string
	kind      runnerReadyResultKind
}

type runnerReadyResult struct {
	success   bool
	errorCode string
	message   string
}

type pendingRunnerReadyResult struct {
	generation int64
	result     chan runnerReadyResult
}

type runnerReadyResultTracker struct {
	mu      sync.Mutex
	pending map[runnerReadyResultKey]*pendingRunnerReadyResult
}

func newRunnerReadyResultTracker() *runnerReadyResultTracker {
	return &runnerReadyResultTracker{
		pending: make(map[runnerReadyResultKey]*pendingRunnerReadyResult),
	}
}

func (t *runnerReadyResultTracker) register(
	runnerID int64,
	generation int64,
	commandID string,
	kind runnerReadyResultKind,
) (<-chan runnerReadyResult, func()) {
	key := runnerReadyResultKey{runnerID: runnerID, commandID: commandID, kind: kind}
	pending := &pendingRunnerReadyResult{
		generation: generation,
		result:     make(chan runnerReadyResult, 1),
	}

	t.mu.Lock()
	t.pending[key] = pending
	t.mu.Unlock()

	return pending.result, func() {
		t.mu.Lock()
		if t.pending[key] == pending {
			delete(t.pending, key)
		}
		t.mu.Unlock()
	}
}

func (t *runnerReadyResultTracker) complete(
	runnerID int64,
	generation int64,
	commandID string,
	kind runnerReadyResultKind,
	result runnerReadyResult,
) bool {
	key := runnerReadyResultKey{runnerID: runnerID, commandID: commandID, kind: kind}

	t.mu.Lock()
	pending := t.pending[key]
	if pending == nil || pending.generation != generation {
		t.mu.Unlock()
		return false
	}
	delete(t.pending, key)
	t.mu.Unlock()

	pending.result <- result
	return true
}
