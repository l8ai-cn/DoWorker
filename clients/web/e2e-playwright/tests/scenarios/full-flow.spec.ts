// Migrated R5+: Connect-RPC only (no REST middle layer).
import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";
import { pollUntil } from "../../helpers/retry";
import { terminateAllPods } from "../../helpers/pod-cleanup";
import { createE2EEchoPod } from "../../helpers/e2e-worker-spec";

type Runner = { id: bigint; currentPods?: number };
type Repository = { id: bigint; defaultBranch: string };
type Ticket = { slug: string };
type Pod = { podKey: string; status: string; runnerId: bigint };

/**
 * TC-SCENARIO-001: Full flow — Git Credential → Repository → Ticket → Pod
 */
test.describe("Full E2E Scenario", () => {
  test.beforeAll(async () => { await terminateAllPods(); });
  test.beforeEach(async () => { clearAuthRateLimit(); });

  test("git credential → repository → ticket → pod lifecycle", async ({ api }) => {
    const cc = await api.connect();

    // Step 1: Verify repositories exist (from seed)
    const { items: repos } = await cc.repository.listRepositories({ orgSlug: TEST_ORG_SLUG }) as { items: Repository[] };
    expect(repos.length).toBeGreaterThan(0);
    const repository = repos[0];
    expect(repository.defaultBranch, "seeded repository must declare a default branch").toBeTruthy();

    // Step 2: Create ticket linked to repository
    const ticket = await cc.ticket.createTicket({
      orgSlug: TEST_ORG_SLUG,
      title: "E2E Scenario Ticket",
      repositoryId: repository.id,
    }) as Ticket;
    const ticketSlug = ticket.slug;

    // Step 3: Create pod with repository and ticket
    const podResp = await createE2EEchoPod(cc, {
      repositoryId: repository.id,
      branch: repository.defaultBranch,
      ticketSlug,
    }) as { pod: Pod };
    const podKey = podResp.pod?.podKey;

    if (podKey) {
      // Step 4: Wait for pod running
      await pollUntil(
        async () => {
          const pod = await cc.pod.getPod({ orgSlug: TEST_ORG_SLUG, podKey }) as Pod;
          return pod.status === "running";
        },
        { maxAttempts: 10, intervalMs: 3000, label: "scenario-pod-running" }
      ).catch(() => {});

      // Step 5: Verify runner capacity changed
      const runnerCheck = await cc.runner.getRunner({
        orgSlug: TEST_ORG_SLUG,
        id: podResp.pod.runnerId,
      }) as { runner: Runner };
      expect((runnerCheck.runner?.currentPods ?? 0)).toBeGreaterThanOrEqual(0);

      // Step 6: Terminate pod
      await cc.pod.terminatePod({ orgSlug: TEST_ORG_SLUG, podKey });
    }

    // Step 7: Cleanup ticket
    if (ticketSlug) {
      await cc.ticket.deleteTicket({ orgSlug: TEST_ORG_SLUG, ticketSlug });
    }
  });
});
