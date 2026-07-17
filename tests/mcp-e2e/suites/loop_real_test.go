package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/tests/mcp-e2e/client"
	"github.com/anthropics/agentsmesh/tests/mcp-e2e/fixture"
)

func TestGoalLoop_ListShowsCreatedLoop(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	snapshotID := fixture.NewGoalLoopSnapshot(t, env)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	name := fmt.Sprintf("e2e-goal-loop-list-%d", time.Now().UnixMilli())
	loop, err := rest.CreateGoalLoop(ctx, env.DevOrgSlug, client.CreateGoalLoopRequest{
		Name:                 name,
		WorkerSpecSnapshotID: snapshotID,
		Objective:            "Create a valid goal loop.",
		AcceptanceCriteria:   []string{"The loop is visible in its organization."},
		VerificationCommand:  "true",
	})
	if err != nil {
		t.Fatalf("create goal loop: %v", err)
	}
	if loop.Status != "draft" {
		t.Fatalf("created goal loop status = %q, want draft", loop.Status)
	}

	page, err := rest.ListGoalLoops(ctx, env.DevOrgSlug, name, 10, 0)
	if err != nil {
		t.Fatalf("list goal loops: %v", err)
	}
	for _, item := range page.Items {
		if item.ID == loop.ID && item.Slug == loop.Slug {
			return
		}
	}
	t.Fatalf("created goal loop %q was absent from filtered list: %+v", loop.Slug, page.Items)
}

func TestGoalLoop_StaleSnapshotDoesNotStart(t *testing.T) {
	env := fixture.LoadEnv(t)
	rest := fixture.SharedREST(t, env)
	snapshotID := fixture.NewGoalLoopSnapshot(t, env)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	loop, err := rest.CreateGoalLoop(ctx, env.DevOrgSlug, client.CreateGoalLoopRequest{
		Name:                 fmt.Sprintf("e2e-goal-loop-start-%d", time.Now().UnixMilli()),
		WorkerSpecSnapshotID: snapshotID,
		Objective:            "Reject a stale runtime before launch.",
		AcceptanceCriteria:   []string{"The loop remains a draft."},
		VerificationCommand:  "true",
	})
	if err != nil {
		t.Fatalf("create goal loop: %v", err)
	}
	if _, err := rest.StartGoalLoop(ctx, env.DevOrgSlug, loop.Slug); err == nil {
		t.Fatal("start goal loop unexpectedly accepted a stale worker snapshot")
	}

	db, err := client.OpenDB(env.PostgresDSN)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	var status, launchError string
	if err := db.QueryRow(ctx,
		`SELECT status, COALESCE(verification_error, '') FROM goal_loops WHERE id = $1`, loop.ID,
	).Scan(&status, &launchError); err != nil {
		t.Fatalf("load failed goal loop: %v", err)
	}
	if status != "draft" {
		t.Fatalf("goal loop status = %q, want draft", status)
	}
	if launchError != "" {
		t.Fatalf("stale snapshot validation recorded a launch error: %q", launchError)
	}
}
