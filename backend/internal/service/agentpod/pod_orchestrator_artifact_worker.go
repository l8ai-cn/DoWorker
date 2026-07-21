package agentpod

import (
	"context"
	"fmt"
	"strings"

	"github.com/l8ai-cn/agentcloud/agentfile/extract"
	"github.com/l8ai-cn/agentcloud/agentfile/parser"
	agentDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
)

func (o *PodOrchestrator) runtimeAgentDefinition(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
) (*agentDomain.Agent, error) {
	if req.preResolvedDependencies != nil {
		return artifactWorkerAgent(req.preResolvedDependencies)
	}
	if req.AgentSlug == "" || o.agentResolver == nil {
		return nil, ErrMissingAgentSlug
	}
	agentDef, err := o.agentResolver.GetAgent(ctx, req.AgentSlug)
	if err != nil {
		return nil, ErrMissingAgentSlug
	}
	return agentDef, nil
}

func artifactWorkerAgent(
	document *workerdependency.Document,
) (*agentDomain.Agent, error) {
	if document == nil || strings.TrimSpace(document.Worker.AgentfileSource) == "" ||
		document.Worker.WorkerType.String() == "" {
		return nil, ErrWorkerSpecDependencyUnavailable
	}
	if document.Worker.AdapterID.String() == "" {
		return nil, ErrMissingAgentAdapter
	}
	program, parseErrors := parser.Parse(document.Worker.AgentfileSource)
	if len(parseErrors) > 0 {
		return nil, fmt.Errorf(
			"%w: artifact AgentFile is invalid: %s",
			ErrWorkerSpecDependencyUnavailable,
			parseErrors[0],
		)
	}
	spec := extract.Extract(program)
	modes := artifactWorkerModes(program)
	source := document.Worker.AgentfileSource
	return &agentDomain.Agent{
		Slug:            document.Worker.WorkerType.String(),
		LaunchCommand:   spec.Agent.Executable,
		Executable:      spec.Agent.Executable,
		AdapterID:       document.Worker.AdapterID.String(),
		AgentfileSource: &source,
		SupportedModes:  strings.Join(modes, ","),
	}, nil
}

func artifactWorkerModes(program *parser.Program) []string {
	seen := map[string]struct{}{}
	modes := []string{}
	for _, declaration := range program.Declarations {
		var mode string
		switch value := declaration.(type) {
		case *parser.ModeDecl:
			mode = value.Mode
		case *parser.ModeArgsDecl:
			mode = value.Mode
		}
		if mode == "" {
			continue
		}
		if _, exists := seen[mode]; exists {
			continue
		}
		seen[mode] = struct{}{}
		modes = append(modes, mode)
	}
	return modes
}
