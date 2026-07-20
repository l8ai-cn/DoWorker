package workspace

import (
	"context"
	"fmt"
)

type PreparationStep interface {
	Name() string
	Execute(ctx context.Context, prepCtx *PreparationContext) error
}

type PreparationContext struct {
	PodID        string
	TicketSlug   string
	BranchName   string
	WorkspaceDir string
	MainRepoDir  string
	BaseEnvVars  map[string]string
	UnsetEnvVars []string
}

func (c *PreparationContext) GetEnvVars() map[string]string {
	result := make(map[string]string)
	for key, value := range c.BaseEnvVars {
		result[key] = value
	}
	result["WORKSPACE_DIR"] = c.WorkspaceDir
	if c.MainRepoDir != "" {
		result["MAIN_REPO_DIR"] = c.MainRepoDir
	}
	if c.TicketSlug != "" {
		result["TICKET_SLUG"] = c.TicketSlug
	}
	if c.BranchName != "" {
		result["BRANCH_NAME"] = c.BranchName
	}
	return result
}

func (c *PreparationContext) String() string {
	return fmt.Sprintf(
		"PreparationContext{PodID: %s, Ticket: %s, WorkspaceDir: %s}",
		c.PodID, c.TicketSlug, c.WorkspaceDir,
	)
}

type PreparationError struct {
	Step   string
	Cause  error
	Output string
}

func (e *PreparationError) Error() string {
	if e.Output != "" {
		return fmt.Sprintf("preparation step '%s' failed: %v\nOutput: %s", e.Step, e.Cause, e.Output)
	}
	return fmt.Sprintf("preparation step '%s' failed: %v", e.Step, e.Cause)
}

func (e *PreparationError) Unwrap() error {
	return e.Cause
}
