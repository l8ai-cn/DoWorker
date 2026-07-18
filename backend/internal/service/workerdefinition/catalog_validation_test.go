package workerdefinition

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateConfigDocumentSchemaRequiresRuntimeContract(t *testing.T) {
	err := validateConfigDocumentSchema(json.RawMessage(`{
		"type":"array",
		"items":{
			"type":"object",
			"required":["id","format","target_path","required"],
			"properties":{
				"format":{"const":"yaml"},
				"required":{"type":"boolean"}
			}
		}
	}`))

	require.ErrorContains(t, err, "format must be json")
}
