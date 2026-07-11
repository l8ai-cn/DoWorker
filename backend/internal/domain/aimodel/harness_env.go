package aimodel

import "strings"

// HarnessEnvVars maps a resolved pool model into process env for a harness.
// Used at Worker create time (ephemeral USE_ENV_BUNDLE), not persisted rows.
func HarnessEnvVars(agentSlug, overrideModel string, m *AIModel, credentials map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	switch agentSlug {
	case "codex-cli":
		return openAIHarnessEnv(m, credentials, overrideModel)
	case "claude-code":
		return anthropicHarnessEnv(m, credentials)
	case "gemini-cli":
		return geminiHarnessEnv(m, credentials)
	case "openclaw", "harn":
		return multiProviderHarnessEnv(m, credentials, overrideModel)
	default:
		return nil
	}
}

// PreferredProvider returns the model-pool provider type a harness expects
// when auto-selecting a default legacy pool model.
func PreferredProvider(agentSlug string) string {
	switch agentSlug {
	case "codex-cli":
		return ProviderTypeOpenAI
	case "claude-code":
		return ProviderTypeAnthropic
	case "gemini-cli":
		return ProviderTypeGemini
	default:
		return ""
	}
}

// PreferredProviders returns the ordered provider types a harness prefers when
// auto-selecting a default model, tried before the plain org/user default.
//
// do-agent is provider-agnostic but its settings.json only speaks the
// chat/completions- or anthropic-wire; a bare "openai" org default may be a
// Responses-only gateway (e.g. codex) it cannot call. It therefore prefers
// anthropic-compatible providers (anthropic/minimax both map to kind=anthropic)
// and falls back to the org default when none exist.
func PreferredProviders(agentSlug string) []string {
	if agentSlug == "do-agent" {
		return []string{ProviderTypeAnthropic, ProviderTypeMiniMax}
	}
	if agentSlug == "openclaw" || agentSlug == "harn" {
		return []string{ProviderTypeOpenAI, ProviderTypeAnthropic, ProviderTypeGemini}
	}
	if p := PreferredProvider(agentSlug); p != "" {
		return []string{p}
	}
	return nil
}

func openAIHarnessEnv(m *AIModel, credentials map[string]string, overrideModel string) map[string]string {
	if m.ProviderType != ProviderTypeOpenAI {
		return nil
	}
	apiKey := credentialKey(credentials, "api_key", "auth_token")
	if apiKey == "" {
		return nil
	}
	out := map[string]string{"OPENAI_API_KEY": apiKey}
	if base := strings.TrimSpace(m.BaseURL); base != "" {
		out["OPENAI_BASE_URL"] = base
	}
	model := strings.TrimSpace(overrideModel)
	if model == "" {
		model = strings.TrimSpace(m.Model)
	}
	if model != "" {
		out["OPENAI_MODEL"] = model
	}
	return out
}

func anthropicHarnessEnv(m *AIModel, credentials map[string]string) map[string]string {
	if m.ProviderType != ProviderTypeAnthropic {
		return nil
	}
	apiKey := credentialKey(credentials, "api_key")
	authToken := credentialKey(credentials, "auth_token")
	if apiKey == "" && authToken == "" {
		return nil
	}
	out := map[string]string{}
	if apiKey != "" {
		out["ANTHROPIC_API_KEY"] = apiKey
	}
	if authToken != "" {
		out["ANTHROPIC_AUTH_TOKEN"] = authToken
	}
	if base := strings.TrimSpace(m.BaseURL); base != "" {
		out["ANTHROPIC_BASE_URL"] = base
	}
	return out
}

func geminiHarnessEnv(m *AIModel, credentials map[string]string) map[string]string {
	if m.ProviderType != ProviderTypeGemini {
		return nil
	}
	apiKey := credentialKey(credentials, "api_key")
	if apiKey == "" {
		return nil
	}
	return map[string]string{"GOOGLE_API_KEY": apiKey}
}

func multiProviderHarnessEnv(m *AIModel, credentials map[string]string, overrideModel string) map[string]string {
	switch m.ProviderType {
	case ProviderTypeOpenAI:
		return openAIHarnessEnv(m, credentials, overrideModel)
	case ProviderTypeAnthropic:
		env := anthropicHarnessEnv(m, credentials)
		if env == nil {
			return nil
		}
		if model := harnessModel(m, overrideModel); model != "" {
			env["ANTHROPIC_MODEL"] = model
		}
		return env
	case ProviderTypeGemini:
		apiKey := credentialKey(credentials, "api_key")
		if apiKey == "" {
			return nil
		}
		out := map[string]string{
			"GOOGLE_API_KEY": apiKey,
			"GEMINI_API_KEY": apiKey,
		}
		if model := harnessModel(m, overrideModel); model != "" {
			out["GEMINI_MODEL"] = model
		}
		return out
	default:
		return nil
	}
}

func harnessModel(m *AIModel, overrideModel string) string {
	model := strings.TrimSpace(overrideModel)
	if model == "" {
		model = strings.TrimSpace(m.Model)
	}
	return model
}

func credentialKey(credentials map[string]string, keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(credentials[k]); v != "" {
			return v
		}
	}
	return ""
}
