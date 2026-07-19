import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./modelConfigsApi", () => ({ listModelResources: vi.fn() }));
vi.mock("./workerSessionPlan", () => ({ buildSessionWorkerPlan: vi.fn() }));

import { listModelResources } from "./modelConfigsApi";
import { buildSessionWorkerPlan } from "./workerSessionPlan";
import {
  createWorkerSessionBody,
  crossAgentWorkerSessionBody,
  importWorkerSessionBody,
} from "./workerSessionRequestBodies";

const plan = {
  worker_spec: {
    options_revision: "catalog-9",
    runtime_image_id: 11,
    placement_policy: "automatic" as const,
    compute_target_id: 21,
    deployment_mode: "pooled",
    resource_profile_id: 31,
  },
  automation_level: "autonomous" as const,
  model_resource_id: 42,
};
const selection = {
  workerTypeSlug: "codex-cli",
  supportedModes: ["acp", "pty"] as const,
  requiresModelResource: true,
};

beforeEach(() => {
  vi.mocked(buildSessionWorkerPlan).mockReset();
  vi.mocked(listModelResources).mockReset();
  vi.mocked(buildSessionWorkerPlan).mockResolvedValue(plan);
});

describe("worker session request bodies", () => {
  it("creates and imports with the exact authoritative plan", async () => {
    await expect(
      createWorkerSessionBody({
        agentId: "codex-cli",
        ...selection,
        initialItems: [],
        hostId: "host_1",
        workspace: "/workspace/project",
        parentSessionId: "conv_parent",
        subAgentName: null,
      }),
    ).resolves.toEqual({
      agent_id: "codex-cli",
      initial_items: [],
      host_id: "host_1",
      workspace: "/workspace/project",
      parent_session_id: "conv_parent",
      sub_agent_name: null,
      ...plan,
    });
    await expect(
      importWorkerSessionBody({
        agentId: "codex-cli",
        ...selection,
        sourcePath: "/tmp/rollout.jsonl",
        title: "Imported",
        hostId: "host_a",
      }),
    ).resolves.toEqual({
      source_path: "/tmp/rollout.jsonl",
      agent_id: "codex-cli",
      title: "Imported",
      host_id: "host_a",
      ...plan,
    });
  });

  it("uses autonomous for ACP and interactive for PTY through the plan builder", async () => {
    await createWorkerSessionBody({ agentId: "codex-cli", initialItems: [], mode: "pty", ...selection });
    expect(buildSessionWorkerPlan).toHaveBeenCalledWith({
      selection,
      mode: "pty",
      modelResourceId: undefined,
      resolveModelResourceId: expect.any(Function),
    });
  });

  it("keeps a same-agent fork as an empty configuration snapshot operation", async () => {
    await expect(
      crossAgentWorkerSessionBody({
        agentId: "codex-cli",
        ...selection,
        sourceAgentId: "codex-cli",
        title: "Fork",
      }),
    ).resolves.toEqual({ agent_id: "codex-cli", title: "Fork" });
    expect(buildSessionWorkerPlan).not.toHaveBeenCalled();
  });

  it("adds the full plan only for a cross-agent operation", async () => {
    await expect(
      crossAgentWorkerSessionBody({
        agentId: "agent_entity_42",
        ...selection,
        workerTypeSlug: "claude-code",
        sourceAgentId: "codex-cli",
        upToResponseId: "resp_1",
      }),
    ).resolves.toEqual({ agent_id: "agent_entity_42", up_to_response_id: "resp_1", ...plan });
    expect(buildSessionWorkerPlan).toHaveBeenCalledWith({
      selection: {
        ...selection,
        workerTypeSlug: "claude-code",
      },
      mode: "acp",
      modelResourceId: undefined,
      resolveModelResourceId: expect.any(Function),
    });
  });

  it("requires one valid default resource when a required-worker plan resolves it", async () => {
    vi.mocked(listModelResources).mockResolvedValue([
      { id: 42, is_default: true, name: "GPT", provider_key: "openai", model: "gpt-5" },
    ]);
    const body = await createWorkerSessionBody({ agentId: "codex-cli", initialItems: [], ...selection });
    const resolver = vi.mocked(buildSessionWorkerPlan).mock.calls[0]?.[0].resolveModelResourceId;
    await expect(resolver?.()).resolves.toBe(42);
    expect(body).toMatchObject(plan);
  });

  it.each([
    [[], "No default model resource is configured"],
    [[
      { id: 1, is_default: true, name: "A", provider_key: "a", model: "a" },
      { id: 2, is_default: true, name: "B", provider_key: "b", model: "b" },
    ], "No default model resource is configured"],
    [[{ id: 0, is_default: true, name: "A", provider_key: "a", model: "a" }], "Invalid model resource id"],
  ])("blocks invalid default model resources", async (resources, message) => {
    vi.mocked(listModelResources).mockResolvedValue(resources);
    await createWorkerSessionBody({ agentId: "codex-cli", initialItems: [], ...selection });
    const resolver = vi.mocked(buildSessionWorkerPlan).mock.calls[0]?.[0].resolveModelResourceId;
    await expect(resolver?.()).rejects.toThrow(message);
  });

});
