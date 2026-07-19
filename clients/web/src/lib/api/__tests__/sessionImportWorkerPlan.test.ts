import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/api/facade/podConnect", () => ({ listWorkerCreateOptions: vi.fn() }));

import { listWorkerCreateOptions } from "@/lib/api/facade/podConnect";
import {
  buildSessionImportWorkerPlan,
  getSessionImportWorkerRequirement,
} from "../sessionImportWorkerPlan";

const optionsMock = vi.mocked(listWorkerCreateOptions);

beforeEach(() => {
  optionsMock.mockReset();
  optionsMock.mockResolvedValue(options());
});

describe("session import Worker plan", () => {
  it("builds the exact ACP-only import plan from authoritative options", async () => {
    await expect(
      buildSessionImportWorkerPlan({
        orgSlug: "dev-org",
        workerTypeSlug: "codex-cli",
        modelResourceId: 42,
      }),
    ).resolves.toEqual({
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
    expect(optionsMock).toHaveBeenNthCalledWith(1, "dev-org", {
      worker_type_slug: "codex-cli",
    });
    expect(optionsMock).toHaveBeenNthCalledWith(2, "dev-org", {
      worker_type_slug: "codex-cli",
      compute_target_id: 21,
    });
    expect(optionsMock).toHaveBeenNthCalledWith(3, "dev-org", {
      worker_type_slug: "codex-cli",
      compute_target_id: 21,
      deployment_mode: "pooled",
    });
  });

  it("rejects a non-ACP Worker and never creates a partial plan", async () => {
    optionsMock.mockResolvedValue(options({ modes: ["pty"] }));

    await expect(
      buildSessionImportWorkerPlan({
        orgSlug: "dev-org",
        workerTypeSlug: "codex-cli",
      }),
    ).rejects.toThrow("不支持 ACP");
    expect(optionsMock).toHaveBeenCalledTimes(1);
  });

  it("rejects mixed option revisions and missing required model resource", async () => {
    optionsMock
      .mockResolvedValueOnce(options({ revision: "catalog-1", requiresModelResource: false }))
      .mockResolvedValueOnce(options({ revision: "catalog-2", requiresModelResource: false }));

    await expect(
      buildSessionImportWorkerPlan({
        orgSlug: "dev-org",
        workerTypeSlug: "codex-cli",
      }),
    ).rejects.toThrow("已变化");

    optionsMock.mockReset();
    optionsMock.mockResolvedValue(options({ requiresModelResource: true }));
    await expect(
      buildSessionImportWorkerPlan({
        orgSlug: "dev-org",
        workerTypeSlug: "codex-cli",
      }),
    ).rejects.toThrow("需要明确选择模型资源");
  });

  it("uses the Worker option requirement instead of agent-name heuristics", async () => {
    optionsMock.mockResolvedValue(options({
      requiresModelResource: true,
      adapters: ["openai-compatible"],
    }));

    await expect(
      getSessionImportWorkerRequirement("dev-org", "codex-cli"),
    ).resolves.toEqual({
      requiresModelResource: true,
      modelProtocolAdapters: ["openai-compatible"],
    });
  });
});

function options(overrides: {
  revision?: string;
  modes?: string[];
  requiresModelResource?: boolean;
  adapters?: string[];
} = {}) {
  return {
    revision: overrides.revision ?? "catalog-9",
    worker_types: [{
      slug: "codex-cli",
      selectable: true,
      supported_interaction_modes: overrides.modes ?? ["acp", "pty"],
      requires_model_resource: overrides.requiresModelResource ?? true,
      model_protocol_adapters: overrides.adapters ?? [],
    }],
    runtime_images: [{
      id: 11,
      selectable: true,
      worker_type_slugs: ["codex-cli"],
    }],
    compute_targets: [{ id: 21, selectable: true }],
    deployment_modes: [{ value: "pooled", selectable: true }],
    resource_profiles: [{ id: 31, selectable: true }],
  };
}
