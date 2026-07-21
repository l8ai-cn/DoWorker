package schema

import (
	"strings"

	"github.com/l8ai-cn/agentcloud/agentfile/parser"
)

// ConfigFile describes a JSON (or other) config file the agent reads at runtime.
type ConfigFile struct {
	ID       string // e.g. "settings" from settings.json
	PathEnv  string // ENV var whose value is the resolved path (DO_AGENT_SETTINGS)
	Format   string // json
	PathHint string // human-readable path fragment from AgentFile source
}

func extractConfigFiles(prog *parser.Program) []ConfigFile {
	var out []ConfigFile
	for _, decl := range prog.Declarations {
		d, ok := decl.(*parser.EnvDecl)
		if !ok || d.ValueExpr == nil {
			continue
		}
		hint := exprPathHint(d.ValueExpr)
		if !strings.Contains(hint, ".json") {
			continue
		}
		id := configFileID(hint)
		if id == "" {
			continue
		}
		out = append(out, ConfigFile{
			ID:       id,
			PathEnv:  d.Name,
			Format:   "json",
			PathHint: hint,
		})
	}
	return out
}

func configFileID(pathHint string) string {
	idx := strings.LastIndex(pathHint, "/")
	base := pathHint
	if idx >= 0 {
		base = pathHint[idx+1:]
	}
	base = strings.TrimSuffix(base, ".json")
	base = strings.Trim(base, `"`)
	if base == "" {
		return ""
	}
	return base
}

func exprPathHint(expr parser.Expr) string {
	switch e := expr.(type) {
	case *parser.StringLit:
		return e.Value
	case *parser.BinaryExpr:
		if e.Op == "+" {
			return exprPathHint(e.Left) + exprPathHint(e.Right)
		}
		return ""
	default:
		return ""
	}
}
