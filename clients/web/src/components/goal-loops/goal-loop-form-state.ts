import type { GoalLoopWorkerSnapshot } from "@/lib/viewModels/goal-loop";

export interface GoalLoopFormState {
  name: string;
  description: string;
  workerSnapshotId: string;
  objective: string;
  criteria: string;
  verificationCommand: string;
  maxIterations: string;
  tokenBudget: string;
  timeoutMinutes: string;
  noProgressLimit: string;
  sameErrorLimit: string;
  escalationPolicy: "pause" | "fail";
}

export const initialGoalLoopForm: GoalLoopFormState = {
  name: "",
  description: "",
  workerSnapshotId: "",
  objective: "",
  criteria: "",
  verificationCommand: "",
  maxIterations: "10",
  tokenBudget: "",
  timeoutMinutes: "60",
  noProgressLimit: "3",
  sameErrorLimit: "2",
  escalationPolicy: "pause",
};

export function optionalNumber(value: string): number | undefined {
  return value.trim() === "" ? undefined : Number(value);
}

export function workerLabel(worker: GoalLoopWorkerSnapshot): string {
  const workerType = worker.worker_type || "Worker";
  return worker.alias ? `${worker.alias} · ${workerType}` : `${workerType} · 快照 #${worker.id}`;
}
