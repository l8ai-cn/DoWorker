package expert

import (
	"encoding/json"
	"strings"
)

type expertMetadata struct {
	Avatar     string `json:"avatar,omitempty"`
	ExpertType string `json:"expertType,omitempty"`
}

func parseExpertMetadata(raw json.RawMessage) expertMetadata {
	var metadata expertMetadata
	if len(raw) == 0 || string(raw) == "null" {
		return metadata
	}
	_ = json.Unmarshal(raw, &metadata)
	return metadata
}

func mergeMetadata(
	raw json.RawMessage,
	avatarPath, expertType *string,
) json.RawMessage {
	values := map[string]any{}
	if len(raw) > 0 && string(raw) != "null" {
		_ = json.Unmarshal(raw, &values)
	}
	if avatarPath != nil {
		if *avatarPath == "" {
			delete(values, "avatar")
		} else {
			values["avatar"] = *avatarPath
		}
	}
	if expertType != nil {
		if strings.TrimSpace(*expertType) == "" {
			delete(values, "expertType")
		} else {
			values["expertType"] = strings.TrimSpace(*expertType)
		}
	}
	encoded, err := json.Marshal(values)
	if err != nil || len(encoded) == 0 {
		return json.RawMessage("{}")
	}
	return encoded
}
