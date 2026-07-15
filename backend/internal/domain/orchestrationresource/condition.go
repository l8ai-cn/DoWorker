package orchestrationresource

import (
	"fmt"
	"time"
)

const (
	ConditionTrue    = "True"
	ConditionFalse   = "False"
	ConditionUnknown = "Unknown"

	maxConditionMessageRunes = 1000
)

type Condition struct {
	Type               string    `json:"type" yaml:"type"`
	Status             string    `json:"status" yaml:"status"`
	Reason             string    `json:"reason,omitempty" yaml:"reason,omitempty"`
	Message            string    `json:"message,omitempty" yaml:"message,omitempty"`
	ObservedGeneration int64     `json:"observedGeneration,omitempty" yaml:"observedGeneration,omitempty"`
	LastTransitionTime time.Time `json:"lastTransitionTime" yaml:"lastTransitionTime"`
}

func (condition Condition) Validate() error {
	if !pascalCaseIdentifierPattern.MatchString(condition.Type) {
		return fmt.Errorf("condition.type is invalid: %s", summarizeValue(condition.Type))
	}
	switch condition.Status {
	case ConditionTrue, ConditionFalse, ConditionUnknown:
	default:
		return fmt.Errorf("condition.status is invalid: %s", summarizeValue(condition.Status))
	}
	if condition.Reason != "" && !pascalCaseIdentifierPattern.MatchString(condition.Reason) {
		return fmt.Errorf("condition.reason is invalid: %s", summarizeValue(condition.Reason))
	}
	if err := validateMetadataText(
		"condition.message",
		condition.Message,
		maxConditionMessageRunes,
	); err != nil {
		return err
	}
	if condition.ObservedGeneration < 0 {
		return fmt.Errorf("condition.observedGeneration must not be negative")
	}
	if condition.LastTransitionTime.IsZero() {
		return fmt.Errorf("condition.lastTransitionTime must not be zero")
	}
	if _, err := condition.LastTransitionTime.MarshalJSON(); err != nil {
		return fmt.Errorf("condition.lastTransitionTime: %w", err)
	}
	return nil
}
