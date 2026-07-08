package tunnel

import (
	"context"
	"testing"
	"time"
)

func TestPodLimiter_AcquireRelease(t *testing.T) {
	l := NewPodLimiter(1, 0, 10*time.Millisecond) // concurrency 1, queue 0
	rel, err := l.Acquire(context.Background(), "pod1")
	if err != nil {
		t.Fatal(err)
	}
	// queue 0, second acquire is immediately busy
	if _, err := l.Acquire(context.Background(), "pod1"); err == nil {
		t.Fatal("expected busy")
	}
	rel()
	if r2, err := l.Acquire(context.Background(), "pod1"); err != nil {
		t.Fatal(err)
	} else {
		r2()
	}
}

func TestPodLimiter_QueueWaitsThenAcquires(t *testing.T) {
	l := NewPodLimiter(1, 1, 200*time.Millisecond) // concurrency 1, queue 1
	rel, err := l.Acquire(context.Background(), "pod1")
	if err != nil {
		t.Fatal(err)
	}
	// A queued acquire should succeed once the first releases.
	done := make(chan error, 1)
	go func() {
		r2, e := l.Acquire(context.Background(), "pod1")
		if e == nil {
			r2()
		}
		done <- e
	}()
	time.Sleep(20 * time.Millisecond)
	rel()
	if err := <-done; err != nil {
		t.Fatalf("queued acquire should succeed: %v", err)
	}
}

func TestPodLimiter_IsolatedPerPod(t *testing.T) {
	l := NewPodLimiter(1, 0, 10*time.Millisecond)
	r1, err := l.Acquire(context.Background(), "pod1")
	if err != nil {
		t.Fatal(err)
	}
	defer r1()
	// Different pod has its own slot.
	r2, err := l.Acquire(context.Background(), "pod2")
	if err != nil {
		t.Fatalf("different pod must not be blocked: %v", err)
	}
	r2()
}
