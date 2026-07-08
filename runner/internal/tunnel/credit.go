package tunnel

import (
	"context"
	"sync"
)

// creditWindow is the runner-side mirror of the gateway flow-control window: a
// sender blocks in acquire until the gateway grants credits (via CREDIT frames)
// through add. This bounds in-flight RESP body memory to the window size.
type creditWindow struct {
	mu     sync.Mutex
	cond   *sync.Cond
	avail  int
	closed bool
}

func newCreditWindow(initial int) *creditWindow {
	w := &creditWindow{avail: initial}
	w.cond = sync.NewCond(&w.mu)
	return w
}

func (w *creditWindow) acquire(ctx context.Context, n int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for w.avail < n && !w.closed {
		if err := ctx.Err(); err != nil {
			return err
		}
		done := make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
				w.cond.Broadcast()
			case <-done:
			}
		}()
		w.cond.Wait()
		close(done)
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if w.closed {
		return context.Canceled
	}
	w.avail -= n
	return nil
}

func (w *creditWindow) add(n int) {
	w.mu.Lock()
	w.avail += n
	w.mu.Unlock()
	w.cond.Broadcast()
}

func (w *creditWindow) close() {
	w.mu.Lock()
	w.closed = true
	w.mu.Unlock()
	w.cond.Broadcast()
}
