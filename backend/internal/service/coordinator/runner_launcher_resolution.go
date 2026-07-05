package coordinator

import (
	"fmt"
	"strings"
)

func (e runnerContainerEnv) imageForAgent(agentSlug string) (string, error) {
	agentSlug = strings.TrimSpace(agentSlug)
	if image := strings.TrimSpace(e.AgentImages[agentSlug]); image != "" {
		return image, nil
	}
	return "", fmt.Errorf("coordinator: COORDINATOR_RUNNER_IMAGES must include %q", agentSlug)
}

func (c dockerLauncherConfig) composeServiceForAgent(agentSlug string) (string, error) {
	agentSlug = strings.TrimSpace(agentSlug)
	if service := strings.TrimSpace(c.ComposeServices[agentSlug]); service != "" {
		return service, nil
	}
	return "", fmt.Errorf("coordinator: COORDINATOR_RUNNER_DOCKER_COMPOSE_SERVICES must include %q", agentSlug)
}
