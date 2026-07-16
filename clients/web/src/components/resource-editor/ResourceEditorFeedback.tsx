"use client";

import { useTranslations } from "next-intl";
import { IssueSeverity } from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import { AlertMessage } from "@/components/ui/alert-message";
import type { ResourceDraftState } from "./resource-draft-reducer";

interface ResourceEditorFeedbackProps {
  state: ResourceDraftState;
}

export function ResourceEditorFeedback({
  state,
}: ResourceEditorFeedbackProps) {
  const t = useTranslations("resourceEditor");
  if (state.validation.status === "error") {
    return <AlertMessage type="error" message={state.validation.error} />;
  }
  if (state.plan.status === "error") {
    return <AlertMessage type="error" message={state.plan.error} />;
  }
  if (state.plan.status === "expired") {
    return <AlertMessage type="warning" message={t("plan.expired")} />;
  }
  if (state.apply.status === "error") {
    return <AlertMessage type="error" message={state.apply.error} />;
  }
  if (
    state.validation.status === "ready" &&
    state.validation.response.issues.length > 0
  ) {
    return (
      <div className="space-y-2">
        {state.validation.response.issues.map((issue, index) => (
          <AlertMessage
            key={`${issue.path}-${issue.code}-${index}`}
            type={issue.severity === IssueSeverity.BLOCKING
              ? "error"
              : "warning"}
            message={`${issue.path || "/"}: ${issue.message}`}
          />
        ))}
      </div>
    );
  }
  if (
    state.validation.status === "ready" &&
    state.validation.response.issues.length === 0
  ) {
    return <AlertMessage type="success" message={t("validation.valid")} />;
  }
  if (state.apply.status !== "ready") return null;
  const result = state.apply.result;
  if ("workerSpecSnapshotId" in result) {
    return (
      <AlertMessage
        type="success"
        message={`${t("result.revision", {
          revision: result.resource?.revision.toString() ?? "0",
        })} · ${t("result.snapshot", {
          snapshot: result.workerSpecSnapshotId.toString(),
        })}`}
      />
    );
  }
  return (
    <AlertMessage
      type="success"
      message={t("result.revision", { revision: result.revision.toString() })}
    />
  );
}
