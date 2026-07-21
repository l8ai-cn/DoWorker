package envbundle

import (
	"encoding/json"
	"fmt"
	"strings"

	envbundledomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
)

func validateBundleData(kind string, data map[string]string) error {
	if kind != envbundledomain.KindConfig {
		return nil
	}
	_, err := decodeConfigDocument(data)
	return err
}

func decodeConfigDocument(data map[string]string) (map[string]interface{}, error) {
	raw, ok := data[envbundledomain.ConfigJSONDataKey]
	if !ok || strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("%w: %q is required", ErrInvalidConfig, envbundledomain.ConfigJSONDataKey)
	}
	var document map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &document); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}
	if document == nil {
		return nil, fmt.Errorf("%w: JSON object is required", ErrInvalidConfig)
	}
	return document, nil
}
