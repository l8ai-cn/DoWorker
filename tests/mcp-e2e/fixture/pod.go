package fixture

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

func NewEchoPod(t *testing.T, env *Env, rest *client.REST) *EchoPod {
	t.Helper()
	return newEchoPod(t, env, rest, "pty")
}

func NewACPEchoPod(t *testing.T, env *Env, rest *client.REST) *EchoPod {
	t.Helper()
	return newEchoPod(t, env, rest, "acp")
}

func newEchoPod(
	t *testing.T,
	env *Env,
	rest *client.REST,
	interactionMode string,
) *EchoPod {
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

	mcpEndpoint, err := mcpEndpointForRunner(env, rest, pod.RunnerID)
	if err != nil {
		t.Fatalf("resolve MCP endpoint for pod %s: %v", pod.PodKey, err)
	}
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

func uniqueAlias(prefix string) string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%s-%d-%s", prefix, time.Now().UnixMilli(), hex.EncodeToString(b[:]))
}

func UniqueResourceName(prefix string) string {
	return uniqueAlias(prefix)
}
