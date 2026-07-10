package v1

import "encoding/json"

var legacyPodCreateModelFields = []string{
	"credential" + "_profile_id",
	"model" + "_config_id",
	"virtual_api" + "_key_id",
}

func legacyPodCreateModelField(raw []byte) (string, bool) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", false
	}
	for _, field := range legacyPodCreateModelFields {
		if _, ok := payload[field]; ok {
			return field, true
		}
	}
	return "", false
}
