package runner

import (
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func TestHandleHeartbeat(t *testing.T) {
	pc, _, _, db := setupPodEventHandlerDeps(t)

	// Create a runner
	r := &runner.Runner{
		OrganizationID: 1,
		NodeID:         "heartbeat-node",
		Status:         "online",
	}
	if err := db.Create(r).Error; err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	// Create a pod
	staleAt := time.Now().Add(-time.Hour)
	db.Exec(`INSERT INTO pods (pod_key, runner_id, status, last_activity) VALUES (?, ?, ?, ?)`,
		"heartbeat-pod-1", r.ID, agentpod.StatusRunning, staleAt)

	// Send heartbeat (using Proto type)
	data := &runnerv1.HeartbeatData{
		Pods: []*runnerv1.PodInfo{
			{PodKey: "heartbeat-pod-1", Status: "running"},
		},
	}

	pc.handleHeartbeat(r.ID, data)

	// Verify heartbeat was recorded (check buffer)
	if pc.heartbeatBatcher.BufferSize() != 1 {
		t.Errorf("heartbeat should be recorded, buffer size: %d", pc.heartbeatBatcher.BufferSize())
	}

	var lastActivity time.Time
	db.Raw(`SELECT last_activity FROM pods WHERE pod_key = ?`, "heartbeat-pod-1").Scan(&lastActivity)
	if !lastActivity.After(staleAt.Add(30 * time.Minute)) {
		t.Errorf("last_activity was not refreshed: got %v", lastActivity)
	}
}

func TestHandleHeartbeatSyncsAgentStatus(t *testing.T) {
	pc, _, _, db := setupPodEventHandlerDeps(t)

	// Create a runner
	r := &runner.Runner{
		OrganizationID: 1,
		NodeID:         "hb-agent-sync-node",
		Status:         "online",
	}
	if err := db.Create(r).Error; err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	// Create a pod with idle agent_status
	db.Exec(`INSERT INTO pods (pod_key, runner_id, status, agent_status) VALUES (?, ?, ?, ?)`,
		"hb-agent-pod-1", r.ID, agentpod.StatusRunning, agentpod.AgentStatusIdle)

	// Send heartbeat with AgentStatus set to executing
	data := &runnerv1.HeartbeatData{
		Pods: []*runnerv1.PodInfo{
			{PodKey: "hb-agent-pod-1", Status: "running", AgentStatus: "executing"},
		},
	}

	pc.handleHeartbeat(r.ID, data)

	// Verify agent_status was updated in DB
	var agentStatus string
	db.Raw(`SELECT agent_status FROM pods WHERE pod_key = ?`, "hb-agent-pod-1").
		Scan(&agentStatus)

	if agentStatus != agentpod.AgentStatusExecuting {
		t.Errorf("agent_status: got %q, want %q", agentStatus, agentpod.AgentStatusExecuting)
	}
}

func TestHandleHeartbeatSkipsEmptyAgentStatus(t *testing.T) {
	pc, _, _, db := setupPodEventHandlerDeps(t)

	// Create a runner
	r := &runner.Runner{
		OrganizationID: 1,
		NodeID:         "hb-empty-agent-node",
		Status:         "online",
	}
	if err := db.Create(r).Error; err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	// Create a pod with executing agent_status
	db.Exec(`INSERT INTO pods (pod_key, runner_id, status, agent_status) VALUES (?, ?, ?, ?)`,
		"hb-empty-pod-1", r.ID, agentpod.StatusRunning, agentpod.AgentStatusExecuting)

	// Send heartbeat with empty AgentStatus
	data := &runnerv1.HeartbeatData{
		Pods: []*runnerv1.PodInfo{
			{PodKey: "hb-empty-pod-1", Status: "running", AgentStatus: ""},
		},
	}

	pc.handleHeartbeat(r.ID, data)

	// Verify agent_status was NOT modified (should still be executing)
	var agentStatus string
	db.Raw(`SELECT agent_status FROM pods WHERE pod_key = ?`, "hb-empty-pod-1").
		Scan(&agentStatus)

	if agentStatus != agentpod.AgentStatusExecuting {
		t.Errorf("agent_status should not be modified when heartbeat AgentStatus is empty: got %q, want %q",
			agentStatus, agentpod.AgentStatusExecuting)
	}
}

func TestHandleHeartbeatReconcilePods(t *testing.T) {
	pc, _, tr, db := setupPodEventHandlerDeps(t)

	// Create a runner
	r := &runner.Runner{
		OrganizationID: 1,
		NodeID:         "reconcile-node",
		Status:         "online",
	}
	if err := db.Create(r).Error; err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	// Create pods in DB
	db.Exec(`INSERT INTO pods (pod_key, runner_id, status) VALUES (?, ?, ?)`,
		"reconcile-pod-1", r.ID, agentpod.StatusRunning)
	db.Exec(`INSERT INTO pods (pod_key, runner_id, status) VALUES (?, ?, ?)`,
		"reconcile-pod-2", r.ID, agentpod.StatusRunning)

	// Send heartbeat with only pod-1 (using Proto type)
	data := &runnerv1.HeartbeatData{
		Pods: []*runnerv1.PodInfo{
			{PodKey: "reconcile-pod-1", Status: "running"},
		},
	}

	// Need orphanMissThreshold heartbeats for pod-2 to become orphaned
	for i := 0; i < orphanMissThreshold; i++ {
		pc.handleHeartbeat(r.ID, data)
	}

	// Verify pod-1 is still running and registered
	var status1 string
	db.Raw(`SELECT status FROM pods WHERE pod_key = ?`, "reconcile-pod-1").Scan(&status1)
	if status1 != agentpod.StatusRunning {
		t.Errorf("pod-1 status: got %q, want %q", status1, agentpod.StatusRunning)
	}
	if !tr.IsPodRegistered("reconcile-pod-1") {
		t.Error("pod-1 should be registered with terminal router")
	}

	// Verify pod-2 is orphaned
	var status2 string
	db.Raw(`SELECT status FROM pods WHERE pod_key = ?`, "reconcile-pod-2").Scan(&status2)
	if status2 != agentpod.StatusOrphaned {
		t.Errorf("pod-2 status: got %q, want %q", status2, agentpod.StatusOrphaned)
	}
}

func TestHandleHeartbeatRestoreOrphanedPod(t *testing.T) {
	pc, _, _, db := setupPodEventHandlerDeps(t)

	// Create a runner
	r := &runner.Runner{
		OrganizationID: 1,
		NodeID:         "restore-node",
		Status:         "online",
	}
	if err := db.Create(r).Error; err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	// Create an orphaned pod
	db.Exec(`INSERT INTO pods (pod_key, runner_id, status) VALUES (?, ?, ?)`,
		"orphan-pod-1", r.ID, agentpod.StatusOrphaned)

	// Send heartbeat reporting the orphaned pod as running (using Proto type)
	data := &runnerv1.HeartbeatData{
		Pods: []*runnerv1.PodInfo{
			{PodKey: "orphan-pod-1", Status: "running"},
		},
	}

	pc.handleHeartbeat(r.ID, data)

	// Verify pod was restored
	var status string
	db.Raw(`SELECT status FROM pods WHERE pod_key = ?`, "orphan-pod-1").Scan(&status)
	if status != agentpod.StatusRunning {
		t.Errorf("orphaned pod should be restored: got %q, want %q", status, agentpod.StatusRunning)
	}
}

func TestHandleHeartbeatRestoresDisconnectedPod(t *testing.T) {
	pc, _, _, db := setupPodEventHandlerDeps(t)

	r := &runner.Runner{
		OrganizationID: 1,
		NodeID:         "restore-disconnected-node",
		Status:         "online",
	}
	if err := db.Create(r).Error; err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	staleAt := time.Now().Add(-time.Hour)
	db.Exec(`INSERT INTO pods (pod_key, runner_id, status, last_activity) VALUES (?, ?, ?, ?)`,
		"disconnected-pod-1", r.ID, agentpod.StatusDisconnected, staleAt)

	data := &runnerv1.HeartbeatData{
		Pods: []*runnerv1.PodInfo{
			{PodKey: "disconnected-pod-1", Status: "running"},
		},
	}

	pc.handleHeartbeat(r.ID, data)

	var pod struct {
		Status       string
		LastActivity time.Time
	}
	db.Raw(`SELECT status, last_activity FROM pods WHERE pod_key = ?`, "disconnected-pod-1").Scan(&pod)

	if pod.Status != agentpod.StatusRunning {
		t.Errorf("disconnected pod should be restored: got %q", pod.Status)
	}
	if !pod.LastActivity.After(staleAt.Add(30 * time.Minute)) {
		t.Errorf("last_activity was not refreshed: got %v", pod.LastActivity)
	}
}
