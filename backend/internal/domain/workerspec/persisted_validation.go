package workerspec

func NormalizeAndValidatePersisted(spec Spec) (Spec, error) {
	normalized, err := Normalize(spec)
	if err != nil {
		return Spec{}, err
	}
	if err := validateSpec(normalized, validatePersistedModelBinding); err != nil {
		return Spec{}, err
	}
	return normalized, nil
}

func HasResolvedProtocolAdapters(spec Spec) bool {
	main := spec.Runtime.ModelBinding
	if !main.IsEmpty() && main.ProtocolAdapter == "" {
		return false
	}
	for _, binding := range spec.Runtime.ToolModelBindings {
		if binding.ModelBinding.ProtocolAdapter == "" {
			return false
		}
	}
	return true
}
