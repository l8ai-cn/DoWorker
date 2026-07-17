import {
  Tooltip,
  TooltipContent,
  TooltipPortal,
  TooltipProvider,
  TooltipTrigger,
} from "@radix-ui/react-tooltip";
import { AlertCircle, CheckCircle2, CircleDot, PlayCircle, Wrench } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { LoopWorkbenchSnapshot } from "@/lib/viewModels/loop-program";
import type { LoopDiagnostic } from "@proto/goalloop/v1/goalloop_pb";
import type { LoopStatusMessages } from "./loop-workbench-messages";

interface LoopStatusPanelProps {
  snapshot: LoopWorkbenchSnapshot;
  error?: string;
  selectedNodeId?: string;
  messages: LoopStatusMessages;
  repairingTarget?: Pick<LoopDiagnostic, "nodeId" | "fieldPath">;
  onRepairDiagnostic?: (diagnostic: LoopDiagnostic) => void;
}

export function LoopStatusPanel({
  snapshot,
  error,
  selectedNodeId,
  messages,
  repairingTarget,
  onRepairDiagnostic,
}: LoopStatusPanelProps) {
  const valid = snapshot.parseStatus === "valid";
  return (
    <div className="grid min-h-0 grid-cols-1 border-t border-border lg:grid-cols-2">
      <section className="min-h-0 border-b border-border p-4 lg:border-b-0 lg:border-r">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-xs font-semibold uppercase text-muted-foreground">
            {messages.diagnosticsTitle}
          </h2>
          <Badge variant={valid ? "success" : snapshot.parseStatus === "parsing" ? "info" : "warning"}>
            {messages.parseStatusLabel(snapshot.parseStatus)}
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
            {messages.valid}
          </p>
        ) : (
          <ul className="max-h-28 space-y-2 overflow-auto">
            {snapshot.diagnostics.map((diagnostic, index) => (
              <li
                className="flex min-h-7 items-start justify-between gap-2 text-xs"
                key={`${diagnostic.code}-${diagnostic.nodeId}-${diagnostic.fieldPath}-${index}`}
              >
                <span className="min-w-0 pt-1">
                  <span className="text-destructive">
                    {messages.diagnosticLabel(diagnostic.code)}
                  </span>
                  <span className="ml-2 text-muted-foreground">
                    {messages.diagnosticLocation(diagnostic.line, diagnostic.column)}
                  </span>
                </span>
                {isRepairableDiagnostic(diagnostic) && onRepairDiagnostic && (
                  <RepairDiagnosticButton
                    diagnostic={diagnostic}
                    messages={messages}
                    repairingTarget={repairingTarget}
                    onRepairDiagnostic={onRepairDiagnostic}
                  />
                )}
              </li>
            ))}
          </ul>
        )}
        {selectedNodeId && (
          <p className="mt-3 flex items-center gap-2 text-xs text-muted-foreground">
            <CircleDot className="h-3.5 w-3.5" />
            {messages.nodeLabel}<code className="font-mono text-foreground">{selectedNodeId}</code>
          </p>
        )}
      </section>
      <section className="p-4">
        <h2 className="mb-3 text-xs font-semibold uppercase text-muted-foreground">
          {messages.runTitle}
        </h2>
        {snapshot.run ? (
          <div className="space-y-2 text-xs">
            <p className="flex items-center gap-2">
              <PlayCircle className="h-3.5 w-3.5 text-primary" />
              <code className="font-mono">{snapshot.run.slug}</code>
            </p>
            <p className="text-muted-foreground">
              {messages.runStatusLabel}<span className="font-medium text-foreground">
                {messages.loopRunStatusLabel(snapshot.run.status)}
              </span>
              {snapshot.run.podKey ? messages.runInstance(snapshot.run.podKey) : ""}
            </p>
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">{messages.noRun}</p>
        )}
      </section>
    </div>
  );
}

const repairableDiagnosticCodes = new Set([
  "loop.value.out-of-range",
  "loop.repeat.max-exceeds-limit",
]);

function isRepairableDiagnostic(diagnostic: LoopDiagnostic): boolean {
  return Boolean(
    repairableDiagnosticCodes.has(diagnostic.code) &&
    diagnostic.nodeId &&
    diagnostic.fieldPath,
  );
}

function RepairDiagnosticButton({
  diagnostic,
  messages,
  repairingTarget,
  onRepairDiagnostic,
}: {
  diagnostic: LoopDiagnostic;
  messages: LoopStatusMessages;
  repairingTarget?: Pick<LoopDiagnostic, "nodeId" | "fieldPath">;
  onRepairDiagnostic: (diagnostic: LoopDiagnostic) => void;
}) {
  const repairing =
    repairingTarget?.nodeId === diagnostic.nodeId &&
    repairingTarget.fieldPath === diagnostic.fieldPath;
  const label = repairing
    ? messages.repairingDiagnostic
    : messages.repairDiagnostic;

  return (
    <TooltipProvider delayDuration={300}>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            aria-label={label}
            className="h-7 w-7 shrink-0"
            disabled={repairing}
            size="icon"
            variant="ghost"
            onClick={() => onRepairDiagnostic(diagnostic)}
          >
            <Wrench className="h-3.5 w-3.5" />
          </Button>
        </TooltipTrigger>
        <TooltipPortal>
          <TooltipContent
            className="z-[100002] rounded-md bg-popover px-2 py-1 text-xs text-popover-foreground shadow-[var(--shadow-soft)]"
            side="top"
          >
            {label}
          </TooltipContent>
        </TooltipPortal>
      </Tooltip>
    </TooltipProvider>
  );
}
