"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { usePodCreationData } from "@/components/pod/hooks";
import { getPod } from "@/lib/api/facade/podConnect";
import {
  fetchAllSessionItems,
  importCodexSession,
} from "@/lib/api/sessionImportApi";
import {
  codexItemsToAcpSnapshot,
  ACP_SNAPSHOT_MSG_TYPE,
} from "@/lib/codexItemsToAcpSnapshot";
import { dispatchAcpRelayEvent } from "@/stores/acpEventDispatcher";
import { usePodStore } from "@/stores/pod";
import { readCurrentOrg } from "@/stores/auth";
import { refreshImportedSessionsList } from "@/components/ide/sidebar/ImportedSessionsSection";
import { ImportCodexFormFields } from "./ImportCodexFormFields";
import { ImportCodexModelResourceSelect } from "./ImportCodexModelResourceSelect";

export function ImportCodexDialog({
  open,
  onOpenChange,
  onImported,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onImported: (podKey: string) => void;
}) {
  const t = useTranslations();
  const { availableAgents, loading: loadingAgents } = usePodCreationData(open);
  const orgSlug = readCurrentOrg()?.slug;

  const [sourcePath, setSourcePath] = useState("");
  const [title, setTitle] = useState("");
  const [selectedAgentSlug, setSelectedAgentSlug] = useState<string | null>(null);
  const [selectedModelResourceId, setSelectedModelResourceId] = useState<number | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open || selectedAgentSlug || availableAgents.length === 0) return;
    setSelectedAgentSlug(availableAgents[0]!.slug);
  }, [open, availableAgents, selectedAgentSlug]);

  function resetForm(): void {
    setSourcePath("");
    setTitle("");
    setSelectedAgentSlug(null);
    setSelectedModelResourceId(null);
    setError(null);
    setSubmitting(false);
  }

  function handleOpenChange(next: boolean): void {
    if (!next) resetForm();
    onOpenChange(next);
  }

  async function handleImport(): Promise<void> {
    const path = sourcePath.trim();
    if (!path) {
      setError(t("workers.create.import.errors.sourceRequired"));
      return;
    }
    if (!selectedAgentSlug) {
      setError(t("workers.create.import.errors.agentRequired"));
      return;
    }

    if (!orgSlug) {
      setError(t("workers.create.import.errors.notAuthenticated"));
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      const result = await importCodexSession(path, selectedAgentSlug, {
        title: title.trim() || undefined,
        modelResourceId: selectedModelResourceId ?? undefined,
      });

      const items = await fetchAllSessionItems(result.sessionId);
      const snapshot = codexItemsToAcpSnapshot(result.sessionId, items);
      dispatchAcpRelayEvent(result.podKey, ACP_SNAPSHOT_MSG_TYPE, snapshot);

      try {
        const pod = await getPod(orgSlug, result.podKey);
        usePodStore.getState().upsertPod(pod);
      } catch {
        // Pod metadata is optional for preview; workspace opens by pod_key.
      }

      toast.success(t("workers.create.import.success"), {
        description: t("workers.create.import.successDetail", {
          count: result.itemCount,
        }),
      });

      void usePodStore.getState().fetchSidebarPods("mine", { silent: true });
      refreshImportedSessionsList();

      handleOpenChange(false);
      onImported(result.podKey);
    } catch (e) {
      setError(
        e instanceof Error ? e.message : t("workers.create.import.errors.failed"),
      );
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent
        data-testid="import-codex-dialog"
        className="flex max-h-[85vh] flex-col gap-4 sm:max-w-lg"
      >
        <DialogHeader>
          <DialogTitle>{t("workers.create.import.title")}</DialogTitle>
        </DialogHeader>

        <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto">
          <ImportCodexFormFields
            agents={availableAgents}
            sourcePath={sourcePath}
            title={title}
            selectedAgentSlug={selectedAgentSlug}
            onSourcePathChange={setSourcePath}
            onTitleChange={setTitle}
            onAgentChange={setSelectedAgentSlug}
            t={t}
          />

          <ImportCodexModelResourceSelect
            open={open}
            orgSlug={orgSlug ?? ""}
            workerTypeSlug={selectedAgentSlug}
            selectedResourceId={selectedModelResourceId}
            onSelect={setSelectedModelResourceId}
            t={t}
          />

          {loadingAgents && (
            <p className="text-xs text-muted-foreground">{t("common.loading")}</p>
          )}

          {error !== null && (
            <p data-testid="import-codex-error" className="text-xs text-destructive">
              {error}
            </p>
          )}
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => handleOpenChange(false)} disabled={submitting}>
            {t("common.cancel")}
          </Button>
          <Button
            data-testid="import-codex-submit"
            onClick={handleImport}
            disabled={!sourcePath.trim() || !selectedAgentSlug || submitting}
          >
            {submitting ? t("workers.create.import.submitting") : t("workers.create.import.submit")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
