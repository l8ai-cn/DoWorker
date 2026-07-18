import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  CompileLoopProgramResponseSchema,
  GenerateLoopProgramRequestSchema,
  LoopDraftSnapshotSchema,
  RepairLoopProgramRequestSchema,
  RepairLoopProgramResponseSchema,
} from "@proto/goalloop/v1/goalloop_pb";

const mocks = vi.hoisted(() => ({
  generate: vi.fn(),
  repair: vi.fn(),
  applyAIDraft: vi.fn(),
  snapshot: vi.fn(),
}));

vi.mock("@/lib/wasm-core", () => ({
  initWasmCore: vi.fn().mockResolvedValue(undefined),
  getGoalLoopService: () => ({
    generateLoopProgramConnect: mocks.generate,
    repairLoopProgramConnect: mocks.repair,
  }),
  getLoopBuilderState: () => ({
    apply_ai_draft_response: mocks.applyAIDraft,
    snapshot_bytes: mocks.snapshot,
  }),
}));

import {
  applyLoopAIDraft,
  decodeLoopAIRepair,
  requestLoopAIDraft,
  requestLoopAIRepair,
} from "../connect/loopAIConnect";

describe("Loop AI Connect adapter", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    const response = create(CompileLoopProgramResponseSchema, {
      canonicalSource: "loop generated {}",
      program: { schemaVersion: 1 },
      revision: 7n,
    });
    mocks.generate.mockResolvedValue(
      toBinary(CompileLoopProgramResponseSchema, response),
    );
    mocks.snapshot.mockReturnValue(
      toBinary(
        LoopDraftSnapshotSchema,
        create(LoopDraftSnapshotSchema, {
          source: "loop generated {}",
          revision: 8n,
          semanticRevision: 3n,
          parseStatus: "valid",
          activeEditor: "blocks",
        }),
      ),
    );
  });

  it("generates from source and prompt without diagnostics or a Worker snapshot", async () => {
    await requestLoopAIDraft({
      orgSlug: "acme",
      prompt: "制作专业 PPT",
      currentSource: "loop current {}",
      modelResourceId: 42,
      locale: "zh-CN",
      revision: 7,
    });

    const request = fromBinary(
      GenerateLoopProgramRequestSchema,
      mocks.generate.mock.calls[0][0],
    );
    expect(request).toMatchObject({
      orgSlug: "acme",
      prompt: "制作专业 PPT",
      currentSource: "loop current {}",
      modelResourceId: 42n,
      locale: "zh-CN",
      revision: 7n,
    });
    expect(request).not.toHaveProperty("currentDiagnostics");
    expect(request).not.toHaveProperty("workerSpecSnapshotId");
  });

  it("lets Rust reject stale proposals without mutating the snapshot", async () => {
    mocks.applyAIDraft.mockReturnValue(false);

    const result = await applyLoopAIDraft(new Uint8Array([1, 2, 3]));

    expect(result.applied).toBe(false);
    expect(result.snapshot.source).toBe("loop generated {}");
    expect(mocks.applyAIDraft).toHaveBeenCalledWith(
      new Uint8Array([1, 2, 3]),
    );
  });

  it("serializes the selected diagnostic into a targeted repair request", async () => {
    mocks.repair.mockResolvedValue(new Uint8Array([9]));

    await requestLoopAIRepair({
      orgSlug: "acme",
      source: "loop current {}",
      modelResourceId: 42,
      locale: "zh-CN",
      revision: 7,
      diagnosticCode: "loop.value.out-of-range",
      nodeId: "n-limits",
      fieldPath: "limits.iterations",
      prompt: "保持预算严格",
    });

    const request = fromBinary(
      RepairLoopProgramRequestSchema,
      mocks.repair.mock.calls[0][0],
    );
    expect(request).toMatchObject({
      orgSlug: "acme",
      source: "loop current {}",
      modelResourceId: 42n,
      locale: "zh-CN",
      revision: 7n,
      diagnosticCode: "loop.value.out-of-range",
      nodeId: "n-limits",
      fieldPath: "limits.iterations",
      prompt: "保持预算严格",
    });
  });

  it("decodes only a same-revision repair for the requested field", () => {
    const response = create(RepairLoopProgramResponseSchema, {
      proposal: {
        canonicalSource: "loop repaired {}",
        program: { schemaVersion: 1 },
        revision: 7n,
      },
      patch: {
        nodeId: "n-limits",
        fieldPath: "limits.iterations",
        oldValue: 100n,
        newValue: 20n,
      },
    });
    const bytes = toBinary(RepairLoopProgramResponseSchema, response);

    const decoded = decodeLoopAIRepair(bytes, {
      revision: 7,
      nodeId: "n-limits",
      fieldPath: "limits.iterations",
    });

    expect(decoded.proposal.canonicalSource).toBe("loop repaired {}");
    expect(decoded.patch.newValue).toBe(20n);
    expect(fromBinary(CompileLoopProgramResponseSchema, decoded.proposalBytes))
      .toEqual(decoded.proposal);

    expect(() => decodeLoopAIRepair(bytes, {
      revision: 8,
      nodeId: "n-limits",
      fieldPath: "limits.iterations",
    })).toThrow("stale");
    expect(() => decodeLoopAIRepair(bytes, {
      revision: 7,
      nodeId: "n-limits",
      fieldPath: "limits.tokens",
    })).toThrow("target");
  });
});
