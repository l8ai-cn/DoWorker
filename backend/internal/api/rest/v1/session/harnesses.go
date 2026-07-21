package sessionapi

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var harnessLabelOverrides = map[string]string{
	"claude":      "Claude SDK",
	"claude-code": "Claude Code",
	"codex":       "Codex",
	"codex-cli":   "Codex",
	"gemini":      "Gemini",
	"gemini-cli":  "Gemini",
	"opencode":    "OpenCode",
	"cursor-cli":  "Cursor",
	"aider":       "Aider",
	"loopal":      "Loopal",
	"do-agent":    "DoAgent",
	"doagent":     "DoAgent",
	"openclaw":    "OpenClaw",
	"hermes":      "Hermes",
}

type harnessWire struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

func (d *Deps) handleListHarnesses(c *gin.Context) {
	if d.Agent == nil {
		c.JSON(http.StatusOK, gin.H{"data": []harnessWire{}})
		return
	}
	builtin, err := d.Agent.ListBuiltinAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list harnesses"})
		return
	}
	includeInternal := os.Getenv("AGENTCLOUD_INCLUDE_INTERNAL_AGENTS") == "true"
	seen := make(map[string]struct{})
	rows := make([]harnessWire, 0, len(builtin))
	for _, a := range builtin {
		if !a.IsActive || (a.IsInternal && !includeInternal) {
			continue
		}
		id := a.Executable
		if id == "" {
			id = a.Slug
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		label := harnessLabelOverrides[id]
		if label == "" {
			label = harnessLabelOverrides[a.Slug]
		}
		if label == "" {
			label = titleCase(a.Name)
			if label == "" {
				label = titleCase(strings.ReplaceAll(id, "-", " "))
			}
		}
		rows = append(rows, harnessWire{ID: id, Label: label})
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func titleCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
