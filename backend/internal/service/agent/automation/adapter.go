// Package automation translates the unified, cross-agent worker automation
// level (interactive / auto_edit / autonomous) into each agent's native
// permission mechanism. Every agent gets its own adapter file; the registry
// picks one by agent slug with a safe fallback. The output is rendered as
// AgentFile layer text (CONFIG.../MODE...) and appended to the user layer, so
// the whole translation reuses the existing agentfile_layer → launch_args
// pipeline without touching the runner.
package automation

import (
	"fmt"
	"sort"
	"strings"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

// Output is the agent-native translation of one automation level.
type Output struct {
	// InteractionMode forces a MODE line when non-empty (e.g. "acp"); empty
	// leaves the user/base MODE untouched.
	InteractionMode string
	// ConfigOverrides are agentfile CONFIG key→value pairs to inject.
	ConfigOverrides map[string]string
}

// Adapter translates a normalized automation level into one agent's native
// AgentFile CONFIG/MODE overrides.
type Adapter interface {
	Apply(level string) Output
}

var registry = map[string]Adapter{}

func register(slug string, a Adapter) { registry[slug] = a }

// AdapterFor returns the adapter registered for the agent slug, or the default
// fallback (autonomous ⇒ MODE acp, no CONFIG assumptions).
func AdapterFor(agentSlug string) Adapter {
	if a, ok := registry[agentSlug]; ok {
		return a
	}
	return defaultAdapter{}
}

// LayerLinesFor renders the adapter output for (agentSlug, level) as AgentFile
// layer text. The level is normalized first, so empty/unknown ⇒ autonomous.
// canForceMode gates the MODE line: when the resolved agent cannot run ACP
// (pty-only), forcing it would fail mode validation, so we degrade gracefully
// and only apply the CONFIG overrides. Returns "" when nothing to inject.
func LayerLinesFor(agentSlug, level string, canForceMode bool) string {
	out := AdapterFor(agentSlug).Apply(podDomain.NormalizeAutomationLevel(level))
	if !canForceMode {
		out.InteractionMode = ""
	}
	return renderLayer(out)
}

func renderLayer(out Output) string {
	lines := make([]string, 0, len(out.ConfigOverrides)+1)
	keys := make([]string, 0, len(out.ConfigOverrides))
	for k := range out.ConfigOverrides {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("CONFIG %s = %q", k, out.ConfigOverrides[k]))
	}
	if out.InteractionMode != "" {
		lines = append(lines, "MODE "+out.InteractionMode)
	}
	return strings.Join(lines, "\n")
}

// defaultAdapter is the fallback for agents without a fine-grained permission
// knob (gemini / do-agent / aider / opencode / cursor / …). It still enforces
// the automation guarantee: autonomous runs non-interactively via ACP.
type defaultAdapter struct{}

func (defaultAdapter) Apply(level string) Output {
	out := Output{}
	if level == podDomain.AutomationLevelAutonomous {
		out.InteractionMode = podDomain.InteractionModeACP
	}
	return out
}
