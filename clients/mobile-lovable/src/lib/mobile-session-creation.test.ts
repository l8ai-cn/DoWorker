import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  ListWorkerCreateOptionsRequestSchema,
  ListWorkerCreateOptionsResponseSchema,
} from "@agent-cloud/proto/pod/v1/worker_creation_pb";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { apiFetch } from "./api-fetch";
import { readOrgSlug } from "./auth-store";
import { resolveDefaultModelResourceId } from "./model-resources-api";
import { getMobilePodService } from "./mobile-wasm";
import { createMobileWorkerSession } from "./mobile-session-creation";

vi.mock("./api-fetch", () => ({ apiFetch: vi.fn() }));
vi.mock("./auth-store", () => ({ readOrgSlug: vi.fn(() => "dev-org") }));
vi.mock("./model-resources-api", () => ({ resolveDefaultModelResourceId: vi.fn() }));
vi.mock("./mobile-wasm", () => ({ getMobilePodService: vi.fn() }));

const apiFetchMock = vi.mocked(apiFetch);
const readOrgSlugMock = vi.mocked(readOrgSlug);
const resolveDefaultModelResourceIdMock = vi.mocked(resolveDefaultModelResourceId);
const getMobilePodServiceMock = vi.mocked(getMobilePodService);
const optionsMock = vi.fn();

beforeEach(() => {
  apiFetchMock.mockReset();
  optionsMock.mockReset();
  readOrgSlugMock.mockReturnValue("dev-org");
  resolveDefaultModelResourceIdMock.mockReset();
  getMobilePodServiceMock.mockResolvedValue({
    list_worker_create_options_connect: optionsMock,
  } as never);
});

describe("createMobileWorkerSession", () => {
  it("sends the exact authoritative ACP plan and model resource", async () => {
    optionsMock.mockResolvedValue(optionsBytes());
    resolveDefaultModelResourceIdMock.mockResolvedValue(42);
    apiFetchMock.mockResolvedValue(
      new Response(JSON.stringify({ id: "session-1", agent_id: "agent_catalog_1", status: "launching" })),
    );

    await createMobileWorkerSession(
      {
        id: "agent_catalog_1",
        workerTypeSlug: "codex-cli",
        supportedModes: ["acp", "pty"],
        requiresModelResource: true,
      },
      "Fix CI",
      "Please investigate",
      "acp",
    );

    expect(optionsMock).toHaveBeenCalledTimes(3);
    expect(decodeOptionsRequest(0)).toMatchObject({ orgSlug: "dev-org", workerTypeSlug: "codex-cli" });
    expect(decodeOptionsRequest(1)).toMatchObject({ computeTargetId: BigInt(21) });
    expect(decodeOptionsRequest(2)).toMatchObject({ deploymentMode: "pooled" });
    expect(JSON.parse((apiFetchMock.mock.calls[0][1] as RequestInit).body as string)).toEqual({
      agent_id: "agent_catalog_1",
      title: "Fix CI",
      initial_items: [{
        type: "message",
        data: { role: "user", content: [{ type: "input_text", text: "Please investigate" }] },
      }],
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

  it("maps PTY to interactive automation without loading a model resource", async () => {
    optionsMock.mockResolvedValue(optionsBytes({ requiresModelResource: false, modes: ["pty"] }));
    apiFetchMock.mockResolvedValue(
      new Response(JSON.stringify({ id: "session-2", agent_id: "agent_catalog_2", status: "launching" })),
    );

    await createMobileWorkerSession(
      {
        id: "agent_catalog_2",
        workerTypeSlug: "aider",
        supportedModes: ["pty"],
        requiresModelResource: false,
      },
      undefined,
      undefined,
      "pty",
    );

    expect(resolveDefaultModelResourceIdMock).not.toHaveBeenCalled();
    expect(JSON.parse((apiFetchMock.mock.calls[0][1] as RequestInit).body as string)).toMatchObject({
      agent_id: "agent_catalog_2",
      initial_items: [],
      automation_level: "interactive",
    });
  });

  it("fails closed when the options revision changes before session creation", async () => {
    optionsMock
      .mockResolvedValueOnce(optionsBytes({ revision: "catalog-1", requiresModelResource: false }))
      .mockResolvedValueOnce(optionsBytes({ revision: "catalog-2", requiresModelResource: false }));

    await expect(
      createMobileWorkerSession(
        {
          id: "agent_catalog_1",
          workerTypeSlug: "codex-cli",
          supportedModes: ["acp", "pty"],
          requiresModelResource: false,
        },
        undefined,
        undefined,
        "acp",
      ),
    ).rejects.toThrow("Worker 创建选项已变化");
    expect(apiFetchMock).not.toHaveBeenCalled();
  });
});

function decodeOptionsRequest(call: number) {
  return fromBinary(
    ListWorkerCreateOptionsRequestSchema,
    new Uint8Array(optionsMock.mock.calls[call][0]),
  );
}

function optionsBytes(overrides: {
  revision?: string;
  requiresModelResource?: boolean;
  modes?: string[];
} = {}): Uint8Array {
  return toBinary(
    ListWorkerCreateOptionsResponseSchema,
    create(ListWorkerCreateOptionsResponseSchema, {
      revision: overrides.revision ?? "catalog-9",
      workerTypes: [{
        slug: overrides.modes?.[0] === "pty" ? "aider" : "codex-cli",
        selectable: true,
        supportedInteractionModes: overrides.modes ?? ["acp", "pty"],
        requiresModelResource: overrides.requiresModelResource ?? true,
      }],
      runtimeImages: [{ id: BigInt(11), selectable: true, workerTypeSlugs: ["codex-cli", "aider"] }],
      computeTargets: [{ id: BigInt(21), selectable: true }],
      deploymentModes: [{ value: "pooled", selectable: true }],
      resourceProfiles: [{ id: BigInt(31), selectable: true }],
    }),
  );
}
