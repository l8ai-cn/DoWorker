import type {
  SimulationEvidence,
  SimulationVerificationInput,
  SimulationVerificationOutcome,
} from "./simulation-types";

export function wait(
  delayMs: number,
  signal: AbortSignal,
): Promise<void> {
  if (delayMs === 0) return Promise.resolve();
  return new Promise((resolve, reject) => {
    const onAbort = () => {
      clearTimeout(timer);
      reject(new DOMException("Simulation aborted", "AbortError"));
    };
    const timer = setTimeout(() => {
      signal.removeEventListener("abort", onAbort);
      resolve();
    }, delayMs);
    signal.addEventListener("abort", onAbort, { once: true });
  });
}

export function evidence(
  blockId: string,
  kind: SimulationEvidence["kind"],
  message: string,
): SimulationEvidence {
  return {
    id: crypto.randomUUID(),
    blockId,
    kind,
    message,
    timestamp: new Date().toISOString(),
  };
}

export function verificationOutcome(
  input: SimulationVerificationInput,
  iteration: number,
): SimulationVerificationOutcome {
  if (typeof input === "boolean") {
    return input
      ? { passed: true, progressFingerprint: `verified-${iteration}` }
      : {
        passed: false,
        progressFingerprint: "unchanged",
        errorFingerprint: "verification-failed",
      };
  }
  if (
    input.progressFingerprint.trim() === "" ||
    (!input.passed && (!input.errorFingerprint ||
      input.errorFingerprint.trim() === ""))
  ) {
    throw new Error("模拟验证场景缺少进展或错误指纹。");
  }
  return input;
}
