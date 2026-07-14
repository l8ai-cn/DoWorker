import { describe, expect, it, vi } from "vitest";

import type { GoalLoopProgram } from "../domain/loop-types";
import { runSimulation } from "./run-simulation";

const program: GoalLoopProgram = {
  kind: "goal-loop-program",
  schema_version: 1,
  name: "Test",
  worker: { snapshot_id: 42, label: "Codex" },
  objective: "Run task",
  acceptance_criteria: ["Tests pass"],
  verification: { kind: "command", command: "pnpm test" },
  limits: {
    max_iterations: 3,
    token_budget: 10000,
    timeout_minutes: 60,
    no_progress_limit: 2,
    same_error_limit: 2,
  },
  escalation_policy: "pause",
};
const blockIds = ["root", "worker", "limits", "escalation", "task", "verify"];

function options(verificationSequence: boolean[]) {
  const evidence: { kind: string; message: string }[] = [];
  return {
    evidence,
    options: {
      delayMs: 0,
      signal: new AbortController().signal,
      verificationSequence,
      onEvidence: (event: { kind: string; message: string }) =>
        evidence.push(event),
      onHighlight: vi.fn(),
    },
  };
}

describe("runSimulation", () => {
  it("completes only after the simulated verifier passes", async () => {
    const run = options([true]);
    const result = await runSimulation(program, blockIds, run.options);

    expect(result).toEqual({ status: "completed", iterations: 1 });
    expect(run.evidence.some(({ kind }) => kind === "complete")).toBe(true);
    expect(run.evidence.map(({ message }) => message).join(" ")).not.toContain(
      "可信成功证据",
    );
  });

  it("pauses when repeated failures reach the no-progress limit", async () => {
    const run = options([false, false, false]);
    const result = await runSimulation(program, blockIds, run.options);

    expect(result).toEqual({
      status: "paused",
      iterations: 2,
      reason: "no_progress",
    });
    expect(run.evidence.at(-1)?.kind).toBe("paused");
  });

  it("fails at max iterations when escalation is fail", async () => {
    const run = options([false, false, false]);
    const result = await runSimulation({
      ...program,
      limits: {
        ...program.limits,
        max_iterations: 2,
        no_progress_limit: 9,
        same_error_limit: 9,
      },
      escalation_policy: "fail",
    }, blockIds, run.options);

    expect(result).toEqual({
      status: "failed",
      iterations: 2,
      reason: "max_iterations",
    });
  });

  it("tracks repeated errors separately from progress", async () => {
    const run = options([]);
    const result = await runSimulation({
      ...program,
      limits: {
        ...program.limits,
        no_progress_limit: 9,
        same_error_limit: 2,
      },
    }, blockIds, {
      ...run.options,
      verificationSequence: [
        {
          passed: false,
          progressFingerprint: "changed-1",
          errorFingerprint: "same-error",
        },
        {
          passed: false,
          progressFingerprint: "changed-2",
          errorFingerprint: "same-error",
        },
      ],
    });

    expect(result).toEqual({
      status: "paused",
      iterations: 2,
      reason: "same_error",
    });
  });

  it("stops before execution when the token budget is insufficient", async () => {
    const run = options([true]);
    const result = await runSimulation({
      ...program,
      limits: { ...program.limits, token_budget: 1 },
    }, blockIds, run.options);

    expect(result).toEqual({
      status: "paused",
      iterations: 0,
      reason: "token_budget",
    });
  });

  it("does not report success after the timeout deadline", async () => {
    const clock = vi.spyOn(Date, "now")
      .mockReturnValueOnce(0)
      .mockReturnValue(60_000);
    const run = options([true]);
    const result = await runSimulation({
      ...program,
      limits: { ...program.limits, timeout_minutes: 1 },
    }, blockIds, run.options);
    clock.mockRestore();

    expect(result).toEqual({
      status: "paused",
      iterations: 0,
      reason: "timeout",
    });
  });
});
