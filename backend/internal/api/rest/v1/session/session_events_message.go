package sessionapi

import (
	"encoding/json"
	"strings"
)

type messageContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	FileID   string `json:"file_id"`
	Filename string `json:"filename"`
}

func parseMessageContent(data json.RawMessage) (blocks []map[string]any, prompt string) {
	var msg struct {
		Content []messageContentBlock `json:"content"`
	}
	if json.Unmarshal(data, &msg) != nil {
		return nil, ""
	}
	blocks = make([]map[string]any, 0, len(msg.Content))
	var attachMarkers []string
	var textParts []string
	for _, b := range msg.Content {
		switch b.Type {
		case "text", "input_text":
			text := strings.TrimSpace(b.Text)
			if text == "" {
				continue
			}
			blocks = append(blocks, map[string]any{"type": "input_text", "text": text})
			textParts = append(textParts, text)
		case "input_image", "input_file":
			fileID := strings.TrimSpace(b.FileID)
			if fileID == "" {
				continue
			}
			name := strings.TrimSpace(b.Filename)
			if name == "" {
				name = fileID
			}
			blocks = append(blocks, map[string]any{
				"type": b.Type, "file_id": fileID, "filename": name,
			})
			attachMarkers = append(attachMarkers, "[Attached: uploads/"+name+"]")
		}
	}
	prompt = strings.Join(attachMarkers, "\n")
	if len(textParts) > 0 {
		if prompt != "" {
			prompt += "\n\n"
		}
		prompt += strings.Join(textParts, "\n")
	}
	return blocks, prompt
}

func messageHasContent(blocks []map[string]any) bool {
	return len(blocks) > 0
}
