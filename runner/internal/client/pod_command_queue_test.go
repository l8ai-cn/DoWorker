package client

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodCommandQueue_SequentialPerPod(t *testing.T) {
	q := NewPodCommandQueue()
	var mu sync.Mutex
	order := make([]int, 0, 10)

	for i := 0; i < 10; i++ {
		index := i
		require.NoError(t, q.Enqueue("pod-1", func() {
			mu.Lock()
			order = append(order, index)
			mu.Unlock()
		}))
	}

	q.Wait()
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, order)
}

func TestPodCommandQueue_ConcurrentAcrossPods(t *testing.T) {
	q := NewPodCommandQueue()
	var concurrent atomic.Int32
	var maximum atomic.Int32
	var commands sync.WaitGroup

	for i := 0; i < 5; i++ {
		commands.Add(1)
		require.NoError(t, q.Enqueue(fmt.Sprintf("pod-%d", i), func() {
			defer commands.Done()
			current := concurrent.Add(1)
			for {
				previous := maximum.Load()
				if current <= previous || maximum.CompareAndSwap(previous, current) {
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
			concurrent.Add(-1)
		}))
	}

	commands.Wait()
	q.Wait()
	assert.GreaterOrEqual(t, maximum.Load(), int32(2))
}

func TestPodCommandQueue_SaturationReturnsImmediately(t *testing.T) {
	q := NewPodCommandQueue()
	started := make(chan struct{})
	release := make(chan struct{})
	require.NoError(t, q.Enqueue("pod-1", func() {
		close(started)
		<-release
	}))
	<-started

	for i := 0; i < podQueueSize; i++ {
		require.NoError(t, q.Enqueue("pod-1", func() {}))
	}

	result := make(chan error, 1)
	go func() {
		result <- q.Enqueue("pod-1", func() {})
	}()
	select {
	case err := <-result:
		require.ErrorIs(t, err, ErrPodCommandQueueFull)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Enqueue blocked on a saturated pod queue")
	}

	close(release)
	q.Wait()
}

func TestPodCommandQueue_AutoCleanupAndReuse(t *testing.T) {
	q := NewPodCommandQueue()
	for i := 0; i < 100; i++ {
		done := make(chan struct{})
		require.NoError(t, q.Enqueue("pod-1", func() { close(done) }))
		<-done
	}

	q.Wait()
	assertPodCommandQueueIdle(t, q)

	done := make(chan struct{})
	require.NoError(t, q.Enqueue("pod-1", func() { close(done) }))
	<-done
	q.Wait()
	assertPodCommandQueueIdle(t, q)
}

func TestPodCommandQueue_ConcurrentCleanup(t *testing.T) {
	q := NewPodCommandQueue()
	const pods = 100
	var commands sync.WaitGroup
	commands.Add(pods)

	for i := 0; i < pods; i++ {
		podKey := fmt.Sprintf("pod-%d", i)
		require.NoError(t, q.Enqueue(podKey, commands.Done))
	}

	commands.Wait()
	q.Wait()
	assertPodCommandQueueIdle(t, q)
}

func TestPodCommandQueue_PanicRecovery(t *testing.T) {
	q := NewPodCommandQueue()
	done := make(chan struct{})
	require.NoError(t, q.Enqueue("pod-1", func() { panic("simulated crash") }))
	require.NoError(t, q.Enqueue("pod-1", func() { close(done) }))

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("command after panic did not execute")
	}
	q.Wait()
}

func assertPodCommandQueueIdle(t *testing.T, q *PodCommandQueue) {
	t.Helper()
	q.mu.Lock()
	defer q.mu.Unlock()
	require.Empty(t, q.queues)
	require.Zero(t, q.workers)
}
