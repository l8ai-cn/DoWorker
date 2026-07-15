export type GoalLoopStatus =
  | "draft"
  | "active"
  | "paused"
  | "verifying"
  | "completed"
  | "failed"
  | "cancelled";

export interface GoalLoopData {
  id: number;
  slug: string;
  name: string;
  description?: string;
  worker_spec_snapshot_id: number;
  objective: string;
  acceptance_criteria: string[];
  verification_command: string;
  status: GoalLoopStatus;
  pod_key?: string;
  max_iterations: number;
  token_budget?: number;
  timeout_minutes: number;
  no_progress_limit: number;
  same_error_limit: number;
  escalation_policy: "pause" | "fail";
  verification_exit_code?: number;
  verification_output?: string;
  verification_output_truncated: boolean;
  verification_error?: string;
  started_at?: string;
  verified_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface GoalLoopWorkerSnapshot {
  id: number;
  alias: string;
  worker_type: string;
  created_at: string;
}

export interface CreateGoalLoopInput {
  name: string;
  description?: string;
  worker_spec_snapshot_id: number;
  objective: string;
  acceptance_criteria: string[];
  verification_command: string;
  max_iterations?: number;
  token_budget?: number;
  timeout_minutes?: number;
  no_progress_limit?: number;
  same_error_limit?: number;
  escalation_policy: "pause" | "fail";
}
