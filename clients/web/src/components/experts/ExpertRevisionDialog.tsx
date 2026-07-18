"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { SourceFormat } from "@proto/orchestration_resource/v1/orchestration_resource_types_pb";
import { AlertCircle, Loader2, RefreshCw } from "lucide-react";
import { useTranslations } from "next-intl";
import {
  assertResourceDraftIdentity,
  type ResourceIdentity,
} from "@/components/resource-editor/resource-draft-identity";
import { ResourceEditorShell } from "@/components/resource-editor/ResourceEditorShell";
import {
  RESOURCE_API_VERSION,
  type ExpertDraft,
} from "@/components/resource-editor/resource-editor-types";
import { Button } from "@/components/ui/button";
import {
  ResponsiveDialog,
  ResponsiveDialogContent,
  ResponsiveDialogHeader,
  ResponsiveDialogTitle,
} from "@/components/ui/responsive-dialog";
import {
  exportResource,
  getResourceCapabilities,
} from "@/lib/api/facade/orchestrationResource";

interface ExpertRevisionDialogProps {
  open: boolean;
  orgSlug: string;
  expertSlug: string;
  onOpenChange: (open: boolean) => void;
  onApplied: () => void;
}

export function ExpertRevisionDialog({
  open,
  orgSlug,
  expertSlug,
  onOpenChange,
  onApplied,
}: ExpertRevisionDialogProps) {
  const t = useTranslations();
  const request = useRef(0);
  const [draft, setDraft] = useState<ExpertDraft | null>(null);
  const [error, setError] = useState(false);
  const [permissionDenied, setPermissionDenied] = useState(false);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    const requestId = ++request.current;
    setDraft(null);
    setError(false);
    setPermissionDenied(false);
    setLoading(true);
    try {
      const target: ResourceIdentity = {
        apiVersion: RESOURCE_API_VERSION,
        kind: "Expert",
        namespace: orgSlug,
        name: expertSlug,
      };
      const access = await getResourceCapabilities(orgSlug, target);
      if (!access.capabilities) {
        throw new Error("Resource capabilities response is incomplete.");
      }
      if (
        !access.capabilities.canViewSource ||
        !access.capabilities.canPlan
      ) {
        if (request.current === requestId) setPermissionDenied(true);
        return;
      }
      const content = await exportResource(
        orgSlug,
        target,
        SourceFormat.YAML,
      );
      const { parseResourceYaml } = await import(
        "@/components/resource-editor/resource-yaml-codec"
      );
      const parsed = parseResourceYaml(
        new TextDecoder().decode(content),
        "Expert",
      );
      if (request.current !== requestId) return;
      if (parsed.kind !== "Expert") {
        throw new Error("Exported resource is not an Expert.");
      }
      assertResourceDraftIdentity(parsed, target);
      setDraft(parsed);
    } catch {
      if (request.current === requestId) setError(true);
    } finally {
      if (request.current === requestId) setLoading(false);
    }
  }, [expertSlug, orgSlug]);

  useEffect(() => {
    if (!open) {
      request.current += 1;
      setDraft(null);
      setError(false);
      setPermissionDenied(false);
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
            {t("experts.newRevision")}
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
                {t("experts.revisionLoadFailed")}
              </p>
              <Button variant="outline" size="sm" onClick={() => void load()}>
                <RefreshCw className="mr-2 h-4 w-4" />
                {t("common.retry")}
              </Button>
            </div>
          )}
          {permissionDenied && (
            <div
              role="alert"
              className="flex min-h-40 items-center justify-center text-center"
            >
              <p className="max-w-md text-sm text-muted-foreground">
                {t("experts.revisionPermissionDenied")}
              </p>
            </div>
          )}
          {draft && (
            <ResourceEditorShell
              key={`${expertSlug}-${request.current}`}
              orgSlug={orgSlug}
              kind="Expert"
              initialDraft={draft}
              lockedIdentity={{
                apiVersion: RESOURCE_API_VERSION,
                kind: "Expert",
                namespace: orgSlug,
                name: expertSlug,
              }}
              onApplied={onApplied}
            />
          )}
        </div>
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
