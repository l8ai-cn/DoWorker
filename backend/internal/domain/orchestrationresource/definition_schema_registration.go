package orchestrationresource

import "fmt"

const (
	KindPrompt   = "Prompt"
	KindWorker   = "Worker"
	KindExpert   = "Expert"
	KindWorkflow = "Workflow"
	KindGoalLoop = "GoalLoop"
)

func RegisterDefinitionSchemas(registry *Registry) error {
	if registry == nil {
		return fmt.Errorf("definition schema registry must not be nil")
	}
	registrations := []struct {
		kind   string
		schema Schema
	}{
		{KindPrompt, promptSchema()},
		{KindWorker, workerInvocationSchema()},
		{KindExpert, expertResourceSchema()},
		{KindWorkflow, workflowResourceSchema()},
		{KindGoalLoop, goalLoopResourceSchema()},
	}
	for _, registration := range registrations {
		if err := registry.Register(TypeMeta{
			APIVersion: APIVersionV1Alpha1,
			Kind:       registration.kind,
		}, registration.schema); err != nil {
			return err
		}
	}
	return nil
}
