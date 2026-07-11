"use client";

import { AlertTriangle, CheckCircle2, Loader2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { AlertMessage } from "@/components/ui/alert-message";
import type { WorkerPreflightResult } from "@/lib/api/facade/podConnect";
import type { AsyncState } from "../hooks/workerCreateDraft";

interface WorkerPreflightStepProps {
  preflight: AsyncState<WorkerPreflightResult>;
  creating: boolean;
  onRetry: () => void;
  onCreate: () => void;
  t: (key: string) => string;
}

export function WorkerPreflightStep({
  preflight,
  creating,
  onRetry,
  onCreate,
  t,
}: WorkerPreflightStepProps) {
  if (preflight.status === "idle") {
    return (
      <Button type="button" onClick={onRetry}>
        {t("workerCreate.preflight.run")}
      </Button>
    );
  }
  if (preflight.status === "loading") {
    return (
      <div className="flex items-center gap-2 py-8 text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        {t("workerCreate.preflight.loading")}
      </div>
    );
  }
  if (preflight.status === "error") {
    return (
      <div className="space-y-4">
        <AlertMessage type="error" message={preflight.error} />
        <Button type="button" variant="outline" onClick={onRetry}>
          <RefreshCw className="mr-2 h-4 w-4" />
          {t("workerCreate.preflight.retry")}
        </Button>
      </div>
    );
  }

  const blocking = preflight.data.issues.filter((issue) => issue.severity === "blocking");
  const warnings = preflight.data.issues.filter((issue) => issue.severity !== "blocking");
  const hasResolvedSpec = Boolean(preflight.data.resolved_spec_json?.trim());

  return (
    <div className="space-y-5">
      {blocking.length > 0 ? (
        <IssueGroup
          testId="preflight-blocking"
          icon={AlertTriangle}
          title={t("workerCreate.preflight.blocking")}
          issues={blocking}
          tone="destructive"
        />
      ) : !hasResolvedSpec ? (
        <AlertMessage
          type="error"
          message={t("workerCreate.preflight.missingResolvedSpec")}
        />
      ) : (
        <div className="flex items-center gap-2 text-sm text-success">
          <CheckCircle2 className="h-4 w-4" />
          {t("workerCreate.preflight.ready")}
        </div>
      )}
      {warnings.length > 0 && (
        <IssueGroup
          testId="preflight-warnings"
          icon={AlertTriangle}
          title={t("workerCreate.preflight.warnings")}
          issues={warnings}
          tone="warning"
        />
      )}
      <div className="flex justify-end">
        <Button
          type="button"
          onClick={onCreate}
          disabled={blocking.length > 0 || !hasResolvedSpec || creating}
        >
          {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {t("workerCreate.actions.create")}
        </Button>
      </div>
    </div>
  );
}

function IssueGroup(props: {
  testId: string;
  icon: typeof AlertTriangle;
  title: string;
  issues: WorkerPreflightResult["issues"];
  tone: "destructive" | "warning";
}) {
  const Icon = props.icon;
  const color = props.tone === "destructive" ? "text-destructive" : "text-warning";
  return (
    <section data-testid={props.testId} className="rounded-md border border-border p-3">
      <h3 className={`mb-2 flex items-center gap-2 text-sm font-medium ${color}`}>
        <Icon className="h-4 w-4" />
        {props.title}
      </h3>
      <ul className="space-y-1 text-sm text-foreground">
        {props.issues.map((issue) => (
          <li key={`${issue.code}:${issue.field}:${issue.message}`}>{issue.message}</li>
        ))}
      </ul>
    </section>
  );
}
