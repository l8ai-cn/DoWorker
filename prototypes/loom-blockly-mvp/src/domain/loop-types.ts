export interface SourceBlock<T> {
  blockId: string;
  value: T;
}

export interface WorkerSelection {
  snapshotId: number;
  label: string;
}

export interface LoopLimits {
  maxIterations: number;
  tokenBudget: number;
  timeoutMinutes: number;
  noProgressLimit: number;
  sameErrorLimit: number;
}

export interface UnknownBlockType {
  blockId: string;
  type: string;
}

export interface LoopDraft {
  name: string;
  rootBlockId: string;
  worker?: SourceBlock<WorkerSelection>;
  instructions: SourceBlock<string>[];
  acceptanceCriteria: SourceBlock<string>[];
  verification?: SourceBlock<string>;
  limits?: SourceBlock<LoopLimits>;
  escalationPolicy?: SourceBlock<"pause" | "fail">;
  looseBlockIds: string[];
  unknownBlockTypes: UnknownBlockType[];
  adapterDiagnostics: Diagnostic[];
}

export interface GoalLoopProgram {
  kind: "goal-loop-program";
  schema_version: 1;
  name: string;
  worker: {
    snapshot_id: number;
    label: string;
  };
  objective: string;
  acceptance_criteria: string[];
  verification: {
    kind: "command";
    command: string;
  };
  limits: {
    max_iterations: number;
    token_budget: number;
    timeout_minutes: number;
    no_progress_limit: number;
    same_error_limit: number;
  };
  escalation_policy: "pause" | "fail";
}

export interface Diagnostic {
  code: string;
  message: string;
  blockId?: string;
  slot?: string;
}

export interface CompileResult {
  diagnostics: Diagnostic[];
  program?: GoalLoopProgram;
  executionBlockIds: string[];
}
