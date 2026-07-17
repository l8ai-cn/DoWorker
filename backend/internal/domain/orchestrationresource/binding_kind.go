package orchestrationresource

func IsBindingKind(kind string) bool {
	switch kind {
	case KindModelBinding,
		KindRepository,
		KindSkill,
		KindKnowledgeBase,
		KindEnvironmentBundle,
		KindComputeTarget,
		KindResourceProfile,
		KindToolBinding:
		return true
	default:
		return false
	}
}
