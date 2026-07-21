package channel

import "github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"

func (c *Channel) ValidateIdentifiers() error {
	if c.Slug == nil {
		return nil
	}
	return slugkit.ValidateIdentifier("channels.slug", *c.Slug)
}
