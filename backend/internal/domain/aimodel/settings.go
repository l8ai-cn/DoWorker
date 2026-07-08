package aimodel

// DoAgentSettings builds the do-agent settings.json document for a resolved
// model. do-agent expects `provider.<name>.options.{apiKey,baseURL,kind}` plus
// a top-level `model` of `<provider>/<model-id>` (see ~/.agent/settings.json).
func DoAgentSettings(providerType, providerModel, baseURL string, credentials map[string]string) map[string]interface{} {
	providerName := providerType
	apiKey := credentials["api_key"]
	if apiKey == "" {
		apiKey = credentials["auth_token"]
	}

	options := map[string]interface{}{}
	if apiKey != "" {
		options["apiKey"] = apiKey
	}
	if baseURL != "" {
		options["baseURL"] = baseURL
	}
	if kind := doAgentProviderKind(providerType); kind != "" {
		options["kind"] = kind
	}

	model := providerModel
	if model != "" && providerName != "" && !hasProviderPrefix(model) {
		model = providerName + "/" + model
	}

	settings := map[string]interface{}{}
	if len(options) > 0 {
		settings["provider"] = map[string]interface{}{
			providerName: map[string]interface{}{"options": options},
		}
	}
	if model != "" {
		settings["model"] = model
	}
	return settings
}

func doAgentProviderKind(providerType string) string {
	switch providerType {
	case "minimax", "anthropic":
		return "anthropic"
	case "openai":
		return "openai"
	default:
		return ""
	}
}

func hasProviderPrefix(model string) bool {
	for i := 0; i < len(model); i++ {
		if model[i] == '/' {
			return true
		}
	}
	return false
}
