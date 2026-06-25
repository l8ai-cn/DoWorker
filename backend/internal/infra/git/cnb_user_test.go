package git

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCNBGetCurrentUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q, want Bearer test-token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": 123,
			"username": "alice",
			"nickname": "Alice",
			"email": "alice@example.com",
			"avatar_url": "https://cnb.cool/avatar.png"
		}`))
	}))
	defer server.Close()

	provider, err := NewCNBProvider(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewCNBProvider: %v", err)
	}
	user, err := provider.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("GetCurrentUser: %v", err)
	}
	if user.ID != "123" || user.Username != "alice" || user.Name != "Alice" {
		t.Fatalf("unexpected user: %+v", user)
	}
}
