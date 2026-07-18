import { create } from "@bufbuild/protobuf";
import { fireEvent, render, screen } from "@/test/test-utils";
import { describe, expect, it, vi } from "vitest";
import {
  LoopDiagnosticSchema,
  LoopDraftSnapshotSchema,
} from "@proto/goalloop/v1/goalloop_pb";
import { LoopStatusPanel } from "../loop-status-panel";
import type { LoopStatusMessages } from "../loop-workbench-messages";

const messages = {
  diagnosticsTitle: "诊断",
  runTitle: "运行",
  valid: "有效",
  noRun: "未运行",
  repairDiagnostic: "修复此诊断",
  repairingDiagnostic: "正在修复诊断",
  nodeLabel: "节点：",
  runStatusLabel: "状态：",
  parseStatusLabel: (status: string) => status,
  loopRunStatusLabel: (status: string) => status,
  diagnosticLabel: (code: string) => code,
  diagnosticLocation: (line: number, column: number) => `${line}:${column}`,
  runInstance: (podKey: string) => podKey,
} as LoopStatusMessages;

const repairable = create(LoopDiagnosticSchema, {
  code: "loop.value.out-of-range",
  nodeId: "n-limits",
  fieldPath: "limits.iterations",
  line: 2,
  column: 3,
});
const unsupported = create(LoopDiagnosticSchema, {
  code: "loop.text.empty",
  nodeId: "n-agent",
  fieldPath: "repeat.agent.prompt",
  line: 4,
  column: 5,
});

describe("LoopStatusPanel", () => {
  it("offers repair only for a supported diagnostic with an exact target", () => {
    const onRepairDiagnostic = vi.fn();
    render(
      <LoopStatusPanel
        messages={messages}
        onRepairDiagnostic={onRepairDiagnostic}
        snapshot={{
          source: "loop invalid {}",
          canonicalSource: "",
          diagnostics: [repairable, unsupported],
          parseStatus: "syntax-error",
          activeEditor: "code",
          revision: 7,
          semanticRevision: 3,
          run: create(LoopDraftSnapshotSchema),
        }}
      />,
    );

    const repair = screen.getByRole("button", { name: "修复此诊断" });
    expect(repair).toBeEnabled();
    expect(screen.getAllByRole("button", { name: "修复此诊断" })).toHaveLength(1);

    fireEvent.click(repair);
    expect(onRepairDiagnostic).toHaveBeenCalledWith(repairable);
  });

  it("disables the active repair target while the request is running", () => {
    render(
      <LoopStatusPanel
        messages={messages}
        onRepairDiagnostic={vi.fn()}
        repairingTarget={{
          nodeId: "n-limits",
          fieldPath: "limits.iterations",
        }}
        snapshot={{
          source: "loop invalid {}",
          canonicalSource: "",
          diagnostics: [repairable],
          parseStatus: "syntax-error",
          activeEditor: "code",
          revision: 7,
          semanticRevision: 3,
        }}
      />,
    );

    expect(screen.getByRole("button", { name: "正在修复诊断" })).toBeDisabled();
  });
});
