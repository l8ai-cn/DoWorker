package agentpod

import (
	"strings"

	resourcesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
)

func opencodeModelSelector(resource *resourcesvc.ResolvedResource) string {
	model := strings.TrimSpace(resource.Resource.ModelID)
	if strings.Contains(model, "/") {
		return model
	}
	return resource.Connection.ProviderKey.String() + "/" + model
}
