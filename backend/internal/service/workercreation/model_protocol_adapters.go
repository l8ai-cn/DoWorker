package workercreation

import "github.com/anthropics/agentsmesh/backend/pkg/slugkit"

func modelProtocolAdapters(adapters []slugkit.Slug) []string {
	values := make([]string, len(adapters))
	for index, adapter := range adapters {
		values[index] = adapter.String()
	}
	return values
}
