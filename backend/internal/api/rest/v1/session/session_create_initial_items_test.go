package sessionapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitialItemsContainAttachments(t *testing.T) {
	items := []json.RawMessage{json.RawMessage(`{
		"type":"message",
		"data":{
			"role":"user",
			"content":[
				{"type":"input_text","text":"Inspect this"},
				{"type":"input_file","file_id":"file_123","filename":"notes.pdf"}
			]
		}
	}`)}

	require.True(t, initialItemsContainAttachments(items))
}
