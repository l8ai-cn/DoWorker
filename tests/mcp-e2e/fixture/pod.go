package fixture

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/tests/mcp-e2e/client"
)

// EchoPod is a Pod created from the e2e-echo agent that's already registered
// with the runner's MCP server, ready to accept tool calls.
type EchoPod struct {
	Pod *client.Pod
	MCP *client.MCPClient
}

// NewEchoPod creates a fresh pod via REST, waits for the runner to register
// it, then hands back an MCP client scoped to it. t.Cleanup terminates the
// pod so failures still leave the stack clean.
//
// The MCP endpoint the agent connects to is the one belonging to the runner
// the pod was scheduled on: dev-runner → env.MCPBaseURL, dev-runner-2 →
// env.SecondaryMCPBaseURL. We pick by runnerID after backend tells us which
// runner accepted the schedule.
func NewEchoPod(t *testing.T, env *Env, rest *client.REST, runnerID int64) *EchoPod {
	t.Helper()
	// interactive + MODE pty: default autonomous forces MODE acp, which
	// breaks send_pod_input → "got: …" PTY round-trip assertions.
	ptyLayer := "MODE pty"
	return newEchoPod(t, env, rest, runnerID, &ptyLayer)
}

func NewACPEchoPod(t *testing.T, env *Env, rest *client.REST, runnerID int64) *EchoPod {
	t.Helper()
	acpLayer := "MODE acp"
	return newEchoPod(t, env, rest, runnerID, &acpLayer)
}

func newEchoPod(t *testing.T, env *Env, rest *client.REST, runnerID int64, agentfileLayer *string) *EchoPod {
	t.Helper()
	alias := uniqueAlias("e2e-echo")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pod, err := rest.CreatePod(ctx, env.DevOrgSlug, client.CreatePodRequest{
		AgentSlug:       "e2e-echo",
		RunnerID:        runnerID,
		Alias:           &alias,
		AgentfileLayer:  agentfileLayer,
		AutomationLevel: "interactive",
		Cols:            80,
		Rows:            24,
	})
	if err != nil {
		t.Fatalf("create echo pod: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = rest.TerminatePod(ctx, env.DevOrgSlug, pod.PodKey)
	})

	mcpEndpoint := mcpEndpointForRunner(env, rest, runnerID)
	if err := waitPodRegistered(ctx, mcpEndpoint, pod.PodKey, 15*time.Second); err != nil {
		t.Fatalf("pod %s never registered with runner MCP at %s: %v", pod.PodKey, mcpEndpoint, err)
	}
	if err := waitPodRunning(ctx, rest, env.DevOrgSlug, pod.PodKey, 15*time.Second); err != nil {
		t.Fatalf("pod %s never became routable through backend: %v", pod.PodKey, err)
	}

	return &EchoPod{
		Pod: pod,
		MCP: client.NewMCP(mcpEndpoint, pod.PodKey),
	}
}

// mcpEndpointForRunner picks the right runner's MCP HTTP base URL. dev-runner
// and dev-runner-2 are the only configured runners; for any other runnerID we
// fall back to the primary (matches single-runner pre-cross-runner behavior).
func mcpEndpointForRunner(env *Env, rest *client.REST, runnerID int64) string {
	if env.SecondaryMCPBaseURL == "" {
		return env.MCPBaseURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	runners, err := rest.ListRunners(ctx, env.DevOrgSlug)
	if err != nil {
		return env.MCPBaseURL
	}
	for _, r := range runners {
		if r.ID == runnerID && r.NodeID == "dev-runner-2" {
			return env.SecondaryMCPBaseURL
		}
	}
	return env.MCPBaseURL
}

func waitPodRegistered(ctx context.Context, mcpBase, podKey string, timeout time.Duration) error {
	debugURL := mcpDebugURL(mcpBase)
	hc := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, debugURL, nil)
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
			for _, p := range listing.Pods {
				if p.PodKey == podKey {
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

// mcpDebugURL strips the "/mcp" suffix to reach the sibling "/pods" endpoint
// served by the same runner HTTP server.
func mcpDebugURL(mcpBase string) string {
	const suffix = "/mcp"
	if len(mcpBase) > len(suffix) && mcpBase[len(mcpBase)-len(suffix):] == suffix {
		return mcpBase[:len(mcpBase)-len(suffix)] + "/pods"
	}
	return mcpBase + "/../pods"
}

func uniqueAlias(prefix string) string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%s-%d-%s", prefix, time.Now().UnixMilli(), hex.EncodeToString(b[:]))
}
