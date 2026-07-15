import type {
  LoopDiagnostic,
  LoopDraftSnapshot,
  LoopProgram,
} from "@proto/goalloop/v1/goalloop_pb";

export type LoopEditor = "blocks" | "code";

export interface LoopWorkbenchSnapshot {
  source: string;
  canonicalSource: string;
  program?: LoopProgram;
  diagnostics: LoopDiagnostic[];
  parseStatus: string;
  activeEditor: LoopEditor;
  revision: number;
  semanticRevision: number;
  run?: LoopDraftSnapshot["run"];
}

export function createDefaultLoopSource(workerSnapshotId: number): string {
  return `@id(n-checkout-fix)
loop checkout-fix {
  @id(n-coder)
  worker coder = snapshot(${workerSnapshotId})
  limits(iterations: 5, tokens: 80000, timeout: 60m, no_progress: 3, same_error: 2)
  @id(n-fix-cycle)
  repeat fix-cycle(max: 5, until: tests.passed) {
    @id(n-fix-tax)
    agent fix-tax(using: coder) { prompt """修复结算页税额计算，并补充边界测试。""" }
    @id(n-tests)
    verify tests { command "pnpm test --filter billing" accept "完整测试集通过" }
  }
  on_failure pause
}`;
}
