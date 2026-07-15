import { AlertCircle, CheckCircle2, CircleDot, PlayCircle } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import type { LoopWorkbenchSnapshot } from "@/lib/viewModels/loop-program";
import {
  loopDiagnosticLabel,
  loopParseStatusLabel,
  loopRunStatusLabel,
} from "./loop-display-text";

interface LoopStatusPanelProps {
  snapshot: LoopWorkbenchSnapshot;
  error?: string;
  selectedNodeId?: string;
}

export function LoopStatusPanel({ snapshot, error, selectedNodeId }: LoopStatusPanelProps) {
  const valid = snapshot.parseStatus === "valid";
  return (
    <div className="grid min-h-0 grid-cols-1 border-t border-border lg:grid-cols-2">
      <section className="min-h-0 border-b border-border p-4 lg:border-b-0 lg:border-r">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-xs font-semibold uppercase text-muted-foreground">诊断</h2>
          <Badge variant={valid ? "success" : snapshot.parseStatus === "parsing" ? "info" : "warning"}>
            {loopParseStatusLabel(snapshot.parseStatus)}
          </Badge>
        </div>
        {error && (
          <p className="mb-2 flex items-start gap-2 text-xs text-destructive">
            <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />{error}
          </p>
        )}
        {snapshot.diagnostics.length === 0 && !error ? (
          <p className="flex items-center gap-2 text-xs text-muted-foreground">
            <CheckCircle2 className="h-3.5 w-3.5 text-success" />
            当前源码通过后端校验
          </p>
        ) : (
          <ul className="max-h-28 space-y-2 overflow-auto">
            {snapshot.diagnostics.map((diagnostic, index) => (
              <li className="text-xs" key={`${diagnostic.code}-${index}`}>
                <span className="text-destructive">{loopDiagnosticLabel(diagnostic.code)}</span>
                <span className="ml-2 text-muted-foreground">
                  第 {diagnostic.line} 行，第 {diagnostic.column} 列
                </span>
              </li>
            ))}
          </ul>
        )}
        {selectedNodeId && (
          <p className="mt-3 flex items-center gap-2 text-xs text-muted-foreground">
            <CircleDot className="h-3.5 w-3.5" />
            节点：<code className="font-mono text-foreground">{selectedNodeId}</code>
          </p>
        )}
      </section>
      <section className="p-4">
        <h2 className="mb-3 text-xs font-semibold uppercase text-muted-foreground">运行</h2>
        {snapshot.run ? (
          <div className="space-y-2 text-xs">
            <p className="flex items-center gap-2">
              <PlayCircle className="h-3.5 w-3.5 text-primary" />
              <code className="font-mono">{snapshot.run.slug}</code>
            </p>
            <p className="text-muted-foreground">
              状态：<span className="font-medium text-foreground">
                {loopRunStatusLabel(snapshot.run.status)}
              </span>
              {snapshot.run.podKey ? ` · 运行实例 ${snapshot.run.podKey}` : ""}
            </p>
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">尚未启动循环。</p>
        )}
      </section>
    </div>
  );
}
