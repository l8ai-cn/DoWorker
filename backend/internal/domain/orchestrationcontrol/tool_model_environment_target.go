package orchestrationcontrol

func isToolModelEnvironmentTargetObject(key string, value map[string]any) bool {
	if key != "environment" || len(value) != 3 {
		return false
	}
	expectedTargets := map[string]string{
		"api_key":  "SEEDANCE_API_KEY",
		"base_url": "SEEDANCE_BASE_URL",
		"model_id": "SEEDANCE_MODEL",
	}
	for field, expected := range expectedTargets {
		target, ok := value[field].(string)
		if !ok || target != expected {
			return false
		}
	}
	return true
}
