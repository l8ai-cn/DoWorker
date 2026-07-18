package codex

import (
	"encoding/json"
	"fmt"
	"time"
)

const modelListPageSize = 100

func (t *transport) loadModels() error {
	var cursor *string
	var models []string

	for {
		page, err := t.listModels(cursor)
		if err != nil {
			return err
		}
		for _, model := range page.Data {
			if model.Hidden || model.Model == "" {
				continue
			}
			models = append(models, model.Model)
		}
		if page.NextCursor == nil {
			break
		}
		cursor = page.NextCursor
	}
	if len(models) == 0 {
		return fmt.Errorf("model/list returned no visible models")
	}

	t.modelMu.Lock()
	t.models = models
	t.modelMu.Unlock()
	return nil
}

func (t *transport) listModels(cursor *string) (modelListResponse, error) {
	request, err := t.tracker.SendRequest("model/list", modelListParams{
		Cursor: cursor,
		Limit:  modelListPageSize,
	})
	if err != nil {
		return modelListResponse{}, fmt.Errorf("write model/list: %w", err)
	}
	response, err := t.tracker.WaitResponse(request, 30*time.Second)
	if err != nil {
		return modelListResponse{}, fmt.Errorf("wait model/list response: %w", err)
	}
	if response.Error != nil {
		return modelListResponse{}, fmt.Errorf(
			"model/list error: code=%d msg=%s",
			response.Error.Code,
			response.Error.Message,
		)
	}
	var result modelListResponse
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return modelListResponse{}, fmt.Errorf("parse model/list result: %w", err)
	}
	return result, nil
}

func (t *transport) SupportedModels() []string {
	t.modelMu.RLock()
	defer t.modelMu.RUnlock()
	return append([]string(nil), t.models...)
}

func (t *transport) CurrentModel() string {
	t.modelMu.RLock()
	defer t.modelMu.RUnlock()
	return t.model
}

func (t *transport) validateModel(model string) error {
	if model == "" {
		return fmt.Errorf("model required")
	}
	t.modelMu.RLock()
	defer t.modelMu.RUnlock()
	for _, supported := range t.models {
		if supported == model {
			return nil
		}
	}
	return fmt.Errorf("unsupported model %q", model)
}

func (t *transport) setCurrentModel(model string) error {
	if model == "" {
		return fmt.Errorf("thread response missing model")
	}
	t.modelMu.Lock()
	t.model = model
	t.modelMu.Unlock()
	return nil
}

func (t *transport) updateThreadModel(threadID, model string) error {
	if threadID == "" {
		return fmt.Errorf("thread id required")
	}
	if err := t.validateModel(model); err != nil {
		return err
	}
	request, err := t.tracker.SendRequest(
		"thread/settings/update",
		threadSettingsUpdateParams{ThreadID: threadID, Model: model},
	)
	if err != nil {
		return fmt.Errorf("write thread/settings/update: %w", err)
	}
	response, err := t.tracker.WaitResponse(request, 30*time.Second)
	if err != nil {
		return fmt.Errorf("wait thread/settings/update response: %w", err)
	}
	if response.Error != nil {
		return fmt.Errorf(
			"thread/settings/update error: code=%d msg=%s",
			response.Error.Code,
			response.Error.Message,
		)
	}
	return t.setCurrentModel(model)
}
