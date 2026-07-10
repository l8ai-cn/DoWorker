// e2e-mock-agent only ships in the e2e-echo runner images
// (dev-runner / dev-runner-2). Other agent runtimes prune that binary,
// so picking runners[0] from ListAvailableRunners can schedule onto
// claude/gemini/… and fail with "executable file not found".

import { TEST_ORG_SLUG } from "./env";

export interface RunnerWithNode {
  id: bigint;
  nodeId?: string;
}

export const E2E_ECHO_AGENT_SLUG = "e2e-echo";

const E2E_ECHO_NODE_IDS = new Set(["dev-runner", "dev-runner-2"]);

export function pickE2EEchoRunner(runners: RunnerWithNode[] | undefined): RunnerWithNode {
  if (!runners?.length) {
    throw new Error("pickE2EEchoRunner: no online runners");
  }
  const match = runners.find((r) => r.nodeId && E2E_ECHO_NODE_IDS.has(r.nodeId));
  if (!match) {
    throw new Error(
      `pickE2EEchoRunner: no e2e-echo runner online (have: ${runners
        .map((r) => r.nodeId ?? String(r.id))
        .join(", ")})`,
    );
  }
  return match;
}

type ConnectLike = {
  runner: {
    listAvailableRunners: (req: { orgSlug: string }) => Promise<{ items?: RunnerWithNode[] }>;
  };
};

export async function resolveE2EPodCreateTargets(cc: ConnectLike): Promise<{
  runnerId: bigint;
  agentSlug: string;
}> {
  const { items: runners } = await cc.runner.listAvailableRunners({ orgSlug: TEST_ORG_SLUG });
  return {
    runnerId: pickE2EEchoRunner(runners).id,
    agentSlug: E2E_ECHO_AGENT_SLUG,
  };
}
