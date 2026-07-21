package client

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

const podQueueSize = 16

var ErrPodCommandQueueFull = errors.New("pod command queue full")

type podCommandState struct {
	pending []func()
}

type PodCommandQueue struct {
	mu      sync.Mutex
	idle    *sync.Cond
	queues  map[string]*podCommandState
	workers int
}

func NewPodCommandQueue() *PodCommandQueue {
	q := &PodCommandQueue{queues: make(map[string]*podCommandState)}
	q.idle = sync.NewCond(&q.mu)
	return q
}

func (q *PodCommandQueue) Enqueue(podKey string, fn func()) error {
	q.mu.Lock()
	state := q.queues[podKey]
	if state != nil {
		if len(state.pending) >= podQueueSize {
			q.mu.Unlock()
			return fmt.Errorf("%w: %s", ErrPodCommandQueueFull, podKey)
		}
		state.pending = append(state.pending, fn)
		q.mu.Unlock()
		return nil
	}

	state = &podCommandState{pending: []func(){fn}}
	q.queues[podKey] = state
	q.workers++
	q.mu.Unlock()
	go q.run(podKey, state)
	return nil
}

func (q *PodCommandQueue) run(podKey string, state *podCommandState) {
	for {
		q.mu.Lock()
		if len(state.pending) == 0 {
			if q.queues[podKey] == state {
				delete(q.queues, podKey)
			}
			q.workers--
			if q.workers == 0 {
				q.idle.Broadcast()
			}
			q.mu.Unlock()
			return
		}
		fn := state.pending[0]
		state.pending[0] = nil
		state.pending = state.pending[1:]
		q.mu.Unlock()
		q.safeExec(fn)
	}
}

func (q *PodCommandQueue) safeExec(fn func()) {
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.GRPC().Error("Panic in pod command", "panic", recovered, "stack", string(debug.Stack()))
		}
	}()
	fn()
}

func (q *PodCommandQueue) Wait() {
	q.mu.Lock()
	for q.workers > 0 {
		q.idle.Wait()
	}
	q.mu.Unlock()
}
