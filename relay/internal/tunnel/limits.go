package tunnel

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrTargetBusy is returned when a pod's preview concurrency+queue capacity is
// exhausted, mapping to HTTP 429 at the edge.
var ErrTargetBusy = errors.New("target_busy")

// PodLimiter bounds concurrent preview requests per pod. Each pod gets a
// buffered slot channel (capacity = maxConcurrent). When full, up to maxQueue
// additional requests may wait for queueTimeout before being rejected with
// ErrTargetBusy. Preview requests to the same pod run concurrently (no opSlot
// serialization) up to the concurrency limit.
type PodLimiter struct {
	maxConcurrent int
	maxQueue      int
	queueTimeout  time.Duration

	mu    sync.Mutex
	slots map[string]*podSlot
}

type podSlot struct {
	ch       chan struct{}
	queueLen int
}

// NewPodLimiter creates a limiter. maxConcurrent<=0 is treated as 1.
func NewPodLimiter(maxConcurrent, maxQueue int, queueTimeout time.Duration) *PodLimiter {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	if maxQueue < 0 {
		maxQueue = 0
	}
	return &PodLimiter{
		maxConcurrent: maxConcurrent,
		maxQueue:      maxQueue,
		queueTimeout:  queueTimeout,
		slots:         make(map[string]*podSlot),
	}
}

func (l *PodLimiter) slotFor(podKey string) *podSlot {
	l.mu.Lock()
	defer l.mu.Unlock()
	s := l.slots[podKey]
	if s == nil {
		s = &podSlot{ch: make(chan struct{}, l.maxConcurrent)}
		l.slots[podKey] = s
	}
	return s
}

// Acquire reserves a concurrency slot for podKey. It returns a release function
// that must be called exactly once. If the slot is full it queues (bounded by
// maxQueue) up to queueTimeout, otherwise returns ErrTargetBusy.
func (l *PodLimiter) Acquire(ctx context.Context, podKey string) (func(), error) {
	s := l.slotFor(podKey)

	// Fast path: try to grab a slot without blocking.
	select {
	case s.ch <- struct{}{}:
		return l.releaseFunc(s), nil
	default:
	}

	// Slot full: decide whether we may queue.
	l.mu.Lock()
	if s.queueLen >= l.maxQueue {
		l.mu.Unlock()
		return nil, ErrTargetBusy
	}
	s.queueLen++
	l.mu.Unlock()

	defer func() {
		l.mu.Lock()
		s.queueLen--
		l.mu.Unlock()
	}()

	var timeout <-chan time.Time
	if l.queueTimeout > 0 {
		t := time.NewTimer(l.queueTimeout)
		defer t.Stop()
		timeout = t.C
	}

	select {
	case s.ch <- struct{}{}:
		return l.releaseFunc(s), nil
	case <-timeout:
		return nil, ErrTargetBusy
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (l *PodLimiter) releaseFunc(s *podSlot) func() {
	var once sync.Once
	return func() {
		once.Do(func() { <-s.ch })
	}
}
