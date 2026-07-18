package workercreation

func (service *Service) ModelBindingProtocolAdapters(
	workerType string,
) ([]string, bool) {
	if service == nil || service.workspaceDeps.Definitions == nil {
		return nil, false
	}
	definition, found := service.workspaceDeps.Definitions.Get(workerType)
	if !found || !definition.ModelRequirement.Required {
		return nil, false
	}
	return append([]string{}, definition.ModelRequirement.ProtocolAdapters...), true
}
