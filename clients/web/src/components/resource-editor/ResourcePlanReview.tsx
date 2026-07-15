"use client";

import { useTranslations } from "next-intl";
import {
  IssueSeverity,
  ResourceOperation,
} from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import { AlertMessage } from "@/components/ui/alert-message";
import { Badge } from "@/components/ui/badge";
import { Spinner } from "@/components/ui/spinner";
import type { ResourceDraftState } from "./resource-draft-reducer";
import { ResourceSemanticDiff } from "./ResourceSemanticDiff";

interface ResourcePlanReviewProps {
  planState: ResourceDraftState["plan"];
}

export function ResourcePlanReview({
  planState,
}: ResourcePlanReviewProps) {
  const t = useTranslations("resourceEditor");
  if (planState.status === "idle") {
    return (
      <div className="py-16 text-center text-sm text-muted-foreground">
        {t("plan.empty")}
      </div>
    );
  }
  if (planState.status === "loading") {
    return (
      <div className="flex min-h-64 items-center justify-center">
        <Spinner />
      </div>
    );
  }
  if (planState.status === "error") {
    return <AlertMessage type="error" message={planState.error} />;
  }
  if (planState.status === "expired") {
    return <AlertMessage type="warning" message={t("plan.expired")} />;
  }
  const { response } = planState;
  const plan = response.plan;
  return (
    <div className="space-y-6">
      {response.issues.length > 0 && (
        <section className="space-y-2">
          <h3 className="text-sm font-semibold">{t("plan.issues")}</h3>
          {response.issues.map((issue, index) => (
            <AlertMessage
              key={`${issue.path}-${issue.code}-${index}`}
              type={issue.severity === IssueSeverity.BLOCKING
                ? "error"
                : "warning"}
              message={`${issue.path || "/"}: ${issue.message}`}
            />
          ))}
        </section>
      )}
      {plan && (
        <>
          <section className="grid gap-3 border-b border-border pb-5 sm:grid-cols-3">
            <PlanValue
              label={t("plan.operation")}
              value={ResourceOperation[plan.operation]}
            />
            <PlanValue label="Plan ID" value={plan.planId} mono />
            <PlanValue label={t("plan.expires")} value={plan.expiresAt} />
          </section>
          <section className="space-y-3">
            <h3 className="text-sm font-semibold">{t("plan.changes")}</h3>
            <ResourceSemanticDiff
              changes={plan.semanticDiff}
              emptyLabel={t("plan.noChanges")}
            />
          </section>
          <section className="space-y-3">
            <h3 className="text-sm font-semibold">{t("plan.references")}</h3>
            {plan.resolvedReferences.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                {t("plan.noReferences")}
              </p>
            ) : (
              <div className="divide-y divide-border rounded-md border border-border">
                {plan.resolvedReferences.map((reference) => (
                  <div
                    key={`${reference.typeMeta?.kind}-${reference.name}-${reference.revision}`}
                    className="flex flex-wrap items-center gap-2 px-3 py-2.5 text-sm"
                  >
                    <Badge variant="outline">
                      {reference.typeMeta?.kind ?? "Resource"}
                    </Badge>
                    <span className="font-medium">{reference.name}</span>
                    <span className="text-muted-foreground">
                      r{reference.revision.toString()}
                    </span>
                    <code className="ml-auto text-[11px] text-muted-foreground">
                      {reference.digest.slice(0, 12)}
                    </code>
                  </div>
                ))}
              </div>
            )}
          </section>
        </>
      )}
    </div>
  );
}

function PlanValue({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <div className="min-w-0">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className={mono
        ? "mt-1 truncate font-mono text-xs"
        : "mt-1 truncate text-sm font-medium"}
      >
        {value || "—"}
      </div>
    </div>
  );
}
