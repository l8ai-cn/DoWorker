package orchestrationresource

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	"strconv"
)

func TestTypeMetaValidate(t *testing.T) {
	t.Run("accepts valid type metadata", func(t *testing.T) {
		meta := TypeMeta{
			APIVersion: APIVersionV1Alpha1,
			Kind:       "WorkerTemplate",
		}

		require.NoError(t, meta.Validate())
	})

	t.Run("accepts legacy agentsmesh api version", func(t *testing.T) {
		meta := TypeMeta{
			APIVersion: LegacyAPIVersionV1Alpha1,
			Kind:       "WorkerTemplate",
		}

		require.NoError(t, meta.Validate())
	})

	cases := []struct {
		name  string
		meta  TypeMeta
		field string
		value string
	}{
		{
			name:  "rejects unknown api version",
			meta:  TypeMeta{APIVersion: "v1", Kind: "WorkerTemplate"},
			field: "typeMeta.APIVersion",
			value: "v1",
		},
		{
			name:  "rejects empty api version",
			meta:  TypeMeta{APIVersion: "", Kind: "WorkerTemplate"},
			field: "typeMeta.APIVersion",
			value: "",
		},
		{
			name:  "rejects lowercase kind",
			meta:  TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: "worker-template"},
			field: "typeMeta.Kind",
			value: "worker-template",
		},
		{
			name:  "rejects single-character kind",
			meta:  TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: "W"},
			field: "typeMeta.Kind",
			value: "W",
		},
		{
			name:  "rejects too long kind",
			meta:  TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: strings.Repeat("A", 101)},
			field: "typeMeta.Kind",
			value: strings.Repeat("A", 101),
		},
		{
			name:  "rejects kind with underscore",
			meta:  TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: "Worker_Template"},
			field: "typeMeta.Kind",
			value: "Worker_Template",
		},
		{
			name:  "rejects kind with dot",
			meta:  TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: "Worker.Template"},
			field: "typeMeta.Kind",
			value: "Worker.Template",
		},
		{
			name:  "rejects kind with space",
			meta:  TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: "Worker Template"},
			field: "typeMeta.Kind",
			value: "Worker Template",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.meta.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.field)
			if tc.value == "" {
				require.Contains(t, err.Error(), `""`)
				return
			}
			summaryProbe := tc.value
			if len(summaryProbe) > 80 {
				summaryProbe = summaryProbe[:80]
			}
			require.Contains(t, err.Error(), summaryProbe)
		})
	}
}

func TestTypeMetaValidateRejectsLongChineseAndEmojiKindWithUtf8SafeSummary(t *testing.T) {
	meta := TypeMeta{
		APIVersion: APIVersionV1Alpha1,
		Kind:       strings.Repeat("中", 79) + "😀" + strings.Repeat("文", 3),
	}
	require.Error(t, meta.Validate())

	summary1 := summarizeValue(meta.Kind)
	summary2 := summarizeValue(meta.Kind)
	require.Equal(t, summary1, summary2)

	unquoted, err := strconv.Unquote(summary1)
	require.NoError(t, err)
	require.True(t, utf8.ValidString(unquoted))
	require.Equal(t, 80, utf8.RuneCountInString(unquoted))
	require.True(t, strings.ContainsRune(unquoted, '😀'))
}
