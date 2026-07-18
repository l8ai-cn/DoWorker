package fixture

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anthropics/agentsmesh/tests/mcp-e2e/client"
)

func mcpEndpointForRunner(
	env *Env,
	rest *client.REST,
	runnerID int64,
) (string, error) {
	if runnerID <= 0 {
		return "", fmt.Errorf("pod response did not include runner id")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runners, err := rest.ListRunners(ctx, env.DevOrgSlug)
	if err != nil {
		return "", err
	}
	for _, runner := range runners {
		if runner.ID != runnerID {
			continue
		}
		switch runner.NodeID {
		case env.RunnerNode:
			return env.MCPBaseURL, nil
		case "dev-runner-2":
			if env.SecondaryMCPBaseURL == "" {
				return "", fmt.Errorf(
					"secondary MCP endpoint is not configured",
				)
			}
			return env.SecondaryMCPBaseURL, nil
		default:
			return "", fmt.Errorf(
				"runner %d has unsupported node id %q",
				runnerID,
				runner.NodeID,
			)
		}
	}
	return "", fmt.Errorf("runner %d is absent from organization", runnerID)
}

func waitPodRegistered(
	ctx context.Context,
	mcpBase, podKey string,
	timeout time.Duration,
) error {
	debugURL := mcpDebugURL(mcpBase)
	hc := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			debugURL,
			nil,
		)
		if err != nil {
			return err
		}
		resp, err := hc.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(200 * time.Millisecond)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var listing struct {
			Pods []struct {
				PodKey string `json:"pod_key"`
			} `json:"pods"`
		}
		if err := json.Unmarshal(body, &listing); err == nil {
			for _, pod := range listing.Pods {
				if pod.PodKey == podKey {
					return nil
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	if lastErr != nil {
		return fmt.Errorf("timeout polling %s: %w", debugURL, lastErr)
	}
	return fmt.Errorf("pod %s not in runner /pods after %s", podKey, timeout)
}

func waitPodRunning(
	ctx context.Context,
	rest *client.REST,
	orgSlug, podKey string,
	timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		pod, err := rest.GetPod(ctx, orgSlug, podKey)
		if err == nil {
			switch pod.Status {
			case "running":
				return nil
			case "completed", "error", "terminated":
				return fmt.Errorf("pod entered terminal status %q", pod.Status)
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("pod status did not become running after %s", timeout)
}

func mcpDebugURL(mcpBase string) string {
	const suffix = "/mcp"
	if len(mcpBase) > len(suffix) &&
		mcpBase[len(mcpBase)-len(suffix):] == suffix {
		return mcpBase[:len(mcpBase)-len(suffix)] + "/pods"
	}
	return mcpBase + "/../pods"
}
