package orchestrationresource

type WorkerInvocationSpec struct {
	WorkerTemplateRef Reference         `json:"workerTemplateRef" yaml:"workerTemplateRef"`
	PromptRef         *Reference        `json:"promptRef,omitempty" yaml:"promptRef,omitempty"`
	Inputs            map[string]string `json:"inputs" yaml:"inputs"`
	Alias             string            `json:"alias" yaml:"alias"`
}

func workerInvocationSchema() Schema {
	return Schema{
		NewSpec: func() any { return &WorkerInvocationSpec{} },
		Validate: func(metadata Metadata, value any) error {
			spec := value.(*WorkerInvocationSpec)
			if err := validateDefinitionReference(
				metadata,
				"workerTemplateRef",
				KindWorkerTemplate,
				spec.WorkerTemplateRef,
			); err != nil {
				return err
			}
			if spec.PromptRef != nil {
				if err := validateDefinitionReference(
					metadata,
					"promptRef",
					KindPrompt,
					*spec.PromptRef,
				); err != nil {
					return err
				}
			}
			if err := validateDefinitionStringMap(
				"inputs",
				spec.Inputs,
				128,
				8_192,
			); err != nil {
				return err
			}
			return validateDefinitionText("alias", spec.Alias, 100, false)
		},
	}
}
