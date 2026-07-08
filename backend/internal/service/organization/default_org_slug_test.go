package organization

import (
	"testing"

	orgDomain "github.com/anthropics/agentsmesh/backend/internal/domain/organization"
)

func TestPickDefaultOrgSlug(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := PickDefaultOrgSlug(nil); got != "" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("single org", func(t *testing.T) {
		orgs := []*orgDomain.Organization{{Slug: "admin-workspace"}}
		if got := PickDefaultOrgSlug(orgs); got != "admin-workspace" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("deprioritizes admin-workspace", func(t *testing.T) {
		orgs := []*orgDomain.Organization{
			{Slug: "admin-workspace"},
			{Slug: "dev-org"},
		}
		if got := PickDefaultOrgSlug(orgs); got != "dev-org" {
			t.Fatalf("got %q", got)
		}
	})
}
