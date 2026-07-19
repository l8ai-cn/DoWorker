import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/env", () => ({
  getApiBaseUrl: () => "http://localhost:10000/api",
}));

vi.mock("@/lib/wasm-core", () => ({
  getAuthManager: () => ({ get_token: () => "test-token" }),
}));

vi.mock("@/stores/auth", () => ({
  readCurrentOrg: () => ({ slug: "acme" }),
}));

vi.mock("../sessionImportWorkerPlan", () => ({
  buildSessionImportWorkerPlan: vi.fn(),
}));

import { buildSessionImportWorkerPlan } from "../sessionImportWorkerPlan";
import { fetchSessionByPodKey, importCodexSession } from "../sessionImportApi";

describe("fetchSessionByPodKey", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.mocked(buildSessionImportWorkerPlan).mockReset();
  });

  it("returns null only for an absent session association", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(null, { status: 204 }),
    );

    await expect(fetchSessionByPodKey("worker-pod")).resolves.toBeNull();
  });

  it("returns the associated session metadata", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: "conv_123", title: "Imported" }), {
        status: 200,
      }),
    );

    await expect(fetchSessionByPodKey("worker-pod")).resolves.toEqual({
      id: "conv_123",
      title: "Imported",
    });
  });

  it("surfaces server failures instead of treating them as absence", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response("database unavailable", { status: 500 }),
    );

    await expect(fetchSessionByPodKey("worker-pod")).rejects.toThrow(
      "database unavailable",
    );
  });

  it("imports only with an authoritative ACP Worker plan", async () => {
    vi.mocked(buildSessionImportWorkerPlan).mockResolvedValue({
      worker_spec: {
        options_revision: "catalog-9",
        runtime_image_id: 11,
        placement_policy: "automatic",
        compute_target_id: 21,
        deployment_mode: "pooled",
        resource_profile_id: 31,
      },
      automation_level: "autonomous",
      model_resource_id: 42,
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          session: { id: "session_1", agent_id: "codex-cli" },
          pod_key: "pod_1",
          source_kind: "rollout",
          source_id: "source_1",
          item_count: 2,
        }),
        { status: 200 },
      ),
    );

    await importCodexSession("/tmp/rollout", "codex-cli", { modelResourceId: 42 });

    expect(buildSessionImportWorkerPlan).toHaveBeenCalledWith({
      orgSlug: "acme",
      workerTypeSlug: "codex-cli",
      modelResourceId: 42,
    });
    expect(JSON.parse((fetchMock.mock.calls[0][1] as RequestInit).body as string)).toEqual({
      source_path: "/tmp/rollout",
      agent_id: "codex-cli",
      worker_spec: {
        options_revision: "catalog-9",
        runtime_image_id: 11,
        placement_policy: "automatic",
        compute_target_id: 21,
        deployment_mode: "pooled",
        resource_profile_id: 31,
      },
      automation_level: "autonomous",
      model_resource_id: 42,
    });
  });
});
