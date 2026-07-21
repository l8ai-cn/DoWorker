import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  ListWorkerCreateOptionsRequestSchema,
  ListWorkerCreateOptionsResponseSchema,
} from "@agent-cloud/proto/pod/v1/worker_creation_pb";

vi.mock("./identity", () => ({ authenticatedFetch: vi.fn() }));
vi.mock("./agent-cloud", () => ({ readAgentCloudOrgSlug: vi.fn(() => "dev-org") }));

import { authenticatedFetch } from "./identity";
import { buildSessionWorkerPlan, workerRequiresModelResource } from "./workerSessionPlan";

const fetchMock = vi.mocked(authenticatedFetch);
const selection = {
  workerTypeSlug: "codex-cli",
  supportedModes: ["acp", "pty"] as const,
  requiresModelResource: false,
};

beforeEach(() => {
  fetchMock.mockReset();
});

describe("buildSessionWorkerPlan", () => {
  it("uses authoritative Connect options and returns the exact REST session plan", async () => {
    fetchMock.mockImplementation(async () => workerOptionsResponse({ requiresModelResource: true }));

    await expect(
      buildSessionWorkerPlan({
        selection: { ...selection, requiresModelResource: true },
        mode: "acp",
        modelResourceId: 42,
      }),
    ).resolves.toEqual({
      worker_spec: {
        options_revision: "runtime-catalog-9",
        runtime_image_id: 11,
        placement_policy: "automatic",
        compute_target_id: 21,
        deployment_mode: "pooled",
        resource_profile_id: 31,
      },
      automation_level: "autonomous",
      model_resource_id: 42,
    });

    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(decodeRequest(0)).toMatchObject({ orgSlug: "dev-org", workerTypeSlug: "codex-cli" });
    expect(decodeRequest(1)).toMatchObject({
      orgSlug: "dev-org",
      workerTypeSlug: "codex-cli",
      computeTargetId: BigInt(21),
    });
    expect(decodeRequest(2)).toMatchObject({
      orgSlug: "dev-org",
      workerTypeSlug: "codex-cli",
      computeTargetId: BigInt(21),
      deploymentMode: "pooled",
    });
    expect(new Headers(fetchMock.mock.calls[0]?.[1]?.headers).get("Content-Type")).toBe(
      "application/proto",
    );
  });

  it("blocks when the selected Worker has no selectable runtime", async () => {
    fetchMock.mockImplementation(async () =>
      workerOptionsResponse({ runtimeImages: [{ id: BigInt(11), selectable: false }] }),
    );

    await expect(
      buildSessionWorkerPlan({ selection, mode: "pty" }),
    ).rejects.toThrow("no selectable runtime image");
  });

  it("reads the model-resource requirement from the selectable Worker option", async () => {
    fetchMock.mockImplementation(async () => workerOptionsResponse({ requiresModelResource: true }));

    await expect(
      workerRequiresModelResource({ ...selection, requiresModelResource: true }),
    ).resolves.toBe(true);

    fetchMock.mockImplementation(async () => workerOptionsResponse({ workerSelectable: false }));
    await expect(workerRequiresModelResource(selection)).rejects.toThrow("is not selectable");
  });

  it.each([
    ["worker", { workerSelectable: false }, "is not selectable"],
    ["compute target", { computeSelectable: false }, "no selectable compute target"],
    ["deployment mode", { deploymentSelectable: false }, "no selectable deployment mode"],
    ["resource profile", { resourceProfileSelectable: false }, "no selectable resource profile"],
  ])("blocks when no selectable %s exists", async (_name, options, message) => {
    fetchMock.mockImplementation(async () => workerOptionsResponse(options));

    await expect(
      buildSessionWorkerPlan({ selection, mode: "acp" }),
    ).rejects.toThrow(message);
  });

  it("blocks unsupported interaction modes before making a partial plan", async () => {
    fetchMock.mockImplementation(async () => workerOptionsResponse({ modes: ["pty"] }));

    await expect(
      buildSessionWorkerPlan({
        selection: { ...selection, supportedModes: ["pty"] },
        mode: "acp",
      }),
    ).rejects.toThrow("does not support acp sessions");
  });

  it("requires a valid model resource only for Worker types that require one", async () => {
    fetchMock.mockImplementation(async () => workerOptionsResponse({ requiresModelResource: true }));

    await expect(
      buildSessionWorkerPlan({
        selection: { ...selection, requiresModelResource: true },
        mode: "acp",
      }),
    ).rejects.toThrow("requires a model resource");
    await expect(
      buildSessionWorkerPlan({
        selection: { ...selection, requiresModelResource: true },
        mode: "acp",
        modelResourceId: 0,
      }),
    ).rejects.toThrow("Invalid model resource id");
  });

  it("rejects a model resource for a Worker type that does not use one", async () => {
    fetchMock.mockImplementation(async () => workerOptionsResponse());

    await expect(
      buildSessionWorkerPlan({ selection, mode: "acp", modelResourceId: 42 }),
    ).rejects.toThrow("does not accept a model resource");
  });

  it("fails instead of combining options from different catalog revisions", async () => {
    fetchMock
      .mockResolvedValueOnce(workerOptionsResponse({ revision: "catalog-1" }))
      .mockResolvedValueOnce(workerOptionsResponse({ revision: "catalog-2" }));

    await expect(
      buildSessionWorkerPlan({ selection, mode: "pty" }),
    ).rejects.toThrow("options changed");
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("rejects a Worker when catalog metadata no longer matches authoritative options", async () => {
    fetchMock.mockImplementation(async () => workerOptionsResponse({ requiresModelResource: true }));

    await expect(
      buildSessionWorkerPlan({ selection, mode: "acp", modelResourceId: 42 }),
    ).rejects.toThrow("catalog metadata changed");
  });
});

function decodeRequest(call: number) {
  const init = fetchMock.mock.calls[call]?.[1] as RequestInit;
  return fromBinary(
    ListWorkerCreateOptionsRequestSchema,
    new Uint8Array(init.body as Uint8Array),
  );
}

function workerOptionsResponse(overrides: {
  runtimeImages?: Array<{ id: bigint; selectable: boolean }>;
  revision?: string;
  workerSelectable?: boolean;
  computeSelectable?: boolean;
  deploymentSelectable?: boolean;
  resourceProfileSelectable?: boolean;
  requiresModelResource?: boolean;
  modes?: string[];
} = {}): Response {
  const bytes = toBinary(
    ListWorkerCreateOptionsResponseSchema,
    create(ListWorkerCreateOptionsResponseSchema, {
      revision: overrides.revision ?? "runtime-catalog-9",
      workerTypes: [{
        slug: "codex-cli",
        selectable: overrides.workerSelectable ?? true,
        supportedInteractionModes: overrides.modes ?? ["acp", "pty"],
        requiresModelResource: overrides.requiresModelResource ?? false,
      }],
      runtimeImages: (overrides.runtimeImages ?? [{ id: BigInt(11), selectable: true }]).map(
        (image) => ({
          ...image,
          workerTypeSlugs: ["codex-cli"],
        }),
      ),
      computeTargets: [{ id: BigInt(21), selectable: overrides.computeSelectable ?? true }],
      deploymentModes: [{ value: "pooled", selectable: overrides.deploymentSelectable ?? true }],
      resourceProfiles: [{ id: BigInt(31), selectable: overrides.resourceProfileSelectable ?? true }],
    }),
  );
  return new Response(bytes, { status: 200 });
}
