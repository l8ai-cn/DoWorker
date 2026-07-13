import { describe, expect, it } from "vitest";

import { compileLoop } from "./compile-loop";
import type { LoopDraft } from "./loop-types";

const validDraft = {
  name: "Ship compiler",
  rootBlockId: "root",
  worker: {
    blockId: "worker",
    value: { snapshotId: 42, label: "Codex worker" },
  },
  instructions: [
    { blockId: "task-1", value: "Implement the compiler." },
    { blockId: "task-2", value: "Run the verification." },
  ],
  acceptanceCriteria: [
    { blockId: "acceptance-1", value: "All tests pass." },
  ],
  verification: { blockId: "verification", value: "pnpm test" },
  limits: {
    blockId: "limits",
    value: {
      maxIterations: 10,
      tokenBudget: 20_000,
      timeoutMinutes: 60,
      noProgressLimit: 3,
      sameErrorLimit: 2,
    },
  },
  escalationPolicy: { blockId: "escalation", value: "pause" as const },
  looseBlockIds: [],
  unknownBlockTypes: [],
  adapterDiagnostics: [],
} satisfies LoopDraft;

describe("compileLoop", () => {
  it("compiles a complete draft into canonical JSON", () => {
    const result = compileLoop(validDraft);

    expect(result.diagnostics).toEqual([]);
    expect(result.program).toEqual({
      kind: "goal-loop-program",
      schema_version: 1,
      name: "Ship compiler",
      worker: { snapshot_id: 42, label: "Codex worker" },
      objective: "Implement the compiler.\nRun the verification.",
      acceptance_criteria: ["All tests pass."],
      verification: { kind: "command", command: "pnpm test" },
      limits: {
        max_iterations: 10,
        token_budget: 20_000,
        timeout_minutes: 60,
        no_progress_limit: 3,
        same_error_limit: 2,
      },
      escalation_policy: "pause",
    });
  });

  it.each([
    ["worker", { worker: undefined }, "missing-worker"],
    ["task", { instructions: [] }, "missing-instructions"],
    ["acceptance", { acceptanceCriteria: [] }, "missing-acceptance-criteria"],
    ["verification", { verification: undefined }, "missing-verification"],
    ["limits", { limits: undefined }, "missing-limits"],
    [
      "failure handling",
      { escalationPolicy: undefined },
      "missing-escalation-policy",
    ],
  ])("rejects a draft missing %s", (_label, override, code) => {
    const result = compileLoop({ ...validDraft, ...override });

    expect(result.program).toBeUndefined();
    expect(result.diagnostics).toContainEqual(
      expect.objectContaining({
        code,
        message: expect.any(String),
        blockId: "root",
        slot: expect.any(String),
      }),
    );
  });

  it("rejects loose blocks without silently filtering them", () => {
    const result = compileLoop({
      ...validDraft,
      looseBlockIds: ["loose-1", "loose-2"],
    });

    expect(result.program).toBeUndefined();
    expect(result.diagnostics).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ code: "loose-block", blockId: "loose-1" }),
        expect.objectContaining({ code: "loose-block", blockId: "loose-2" }),
      ]),
    );
  });

  it("rejects unknown block types", () => {
    const result = compileLoop({
      ...validDraft,
      unknownBlockTypes: [{ blockId: "mystery", type: "loom_magic" }],
    });

    expect(result.program).toBeUndefined();
    expect(result.diagnostics).toContainEqual(
      expect.objectContaining({ code: "unknown-block-type", blockId: "mystery" }),
    );
  });

  it.each([
    ["maxIterations", 0],
    ["maxIterations", 101],
    ["tokenBudget", undefined],
    ["tokenBudget", 0],
    ["timeoutMinutes", -1],
    ["noProgressLimit", 0],
    ["sameErrorLimit", 0],
  ])("rejects invalid %s boundary", (field, value) => {
    const result = compileLoop({
      ...validDraft,
      limits: {
        ...validDraft.limits,
        value: { ...validDraft.limits.value, [field]: value },
      },
    });

    expect(result.program).toBeUndefined();
    expect(result.diagnostics).toContainEqual(
      expect.objectContaining({
        code: "invalid-limits",
        blockId: validDraft.limits.blockId,
      }),
    );
  });

  it("returns execution block ids in execution order", () => {
    expect(compileLoop(validDraft).executionBlockIds).toEqual([
      "root",
      "worker",
      "limits",
      "escalation",
      "task-1",
      "task-2",
      "acceptance-1",
      "verification",
    ]);
  });

  it.each([
    ["name", { name: "  " }, "missing-name"],
    [
      "instruction text",
      { instructions: [{ blockId: "task-1", value: "  " }] },
      "empty-instruction",
    ],
    [
      "acceptance text",
      { acceptanceCriteria: [{ blockId: "acceptance-1", value: "" }] },
      "empty-acceptance-criterion",
    ],
    [
      "verification command",
      { verification: { blockId: "verification", value: "\n" } },
      "empty-verification",
    ],
  ])("rejects blank %s", (_label, override, code) => {
    const result = compileLoop({ ...validDraft, ...override });

    expect(result.program).toBeUndefined();
    expect(result.diagnostics).toContainEqual(
      expect.objectContaining({ code }),
    );
  });

  it.each([
    ["zero", 0],
    ["NaN", Number.NaN],
    ["fraction", 1.5],
  ])("rejects %s worker snapshot ids", (_label, snapshotId) => {
    const result = compileLoop({
      ...validDraft,
      worker: {
        blockId: "worker",
        value: { snapshotId, label: "Codex worker" },
      },
    });

    expect(result.program).toBeUndefined();
    expect(result.diagnostics).toContainEqual(
      expect.objectContaining({
        code: "invalid-worker-snapshot",
        blockId: "worker",
      }),
    );
  });

  it("rejects limits beyond the safe integer range", () => {
    const result = compileLoop({
      ...validDraft,
      limits: {
        ...validDraft.limits,
        value: {
          ...validDraft.limits.value,
          tokenBudget: Number.MAX_SAFE_INTEGER + 1,
        },
      },
    });

    expect(result.program).toBeUndefined();
    expect(result.diagnostics).toContainEqual(
      expect.objectContaining({ code: "invalid-limits" }),
    );
  });

  it("returns diagnostics instead of throwing on malformed text values", () => {
    const malformed = {
      ...validDraft,
      instructions: [{ blockId: "task-1", value: 7 }],
    } as unknown as LoopDraft;

    expect(() => compileLoop(malformed)).not.toThrow();
    expect(compileLoop(malformed).diagnostics).toContainEqual(
      expect.objectContaining({
        code: "invalid-instruction",
        blockId: "task-1",
      }),
    );
  });

  it("rejects escalation policies outside the published contract", () => {
    const malformed = {
      ...validDraft,
      escalationPolicy: { blockId: "escalation", value: "continue" },
    } as unknown as LoopDraft;
    const result = compileLoop(malformed);

    expect(result.program).toBeUndefined();
    expect(result.diagnostics).toContainEqual(
      expect.objectContaining({
        code: "invalid-escalation-policy",
        blockId: "escalation",
      }),
    );
  });

  it("blocks adapter diagnostics from custom Blockly fields", () => {
    const result = compileLoop({
      ...validDraft,
      adapterDiagnostics: [{
        code: "missing-custom-parameter",
        message: "自定义积木参数 command 不能为空。",
        blockId: "custom-task",
        slot: "command",
      }],
    });

    expect(result.program).toBeUndefined();
    expect(result.diagnostics).toContainEqual(
      expect.objectContaining({ code: "missing-custom-parameter" }),
    );
  });

  it("returns diagnostics for a null draft without throwing", () => {
    const malformed = null as unknown as LoopDraft;

    expect(() => compileLoop(malformed)).not.toThrow();
    expect(compileLoop(malformed)).toEqual({
      diagnostics: [
        expect.objectContaining({ code: "invalid-draft" }),
      ],
      executionBlockIds: [],
    });
  });

  it("rejects malformed block arrays without throwing", () => {
    const malformed = {
      ...validDraft,
      instructions: [null],
    } as unknown as LoopDraft;

    expect(() => compileLoop(malformed)).not.toThrow();
    expect(compileLoop(malformed).diagnostics).toContainEqual(
      expect.objectContaining({ code: "invalid-instruction" }),
    );
    expect(compileLoop(malformed).executionBlockIds.every(
      (blockId) => typeof blockId === "string",
    )).toBe(true);
  });

  it.each([
    [
      "root block id",
      { rootBlockId: 5 },
      "invalid-root-block-id",
    ],
    [
      "loose block metadata",
      { looseBlockIds: undefined },
      "invalid-loose-block-list",
    ],
    [
      "unknown block metadata",
      { unknownBlockTypes: undefined },
      "invalid-unknown-block-list",
    ],
    [
      "instruction block id",
      { instructions: [{ blockId: 9, value: "Run tests" }] },
      "invalid-instruction-block-id",
    ],
    [
      "worker block id",
      {
        worker: {
          blockId: " ",
          value: { snapshotId: 42, label: "Codex worker" },
        },
      },
      "invalid-worker-snapshot",
    ],
    [
      "verification block id",
      { verification: { blockId: "", value: "pnpm test" } },
      "invalid-verification",
    ],
    [
      "limits block id",
      { limits: { ...validDraft.limits, blockId: "\n" } },
      "invalid-limits",
    ],
    [
      "escalation block id",
      { escalationPolicy: { blockId: " ", value: "pause" } },
      "invalid-escalation-policy",
    ],
    [
      "unknown block id",
      { unknownBlockTypes: [{ blockId: "", type: "future-block" }] },
      "invalid-unknown-block-entry",
    ],
    [
      "unknown block type",
      { unknownBlockTypes: [{ blockId: "future", type: " " }] },
      "invalid-unknown-block-entry",
    ],
  ])("rejects invalid %s", (_label, override, code) => {
    const malformed = {
      ...validDraft,
      ...override,
    } as unknown as LoopDraft;

    expect(compileLoop(malformed).diagnostics).toContainEqual(
      expect.objectContaining({ code }),
    );
  });
});
