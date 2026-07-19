import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./identity", () => ({ authenticatedFetch: vi.fn() }));
vi.mock("./workerSessionRequestBodies", () => ({
  createWorkerSessionBody: vi.fn(),
  crossAgentWorkerSessionBody: vi.fn(),
  importWorkerSessionBody: vi.fn(),
}));

import { authenticatedFetch } from "./identity";
import {
  createWorkerSessionBody,
  crossAgentWorkerSessionBody,
  importWorkerSessionBody,
} from "./workerSessionRequestBodies";
import {
  createWorkerSession,
  forkSnapshotSession,
  forkWorkerSession,
  importWorkerSession,
  switchWorkerSessionAgent,
} from "./workerSessionMutations";

const fetchMock = vi.mocked(authenticatedFetch);
const selection = {
  workerTypeSlug: "codex-cli",
  supportedModes: ["acp", "pty"] as const,
  requiresModelResource: true,
};

beforeEach(() => {
  fetchMock.mockReset();
  vi.mocked(createWorkerSessionBody).mockReset();
  vi.mocked(crossAgentWorkerSessionBody).mockReset();
  vi.mocked(importWorkerSessionBody).mockReset();
  fetchMock.mockImplementation(
    async () => new Response(JSON.stringify({ id: "conv_1" }), { status: 200 }),
  );
});

describe("worker session mutations", () => {
  it("creates and imports via their plan request bodies", async () => {
    vi.mocked(createWorkerSessionBody).mockResolvedValue({ agent_id: "codex-cli" } as never);
    vi.mocked(importWorkerSessionBody).mockResolvedValue({ agent_id: "codex-cli" } as never);

    await createWorkerSession({ agentId: "codex-cli", initialItems: [], ...selection });
    await importWorkerSession({ agentId: "codex-cli", sourcePath: "/tmp/rollout.jsonl", ...selection });

    expect(fetchMock.mock.calls.map(([path, init]) => [path, JSON.parse(init?.body as string)])).toEqual([
      ["/v1/sessions", { agent_id: "codex-cli" }],
      ["/v1/sessions/import", { agent_id: "codex-cli" }],
    ]);
  });

  it("uses the cross-agent plan for fork and switch", async () => {
    vi.mocked(crossAgentWorkerSessionBody).mockResolvedValue({ agent_id: "claude-code" } as never);

    await forkWorkerSession({
      sourceId: "conv source",
      sourceAgentId: "codex-cli",
      agentId: "claude-code",
      ...selection,
      workerTypeSlug: "claude-code",
    });
    await switchWorkerSessionAgent({
      sessionId: "conv source",
      sourceAgentId: "codex-cli",
      agentId: "claude-code",
      ...selection,
      workerTypeSlug: "claude-code",
    });

    expect(fetchMock.mock.calls.map(([path]) => path)).toEqual([
      "/v1/sessions/conv%20source/fork",
      "/v1/sessions/conv%20source/switch-agent",
    ]);
  });

  it("uses an empty snapshot body for a same-agent fork", async () => {
    await forkSnapshotSession({
      sourceId: "conv source",
      title: "Fork",
      upToResponseId: "response_1",
    });

    expect(fetchMock).toHaveBeenCalledWith("/v1/sessions/conv%20source/fork", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title: "Fork", up_to_response_id: "response_1" }),
    });
  });

  it("rejects HTTP failures and malformed successful responses", async () => {
    vi.mocked(createWorkerSessionBody).mockResolvedValue({ agent_id: "codex-cli" } as never);
    fetchMock.mockResolvedValueOnce(new Response("nope", { status: 422 }));
    await expect(createWorkerSession({ agentId: "codex-cli", initialItems: [], ...selection })).rejects.toThrow("nope");

    fetchMock.mockResolvedValueOnce(new Response(JSON.stringify({}), { status: 200 }));
    await expect(createWorkerSession({ agentId: "codex-cli", initialItems: [], ...selection })).rejects.toThrow("missing an id");
  });

  it("uses the server's structured error text instead of rendering JSON", async () => {
    vi.mocked(createWorkerSessionBody).mockResolvedValue({ agent_id: "codex-cli" } as never);
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ code: "internal_error", error: "failed to create session" }), {
        status: 500,
      }),
    );

    await expect(createWorkerSession({ agentId: "codex-cli", initialItems: [], ...selection })).rejects.toThrow(
      "failed to create session",
    );
  });
});
