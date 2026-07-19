package sessionapi

import (
	"encoding/json"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile"
)

func promptLayerFromItems(items []json.RawMessage) *string {
	text := promptTextFromInitialItems(items)
	if text == "" {
		return nil
	}
	layer := "PROMPT " + agentfile.FormatStringLiteral(text)
	return &layer
}

func promptTextFromInitialItems(items []json.RawMessage) string {
	for _, raw := range items {
		var evt struct {
			Type string `json:"type"`
			Data struct {
				Role    string `json:"role"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"data"`
		}
		if json.Unmarshal(raw, &evt) != nil || evt.Type != "message" {
			continue
		}
		var parts []string
		for _, block := range evt.Data.Content {
			if (block.Type == "text" || block.Type == "input_text") && block.Text != "" {
				parts = append(parts, block.Text)
			}
		}
		if len(parts) == 0 {
			continue
		}
		return strings.Join(parts, "\n")
	}
	return ""
}
