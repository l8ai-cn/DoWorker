package skill

import (
	"encoding/json"
	"fmt"
)

var authoredSkillConfigFields = []string{
	"schema",
	"slug",
	"name",
	"description",
	"license",
	"tags",
}

func mergeAuthoredSkillConfig(existing, rendered []byte) ([]byte, error) {
	var config map[string]json.RawMessage
	if err := json.Unmarshal(existing, &config); err != nil {
		return nil, fmt.Errorf("skill: parse current skill.json: %w", err)
	}
	var replacement map[string]json.RawMessage
	if err := json.Unmarshal(rendered, &replacement); err != nil {
		return nil, fmt.Errorf("skill: parse rendered skill.json: %w", err)
	}
	for _, field := range authoredSkillConfigFields {
		delete(config, field)
	}
	for field, value := range replacement {
		config[field] = value
	}
	return marshalSkillConfig(config)
}

func replaceSkillConfigFields(
	content []byte,
	replacements map[string]any,
) ([]byte, error) {
	var config map[string]json.RawMessage
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("skill: parse skill.json: %w", err)
	}
	for key, value := range replacements {
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("skill: render skill.json field %s: %w", key, err)
		}
		config[key] = raw
	}
	return marshalSkillConfig(config)
}

func marshalSkillConfig(config map[string]json.RawMessage) ([]byte, error) {
	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("skill: render skill.json: %w", err)
	}
	return append(content, '\n'), nil
}
