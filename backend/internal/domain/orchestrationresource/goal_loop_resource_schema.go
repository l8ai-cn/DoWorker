package orchestrationresource

import "fmt"

type GoalLoopResourceSpec struct {
	WorkerTemplateRef   Reference                `json:"workerTemplateRef" yaml:"workerTemplateRef"`
	Description         string                   `json:"description" yaml:"description"`
	Objective           string                   `json:"objective" yaml:"objective"`
	AcceptanceCriteria  []string                 `json:"acceptanceCriteria" yaml:"acceptanceCriteria"`
	VerificationCommand string                   `json:"verificationCommand" yaml:"verificationCommand"`
	MaxIterations       int                      `json:"maxIterations" yaml:"maxIterations"`
	TokenBudget         *int64                   `json:"tokenBudget,omitempty" yaml:"tokenBudget,omitempty"`
	TimeoutMinutes      int                      `json:"timeoutMinutes" yaml:"timeoutMinutes"`
	NoProgressLimit     int                      `json:"noProgressLimit" yaml:"noProgressLimit"`
	SameErrorLimit      int                      `json:"sameErrorLimit" yaml:"sameErrorLimit"`
	EscalationPolicy    string                   `json:"escalationPolicy" yaml:"escalationPolicy"`
	LoopProgram         *GoalLoopProgramSnapshot `json:"loopProgram,omitempty" yaml:"loopProgram,omitempty"`
}

func goalLoopResourceSchema() Schema {
	return Schema{
		NewSpec: func() any { return &GoalLoopResourceSpec{} },
		Validate: func(metadata Metadata, value any) error {
			return validateGoalLoopResource(metadata, value.(*GoalLoopResourceSpec))
		},
	}
}

func validateGoalLoopResource(
	metadata Metadata,
	spec *GoalLoopResourceSpec,
) error {
	if err := validateDefinitionReference(
		metadata,
		"workerTemplateRef",
		KindWorkerTemplate,
		spec.WorkerTemplateRef,
	); err != nil {
		return err
	}
	if err := validateDefinitionText(
		"description",
		spec.Description,
		4_000,
		false,
	); err != nil {
		return err
	}
	if err := validateDefinitionText(
		"objective",
		spec.Objective,
		16_384,
		true,
	); err != nil {
		return err
	}
	if len(spec.AcceptanceCriteria) == 0 ||
		len(spec.AcceptanceCriteria) > 64 {
		return fmt.Errorf("acceptanceCriteria must contain 1-64 items")
	}
	for index, criterion := range spec.AcceptanceCriteria {
		if err := validateDefinitionText(
			fmt.Sprintf("acceptanceCriteria[%d]", index),
			criterion,
			2_000,
			true,
		); err != nil {
			return err
		}
	}
	if err := validateDefinitionText(
		"verificationCommand",
		spec.VerificationCommand,
		8_192,
		true,
	); err != nil {
		return err
	}
	if spec.MaxIterations < 1 || spec.MaxIterations > 100 {
		return fmt.Errorf("maxIterations must be between 1 and 100")
	}
	if spec.TokenBudget != nil && *spec.TokenBudget <= 0 {
		return fmt.Errorf("tokenBudget must be positive")
	}
	if spec.TimeoutMinutes < 1 || spec.TimeoutMinutes > 1_440 {
		return fmt.Errorf("timeoutMinutes must be between 1 and 1440")
	}
	if spec.NoProgressLimit < 1 || spec.NoProgressLimit > 20 {
		return fmt.Errorf("noProgressLimit must be between 1 and 20")
	}
	if spec.SameErrorLimit < 1 || spec.SameErrorLimit > 20 {
		return fmt.Errorf("sameErrorLimit must be between 1 and 20")
	}
	if spec.EscalationPolicy != "pause" && spec.EscalationPolicy != "fail" {
		return fmt.Errorf("escalationPolicy must be pause or fail")
	}
	return validateGoalLoopProgramSnapshot(metadata, spec)
}
