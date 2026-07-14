import type { SimulationStopReason } from "./simulation-limits";

export interface SimulationEvidence {
  id: string;
  blockId: string;
  kind: "start" | "step" | "verify" | "complete" | "paused" | "failed";
  message: string;
  timestamp: string;
}

export interface SimulationResult {
  status: "completed" | "paused" | "failed";
  iterations: number;
  reason?: SimulationStopReason;
}

export interface SimulationVerificationOutcome {
  passed: boolean;
  progressFingerprint: string;
  errorFingerprint?: string;
}

export type SimulationVerificationInput =
  | boolean
  | SimulationVerificationOutcome;

export interface SimulationOptions {
  signal: AbortSignal;
  delayMs?: number;
  verificationSequence?: SimulationVerificationInput[];
  onEvidence: (event: SimulationEvidence) => void;
  onHighlight: (blockId: string | null) => void;
}
