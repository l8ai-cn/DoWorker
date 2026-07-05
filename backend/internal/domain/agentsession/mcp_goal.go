package agentsession

import "encoding/json"

type McpServer struct {
	Name        string   `json:"name"`
	Transport   string   `json:"transport"`
	Description *string  `json:"description,omitempty"`
	URL         *string  `json:"url,omitempty"`
	Command     *string  `json:"command,omitempty"`
	Args        []string `json:"args,omitempty"`
}

type CodexGoal struct {
	ThreadID        string  `json:"thread_id"`
	Objective       string  `json:"objective"`
	Status          string  `json:"status"`
	TokenBudget     *int64  `json:"token_budget"`
	TokensUsed      int64   `json:"tokens_used"`
	TimeUsedSeconds int64   `json:"time_used_seconds"`
	CreatedAt       *int64  `json:"created_at,omitempty"`
	UpdatedAt       *int64  `json:"updated_at,omitempty"`
}

func ParseMcpServers(raw json.RawMessage) []McpServer {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var out []McpServer
	if json.Unmarshal(raw, &out) != nil {
		return nil
	}
	return out
}

func McpServersToInstalled(servers []McpServer) map[string]any {
	out := make(map[string]any, len(servers))
	for _, s := range servers {
		cfg := map[string]any{"type": s.Transport}
		if s.URL != nil && *s.URL != "" {
			cfg["url"] = *s.URL
		}
		if s.Command != nil && *s.Command != "" {
			cfg["command"] = *s.Command
		}
		if len(s.Args) > 0 {
			cfg["args"] = s.Args
		}
		out[s.Name] = cfg
	}
	return out
}
