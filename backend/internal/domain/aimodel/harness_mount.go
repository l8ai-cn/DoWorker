package aimodel

// HarnessMountKind is how a resolved ai_models row is injected at session create.
type HarnessMountKind int

const (
	HarnessMountNone HarnessMountKind = iota
	// HarnessMountConfig uses ephemeral USE_CONFIG_BUNDLE (do-agent settings.json).
	HarnessMountConfig
	// HarnessMountEnv uses ephemeral USE_ENV_BUNDLE (OPENAI_* / ANTHROPIC_* env).
	HarnessMountEnv
)

// HarnessMountKindFor picks config vs env injection for agentSlug.
// executableIsDoAgent is true when the agent's executable is do-agent.
func HarnessMountKindFor(agentSlug string, executableIsDoAgent bool) HarnessMountKind {
	if executableIsDoAgent {
		return HarnessMountConfig
	}
	if PreferredProvider(agentSlug) != "" {
		return HarnessMountEnv
	}
	return HarnessMountNone
}
