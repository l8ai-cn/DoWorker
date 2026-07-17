package expert

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpertMarketSourceJSONFieldsMatchColumns(t *testing.T) {
	expertType := reflect.TypeOf(Expert{})
	for fieldName, want := range map[string]string{
		"SourceMarketApplicationID": "source_market_application_id,omitempty",
		"SourceMarketReleaseID":     "source_market_release_id,omitempty",
	} {
		field, ok := expertType.FieldByName(fieldName)
		require.True(t, ok, fieldName)
		require.Equal(t, want, field.Tag.Get("json"))
	}
}
