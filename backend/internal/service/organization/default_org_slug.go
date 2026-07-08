package organization

import orgDomain "github.com/anthropics/agentsmesh/backend/internal/domain/organization"

const internalAdminWorkspaceSlug = "admin-workspace"

// PickDefaultOrgSlug chooses the org context for API calls when the client has
// no stored preference. admin-workspace is the system-admin personal sandbox;
// deprioritize it when the user belongs to other orgs (e.g. dev-org with Codex runners).
func PickDefaultOrgSlug(orgs []*orgDomain.Organization) string {
	if len(orgs) == 0 {
		return ""
	}
	if len(orgs) == 1 {
		return orgs[0].Slug
	}
	for _, o := range orgs {
		if o.Slug != internalAdminWorkspaceSlug {
			return o.Slug
		}
	}
	return orgs[0].Slug
}

func CollectOrgSlugs(orgs []*orgDomain.Organization) []string {
	slugs := make([]string, len(orgs))
	for i, o := range orgs {
		slugs[i] = o.Slug
	}
	return slugs
}
