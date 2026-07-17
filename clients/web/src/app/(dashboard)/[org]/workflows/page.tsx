"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { useWorkflowStore, useWorkflows } from "@/stores/workflow";
import { useCurrentOrg } from "@/stores/auth";
import { CenteredSpinner } from "@/components/ui/spinner";
import { AlertCircle, RefreshCw, Repeat } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { ResourceEditorShell } from "@/components/resource-editor/ResourceEditorShell";
import { appliedResource } from "@/components/resource-editor/resource-apply-result";

export default function WorkflowsIndexPage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const orgSlug = params.org as string;
  const currentOrg = useCurrentOrg();
  const workflows = useWorkflows();
  const loading = useWorkflowStore((s) => s.loading);
  const error = useWorkflowStore((s) => s.error);
  const fetchWorkflows = useWorkflowStore((s) => s.fetchWorkflows);
  const clearError = useWorkflowStore((s) => s.clearError);
  const [createdSlug, setCreatedSlug] = useState<string | null>(null);

  useEffect(() => {
    if (currentOrg) fetchWorkflows();
  }, [currentOrg, fetchWorkflows]);

  useEffect(() => {
    if (loading || workflows.length === 0) return;
    const target = createdSlug ?? workflows.find((l) => l.status === "enabled")?.slug ?? workflows[0]?.slug;
    if (target && orgSlug) router.replace(`/${orgSlug}/workflows/${target}`);
  }, [workflows, loading, orgSlug, router, createdSlug]);

  if (error && workflows.length === 0) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-3 py-20 text-center">
        <div className="flex h-12 w-12 items-center justify-center rounded-md bg-destructive/10">
          <AlertCircle className="h-6 w-6 text-destructive" />
        </div>
        <p className="text-sm text-muted-foreground">{error}</p>
        <Button variant="outline" size="sm" className="gap-1.5"
          onClick={() => { clearError(); fetchWorkflows(); }}>
          <RefreshCw className="h-3.5 w-3.5" />
          {t("workflows.retry")}
        </Button>
      </div>
    );
  }

  if (loading && workflows.length === 0) return <CenteredSpinner className="h-full" />;

  if (workflows.length === 0) {
    return (
      <div className="h-full overflow-y-auto">
        <div className="mx-auto w-full max-w-3xl px-6 py-10">
          <header className="mb-8 text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-surface-muted text-muted-foreground">
              <Repeat className="h-7 w-7" />
            </div>
            <h1 className="text-xl font-semibold tracking-tight">{t("workflows.emptyTitle")}</h1>
            <p className="mx-auto mt-2 max-w-lg text-sm text-muted-foreground">
              {t("workflows.emptyDescription")}
            </p>
          </header>
          <ResourceEditorShell
            orgSlug={orgSlug}
            kind="Workflow"
            onApplied={(result) => {
              const name = appliedResource(result)?.identity?.target?.name;
              void fetchWorkflows();
              if (name) setCreatedSlug(name);
            }}
          />
        </div>
      </div>
    );
  }

  return <CenteredSpinner className="h-full" />;
}
