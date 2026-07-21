package apikey

import "github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"

func (k *APIKey) ValidateIdentifiers() error {
	if k.Slug == nil {
		return nil
	}
	return slugkit.ValidateIdentifier("api_keys.slug", *k.Slug)
}
