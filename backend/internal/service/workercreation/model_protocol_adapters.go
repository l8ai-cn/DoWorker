package workercreation

import "github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"

func modelProtocolAdapters(adapters []slugkit.Slug) []string {
	values := make([]string, len(adapters))
	for index, adapter := range adapters {
		values[index] = adapter.String()
	}
	return values
}
