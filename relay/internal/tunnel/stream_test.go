package tunnel

import (
	"context"
	"testing"
	"time"
)

func TestCreditWindow_BlocksAndResumes(t *testing.T) {
	w := newCreditWindow(4)
	if err := w.acquire(context.Background(), 4); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- w.acquire(context.Background(), 3) }() // 窗口耗尽，阻塞
	select {
	case <-done:
		t.Fatal("acquire should block when window empty")
	case <-time.After(50 * time.Millisecond):
	}
	w.add(5) // 补窗
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("acquire should resume after credit added")
	}
}

func TestCreditWindow_CtxCancel(t *testing.T) {
	w := newCreditWindow(0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := w.acquire(ctx, 1); err == nil {
		t.Fatal("expected ctx error")
	}
}
