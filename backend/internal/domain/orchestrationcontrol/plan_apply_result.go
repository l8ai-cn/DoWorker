package orchestrationcontrol

func (plan Plan) validateApplyResult(
	resourceID int64,
	identity ResourceIdentity,
	resourceVersion int64,
	revision int64,
) error {
	if err := identity.Validate(plan.Scope); err != nil {
		return err
	}
	if resourceID <= 0 {
		return invalid("resultResourceId", "must be positive")
	}
	if identity.ResourceTarget != plan.Target {
		return invalid("resultIdentity", "must match the plan target")
	}
	if resourceVersion <= 0 || revision <= 0 || resourceVersion < revision {
		return invalid("result counters", "must identify an existing revision")
	}
	switch plan.Operation {
	case PlanOperationCreate:
		if resourceVersion != 1 || revision != 1 {
			return invalid("result counters", "must start at one for create")
		}
	case PlanOperationUpdate:
		if resourceID != plan.TargetResourceID {
			return invalid("resultResourceId", "must match the update target")
		}
		if identity.UID != plan.BaseUID {
			return invalid("resultIdentity.uid", "must match the update base UID")
		}
		if resourceVersion <= plan.BaseResourceVersion {
			return invalid("resultResourceVersion", "must advance the base version")
		}
	default:
		return invalid("operation", "must be create or update")
	}
	return nil
}
