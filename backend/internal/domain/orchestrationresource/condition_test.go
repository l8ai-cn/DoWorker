package orchestrationresource

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func validConditionForTest() Condition {
	return Condition{
		Type:               "Ready",
		Status:             ConditionTrue,
		Reason:             "Reconciled",
		Message:            "worker 已就绪",
		ObservedGeneration: 3,
		LastTransitionTime: time.Date(2026, time.July, 14, 8, 30, 0, 0, time.UTC),
	}
}

func TestConditionShape(t *testing.T) {
	conditionType := reflect.TypeOf(Condition{})
	expectedFields := map[string]struct {
		goType  reflect.Type
		jsonTag string
		yamlTag string
	}{
		"Type":               {goType: reflect.TypeOf(""), jsonTag: "type", yamlTag: "type"},
		"Status":             {goType: reflect.TypeOf(""), jsonTag: "status", yamlTag: "status"},
		"Reason":             {goType: reflect.TypeOf(""), jsonTag: "reason,omitempty", yamlTag: "reason,omitempty"},
		"Message":            {goType: reflect.TypeOf(""), jsonTag: "message,omitempty", yamlTag: "message,omitempty"},
		"ObservedGeneration": {goType: reflect.TypeOf(int64(0)), jsonTag: "observedGeneration,omitempty", yamlTag: "observedGeneration,omitempty"},
		"LastTransitionTime": {goType: reflect.TypeOf(time.Time{}), jsonTag: "lastTransitionTime", yamlTag: "lastTransitionTime"},
	}

	require.Equal(t, len(expectedFields), conditionType.NumField())
	for name, expected := range expectedFields {
		field, found := conditionType.FieldByName(name)
		require.True(t, found, "missing field %s", name)
		require.Equal(t, expected.goType, field.Type)
		require.Equal(t, expected.jsonTag, field.Tag.Get("json"))
		require.Equal(t, expected.yamlTag, field.Tag.Get("yaml"))
	}
}

func TestConditionValidateAcceptsSupportedStatuses(t *testing.T) {
	for _, status := range []string{ConditionTrue, ConditionFalse, ConditionUnknown} {
		t.Run(status, func(t *testing.T) {
			condition := validConditionForTest()
			condition.Status = status

			require.NoError(t, condition.Validate())
		})
	}
}

func TestConditionValidateAcceptsEmptyReasonAndThousandRuneMessage(t *testing.T) {
	condition := validConditionForTest()
	condition.Reason = ""
	condition.Message = strings.Repeat("界", 1000)

	require.NoError(t, condition.Validate())
}

func TestConditionValidateAcceptsZeroObservedGeneration(t *testing.T) {
	condition := validConditionForTest()
	condition.ObservedGeneration = 0

	require.NoError(t, condition.Validate())
}

func TestConditionValidateAcceptsTypeLengthBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "two characters", value: "AB"},
		{name: "one hundred characters", value: "A" + strings.Repeat("a", 99)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := validConditionForTest()
			condition.Type = tt.value

			require.NoError(t, condition.Validate())
		})
	}
}

func TestConditionValidateAcceptsNonEmptyReasonLengthBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "two characters", value: "AB"},
		{name: "one hundred characters", value: "A" + strings.Repeat("1", 99)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := validConditionForTest()
			condition.Reason = tt.value

			require.NoError(t, condition.Validate())
		})
	}
}

func TestConditionValidateLastTransitionTimeJSONBoundaries(t *testing.T) {
	for _, year := range []int{0, 9999} {
		t.Run(time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).Format("2006"), func(t *testing.T) {
			condition := validConditionForTest()
			condition.LastTransitionTime = time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)

			require.NoError(t, condition.Validate())
		})
	}

	tests := []struct {
		name  string
		value time.Time
	}{
		{name: "year below range", value: time.Date(-1, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{name: "year above range", value: time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{
			name:  "zone offset outside json range",
			value: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.FixedZone("invalid", 24*60*60)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := validConditionForTest()
			condition.LastTransitionTime = tt.value

			err := condition.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), "condition.lastTransitionTime")
		})
	}
}

func TestConditionValidateRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		mutate func(*Condition)
	}{
		{
			name:   "missing type",
			path:   "condition.type",
			mutate: func(condition *Condition) { condition.Type = "" },
		},
		{
			name:   "invalid type",
			path:   "condition.type",
			mutate: func(condition *Condition) { condition.Type = "ready-state" },
		},
		{
			name:   "invalid status",
			path:   "condition.status",
			mutate: func(condition *Condition) { condition.Status = "Ready" },
		},
		{
			name:   "invalid reason",
			path:   "condition.reason",
			mutate: func(condition *Condition) { condition.Reason = "reconcile_failed" },
		},
		{
			name:   "message too long",
			path:   "condition.message",
			mutate: func(condition *Condition) { condition.Message = strings.Repeat("界", 1001) },
		},
		{
			name:   "message contains ascii control",
			path:   "condition.message",
			mutate: func(condition *Condition) { condition.Message = "worker\nready" },
		},
		{
			name:   "message contains unicode control",
			path:   "condition.message",
			mutate: func(condition *Condition) { condition.Message = "worker\u0085ready" },
		},
		{
			name:   "negative observed generation",
			path:   "condition.observedGeneration",
			mutate: func(condition *Condition) { condition.ObservedGeneration = -1 },
		},
		{
			name:   "zero transition time",
			path:   "condition.lastTransitionTime",
			mutate: func(condition *Condition) { condition.LastTransitionTime = time.Time{} },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := validConditionForTest()
			tt.mutate(&condition)

			err := condition.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.path)
		})
	}
}

func TestConditionValidateDoesNotDiscloseFullUntrustedInput(t *testing.T) {
	untrusted := "A" + strings.Repeat("x", 99) + "!"
	condition := validConditionForTest()
	condition.Type = untrusted

	err := condition.Validate()
	require.Error(t, err)
	require.NotContains(t, err.Error(), untrusted)
	require.Contains(t, err.Error(), "condition.type")
}
