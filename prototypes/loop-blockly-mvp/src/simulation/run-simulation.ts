import type { GoalLoopProgram } from "../domain/loop-types";
import {
  estimatedTokens,
  failureStopReason,
  timeoutReached,
  type SimulationStopReason,
} from "./simulation-limits";
import { evidence, verificationOutcome, wait } from "./simulation-runtime";
import type {
  SimulationOptions,
  SimulationResult,
} from "./simulation-types";

export type {
  SimulationEvidence,
  SimulationOptions,
  SimulationResult,
  SimulationVerificationInput,
  SimulationVerificationOutcome,
} from "./simulation-types";

function terminalResult(
  program: GoalLoopProgram,
  rootBlockId: string,
  iterations: number,
  reason: SimulationStopReason,
  options: SimulationOptions,
): SimulationResult {
  const status = program.escalation_policy === "pause" ? "paused" : "failed";
  options.onHighlight(rootBlockId);
  options.onEvidence(evidence(
    rootBlockId,
    status,
    status === "paused"
      ? `触发 ${reason}，模拟已暂停并等待人工处理。`
      : `触发 ${reason}，模拟已按策略失败。`,
  ));
  return { status, iterations, reason };
}

export async function runSimulation(
  program: GoalLoopProgram,
  blockIds: string[],
  options: SimulationOptions,
): Promise<SimulationResult> {
  const delayMs = options.delayMs ?? 420;
  const verificationSequence = options.verificationSequence ?? [true];
  const rootBlockId = blockIds[0];
  const perIterationTokens = estimatedTokens(program);
  const startedAt = Date.now();
  let tokens = 0;
  let lastProgressFingerprint: string | undefined;
  let lastErrorFingerprint: string | undefined;
  let noProgressCount = 0;
  let sameErrorCount = 0;

  if (perIterationTokens > program.limits.token_budget) {
    return terminalResult(
      program,
      rootBlockId,
      0,
      "token_budget",
      options,
    );
  }

  for (let iteration = 1; iteration <= program.limits.max_iterations; iteration += 1) {
    options.signal.throwIfAborted();
    if (tokens + perIterationTokens > program.limits.token_budget) {
      return terminalResult(
        program,
        rootBlockId,
        iteration - 1,
        "token_budget",
        options,
      );
    }
    const verificationInput = verificationSequence[iteration - 1];
    if (verificationInput === undefined) {
      return terminalResult(
        program,
        rootBlockId,
        iteration - 1,
        "scenario_exhausted",
        options,
      );
    }
    const outcome = verificationOutcome(verificationInput, iteration);
    for (const [index, blockId] of blockIds.entries()) {
      options.signal.throwIfAborted();
      if (timeoutReached(program, startedAt)) {
        return terminalResult(
          program,
          rootBlockId,
          iteration - 1,
          "timeout",
          options,
        );
      }
      options.onHighlight(blockId);
      const verifier = index === blockIds.length - 1;
      options.onEvidence(evidence(
        blockId,
        index === 0 ? "start" : verifier ? "verify" : "step",
        index === 0
          ? `开始模拟第 ${iteration} 轮。`
          : verifier
            ? outcome.passed
              ? "模拟验证通过（预设场景）。"
              : "模拟验证未通过（预设场景）。"
            : `已模拟执行步骤 ${index}。`,
      ));
      await wait(delayMs, options.signal);
      if (timeoutReached(program, startedAt)) {
        return terminalResult(
          program,
          rootBlockId,
          iteration - 1,
          "timeout",
          options,
        );
      }
    }
    tokens += perIterationTokens;
    if (outcome.passed) {
      options.onHighlight(rootBlockId);
      options.onEvidence(evidence(
        rootBlockId,
        "complete",
        `模拟完成：第 ${iteration} 轮验证通过。`,
      ));
      await wait(delayMs, options.signal);
      options.onHighlight(null);
      return { status: "completed", iterations: iteration };
    }
    noProgressCount = outcome.progressFingerprint === lastProgressFingerprint
      ? noProgressCount + 1
      : 1;
    sameErrorCount = outcome.errorFingerprint === lastErrorFingerprint
      ? sameErrorCount + 1
      : 1;
    lastProgressFingerprint = outcome.progressFingerprint;
    lastErrorFingerprint = outcome.errorFingerprint;
    const reason = failureStopReason(
      program,
      iteration,
      noProgressCount,
      sameErrorCount,
    );
    if (reason) {
      return terminalResult(program, rootBlockId, iteration, reason, options);
    }
  }
  return terminalResult(
    program,
    rootBlockId,
    program.limits.max_iterations,
    "max_iterations",
    options,
  );
}
