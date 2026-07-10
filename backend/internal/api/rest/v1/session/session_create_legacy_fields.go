package sessionapi

import "encoding/json"

var legacySessionCreateModelFields = []string{
	"credential" + "_profile_id",
	"model",
	"model" + "_config_id",
	"virtual_api" + "_key_id",
}

func legacySessionCreateModelField(raw []byte) (string, bool) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", false
	}
	for _, field := range legacySessionCreateModelFields {
		if _, ok := payload[field]; ok {
			return field, true
		}
	}
	return "", false
}
