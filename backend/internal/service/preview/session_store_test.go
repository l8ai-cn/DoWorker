package preview

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestSessionStoreRedeemsBootstrapAtomically(t *testing.T) {
	store := newSessionStore(t)
	record := SessionRecord{
		ID:        "session-1",
		PodKey:    "pod-1",
		UserID:    42,
		OrgID:     3,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	results := make(chan error, 16)
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- store.Redeem(context.Background(), "bootstrap-1", record, 5*time.Minute)
		}()
	}
	wg.Wait()
	close(results)

	successes := 0
	replays := 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrBootstrapConsumed):
			replays++
		default:
			t.Fatalf("unexpected redeem error: %v", err)
		}
	}
	if successes != 1 || replays != 15 {
		t.Fatalf("successes=%d replays=%d, want 1 and 15", successes, replays)
	}
	if _, err := store.Get(context.Background(), record.ID); err != nil {
		t.Fatalf("session record missing after redemption: %v", err)
	}
}

func TestSessionStoreRevokesAllSessionsForUser(t *testing.T) {
	store := newSessionStore(t)
	for _, record := range []SessionRecord{
		{ID: "session-1", PodKey: "pod-1", UserID: 42, OrgID: 3, ExpiresAt: time.Now().Add(time.Minute)},
		{ID: "session-2", PodKey: "pod-2", UserID: 42, OrgID: 3, ExpiresAt: time.Now().Add(time.Minute)},
		{ID: "session-3", PodKey: "pod-3", UserID: 99, OrgID: 3, ExpiresAt: time.Now().Add(time.Minute)},
	} {
		if err := store.Redeem(context.Background(), "bootstrap-"+record.ID, record, time.Minute); err != nil {
			t.Fatal(err)
		}
	}

	if err := store.RevokeUser(context.Background(), 42); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(context.Background(), "session-1"); !errors.Is(err, ErrSessionInactive) {
		t.Fatalf("session-1 error = %v, want inactive", err)
	}
	if _, err := store.Get(context.Background(), "session-2"); !errors.Is(err, ErrSessionInactive) {
		t.Fatalf("session-2 error = %v, want inactive", err)
	}
	if _, err := store.Get(context.Background(), "session-3"); err != nil {
		t.Fatalf("other user's session revoked: %v", err)
	}
}

func TestSessionStoreKeepsUserIndexUntilLongestSessionExpires(t *testing.T) {
	store := newSessionStore(t)
	longSession := SessionRecord{
		ID:        "session-long",
		PodKey:    "pod-1",
		UserID:    42,
		OrgID:     3,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	shortSession := SessionRecord{
		ID:        "session-short",
		PodKey:    "pod-2",
		UserID:    42,
		OrgID:     3,
		ExpiresAt: time.Now().Add(time.Minute),
	}
	if err := store.Redeem(context.Background(), "bootstrap-long", longSession, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := store.Redeem(context.Background(), "bootstrap-short", shortSession, time.Minute); err != nil {
		t.Fatal(err)
	}

	ttl, err := store.redis.PTTL(context.Background(), userKey(42)).Result()
	if err != nil {
		t.Fatal(err)
	}
	if ttl < 14*time.Minute {
		t.Fatalf("user session index TTL = %v, want longest session TTL", ttl)
	}
}

func newSessionStore(t *testing.T) *SessionStore {
	t.Helper()
	mr := miniredis.RunT(t)
	host, portText, err := net.SplitHostPort(mr.Addr())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(&redis.Options{Addr: net.JoinHostPort(host, strconv.Itoa(port))})
	t.Cleanup(func() { _ = client.Close() })
	return NewSessionStore(client)
}
