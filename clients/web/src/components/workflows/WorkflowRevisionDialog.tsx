"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { SourceFormat } from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import { AlertCircle, Loader2, RefreshCw } from "lucide-react";
import { useTranslations } from "next-intl";
import { ResourceEditorShell } from "@/components/resource-editor/ResourceEditorShell";
import {
  RESOURCE_API_VERSION,
  type WorkflowDraft,
} from "@/components/resource-editor/resource-editor-types";
import { resourceDraftIdentity } from "@/components/resource-editor/resource-draft-identity";
import { Button } from "@/components/ui/button";
import {
  ResponsiveDialog,
  ResponsiveDialogContent,
  ResponsiveDialogHeader,
  ResponsiveDialogTitle,
} from "@/components/ui/responsive-dialog";
import { exportResource } from "@/lib/api/facade/orchestrationResource";

interface WorkflowRevisionDialogProps {
  open: boolean;
  orgSlug: string;
  workflowSlug: string;
  onOpenChange: (open: boolean) => void;
  onApplied: () => void;
}

export function WorkflowRevisionDialog({
  open,
  orgSlug,
  workflowSlug,
  onOpenChange,
  onApplied,
}: WorkflowRevisionDialogProps) {
  const t = useTranslations();
  const request = useRef(0);
  const [draft, setDraft] = useState<WorkflowDraft | null>(null);
  const [error, setError] = useState(false);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    const requestId = ++request.current;
    setDraft(null);
    setError(false);
    setLoading(true);
    try {
      const content = await exportResource(
        orgSlug,
        {
          apiVersion: RESOURCE_API_VERSION,
          kind: "Workflow",
          namespace: orgSlug,
          name: workflowSlug,
        },
        SourceFormat.YAML,
      );
      const { parseResourceYaml } = await import(
        "@/components/resource-editor/resource-yaml-codec"
      );
      const parsed = parseResourceYaml(
        new TextDecoder().decode(content),
        "Workflow",
      );
      if (request.current !== requestId || parsed.kind !== "Workflow") return;
      setDraft(parsed);
    } catch {
      if (request.current === requestId) setError(true);
    } finally {
      if (request.current === requestId) setLoading(false);
    }
  }, [orgSlug, workflowSlug]);

  useEffect(() => {
    if (!open) {
      request.current += 1;
      setDraft(null);
      setError(false);
      setLoading(false);
      return;
    }
    void load();
    return () => {
      request.current += 1;
    };
  }, [load, open]);

  return (
    <ResponsiveDialog open={open} onOpenChange={onOpenChange}>
      <ResponsiveDialogContent className="max-h-[90vh] max-w-5xl overflow-y-auto">
        <ResponsiveDialogHeader onClose={() => onOpenChange(false)}>
          <ResponsiveDialogTitle>
            {t("workflows.newRevision")}
          </ResponsiveDialogTitle>
        </ResponsiveDialogHeader>
        <div className="px-4 pb-4">
          {loading && (
            <div className="flex min-h-40 items-center justify-center">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          )}
          {error && (
            <div
              role="alert"
              className="flex min-h-40 flex-col items-center justify-center gap-3 text-center"
            >
              <AlertCircle className="h-6 w-6 text-destructive" />
              <p className="text-sm text-muted-foreground">
                {t("workflows.revisionLoadFailed")}
              </p>
              <Button variant="outline" size="sm" onClick={() => void load()}>
                <RefreshCw className="mr-2 h-4 w-4" />
                {t("common.retry")}
              </Button>
            </div>
          )}
          {draft && (
            <ResourceEditorShell
              key={`${workflowSlug}-${request.current}`}
              orgSlug={orgSlug}
              kind="Workflow"
              initialDraft={draft}
              lockedIdentity={resourceDraftIdentity(draft)}
              onApplied={onApplied}
            />
          )}
        </div>
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
