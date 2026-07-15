package loopscript

func CompileGoalLoopV1(program *Program) (*GoalLoopLaunchSpec, []Diagnostic) {
	if diagnostics := validateProgram(program, nil); len(diagnostics) != 0 {
		return nil, diagnostics
	}

	loop := program.Loop
	return &GoalLoopLaunchSpec{
		Name:                loop.LocalID,
		Slug:                loop.LocalID,
		WorkerSnapshotID:    loop.Worker.SnapshotID,
		Objective:           loop.Repeat.Agent.Prompt,
		AcceptanceCriteria:  []string{loop.Repeat.Verifier.Accept},
		VerificationCommand: loop.Repeat.Verifier.Command,
		MaxIterations:       int(loop.Repeat.Max),
		TokenBudget:         loop.Limits.Tokens,
		TimeoutMinutes:      int(loop.Limits.TimeoutMins),
		NoProgressLimit:     int(loop.Limits.NoProgress),
		SameErrorLimit:      int(loop.Limits.SameError),
		EscalationPolicy:    string(loop.FailurePolicy),
	}, nil
}
