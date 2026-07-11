import type { Pod } from "@/stores/pod";

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

export function workerSnapshotOptions(workers: Pod[]) {
  const snapshots = new Map<number, Pod>();
  for (const worker of workers) {
    if (worker.worker_spec_snapshot_id !== undefined && !snapshots.has(worker.worker_spec_snapshot_id)) {
      snapshots.set(worker.worker_spec_snapshot_id, worker);
    }
  }
  return [...snapshots.entries()].map(([snapshotId, worker]) => ({ snapshotId, worker }));
}

export function workerLabel(worker: Pod): string {
  const name = worker.alias ?? worker.title ?? worker.pod_key;
  return worker.agent?.name ? `${name} · ${worker.agent.name}` : name;
}
