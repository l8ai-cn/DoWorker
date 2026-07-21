// Package schema derives UI/API schemas from AgentFile declaration sections.
// AgentFile source is the SSOT for CONFIG fields and user-editable ENV
// credentials; this package is the single extraction path shared by the
// backend GetAgentConfigSchema handler and envbundle non-secret key policy.
package schema

import (
	"fmt"

	"github.com/l8ai-cn/agentcloud/agentfile/extract"
	"github.com/l8ai-cn/agentcloud/agentfile/parser"
)

// AgentSchema is the combined config + credential schema for one agent.
type AgentSchema struct {
	ConfigFields     []ConfigField
	CredentialFields []CredentialField
	ConfigFiles      []ConfigFile
}

// ConfigField mirrors a CONFIG declaration for pod/user settings forms.
type ConfigField struct {
	Name    string
	Type    string
	Default interface{}
	Options []string
}

// CredentialField mirrors a user-provided ENV credential (SECRET/TEXT).
type CredentialField struct {
	Name     string
	Type     string // "secret" or "text"
	Optional bool
}

// FromSource parses an AgentFile and extracts CONFIG + credential ENV schemas.
func FromSource(source string) (*AgentSchema, error) {
	prog, errs := parser.Parse(source)
	if len(errs) > 0 {
		return nil, fmt.Errorf("agentfile parse errors: %v", errs)
	}
	return FromProgram(prog), nil
}

// FromProgram extracts schemas from an already-parsed Program.
func FromProgram(prog *parser.Program) *AgentSchema {
	spec := extract.Extract(prog)
	out := &AgentSchema{
		ConfigFields:     make([]ConfigField, 0, len(spec.Config)),
		CredentialFields: make([]CredentialField, 0, len(spec.Env)),
		ConfigFiles:      extractConfigFiles(prog),
	}
	for _, cfg := range spec.Config {
		out.ConfigFields = append(out.ConfigFields, ConfigField{
			Name:    cfg.Name,
			Type:    cfg.Type,
			Default: cfg.Default,
			Options: append([]string(nil), cfg.Options...),
		})
	}
	for _, env := range spec.Env {
		if !isUserCredentialEnv(env.Source) {
			continue
		}
		out.CredentialFields = append(out.CredentialFields, CredentialField{
			Name:     env.Name,
			Type:     env.Source,
			Optional: env.Optional,
		})
	}
	return out
}

// NonSecretCredentialKeys returns ENV names declared as TEXT credentials.
func (s *AgentSchema) NonSecretCredentialKeys() []string {
	var keys []string
	for _, f := range s.CredentialFields {
		if f.Type == "text" {
			keys = append(keys, f.Name)
		}
	}
	return keys
}

func isUserCredentialEnv(source string) bool {
	return source == "secret" || source == "text"
}
