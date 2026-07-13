import { describe, expect, it } from "vitest";

import { compileLoop } from "./compile-loop";

const validDraft = {
  name: "Ship compiler",
  rootBlockId: "root",
  worker: {
    blockId: "worker",
    value: { snapshotId: "snapshot-42", label: "Codex worker" },
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
};

describe("compileLoop", () => {
  it("compiles a complete draft into canonical JSON", () => {
    const result = compileLoop(validDraft);

    expect(result.diagnostics).toEqual([]);
    expect(result.program).toEqual({
      kind: "goal-loop-program",
      schema_version: 1,
      name: "Ship compiler",
      worker: { snapshot_id: "snapshot-42", label: "Codex worker" },
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
  ])("rejects a draft missing %s", (_label, override, code) => {
    const result = compileLoop({ ...validDraft, ...override });

    expect(result.program).toBeUndefined();
    expect(result.diagnostics).toContainEqual(
      expect.objectContaining({ code, message: expect.any(String) }),
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
      "task-1",
      "task-2",
      "acceptance-1",
      "verification",
      "limits",
      "escalation",
    ]);
  });
});
