package codex

func headlessAutomationFields(cwd string) map[string]any {
	params := map[string]any{
		"approvalPolicy": "never",
		"sandbox":        "danger-full-access",
	}
	if cwd != "" {
		params["cwd"] = cwd
	}
	return params
}

func mergeHeadlessFields(params map[string]any, cwd string) map[string]any {
	if params == nil {
		params = map[string]any{}
	}
	for k, v := range headlessAutomationFields(cwd) {
		params[k] = v
	}
	return params
}
