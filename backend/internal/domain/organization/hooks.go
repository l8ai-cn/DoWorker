package organization

import "github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"

func (o *Organization) ValidateIdentifiers() error {
	return slugkit.ValidateIdentifier("organizations.slug", o.Slug)
}
