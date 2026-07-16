"use client";

import { useState } from "react";
import { FileJson, Upload, X } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import { createEnvBundle } from "@/lib/api/facade/envBundleConnect";
import type { EnvBundleSummary } from "@/lib/viewModels/envBundleSummary";
import { CONFIG_BUNDLE_JSON_KEY } from "@/components/settings/AgentConfigPage/configBundleKeys";

interface WorkerConfigFileSelectProps {
  agentSlug: string;
  bundles: EnvBundleSummary[];
  selectedIds: number[];
  onChange: (ids: number[]) => void;
  t: (key: string) => string;
}

const MAX_FILE_BYTES = 1024 * 1024;

export function WorkerConfigFileSelect({
  agentSlug,
  bundles,
  selectedIds,
  onChange,
  t,
}: WorkerConfigFileSelectProps) {
  const [uploaded, setUploaded] = useState<EnvBundleSummary[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const allBundles = [...bundles, ...uploaded.filter((item) =>
    !bundles.some((bundle) => bundle.id === item.id),
  )];
  const selectedId = selectedIds[0];

  const select = (id: number) => onChange(selectedId === id ? [] : [id]);

  const upload = async (file: File | undefined) => {
    if (!file) return;
    setError(null);
    if (file.size > MAX_FILE_BYTES) {
      setError(t("ide.createPod.configFileTooLarge"));
      return;
    }
    try {
      setBusy(true);
      const parsed = JSON.parse(await file.text()) as unknown;
      if (parsed === null || typeof parsed !== "object" || Array.isArray(parsed)) {
        throw new Error(t("ide.createPod.configFileMustBeObject"));
      }
      const bundle = await createEnvBundle({
        agentSlug,
        name: file.name,
        description: t("ide.createPod.configFileUploaded"),
        kind: "config",
        data: { [CONFIG_BUNDLE_JSON_KEY]: JSON.stringify(parsed) },
      });
      const summary: EnvBundleSummary = {
        id: Number(bundle.id),
        name: bundle.name,
        agent_slug: bundle.agentSlug,
        kind: bundle.kind,
        kind_primary: bundle.kindPrimary,
      };
      setUploaded((items) => [...items.filter((item) => item.id !== summary.id), summary]);
      onChange([summary.id]);
    } catch (uploadError) {
      setError(uploadError instanceof Error
        ? uploadError.message
        : t("ide.createPod.configFileUploadFailed"));
    } finally {
      setBusy(false);
    }
  };

  return (
    <section className="space-y-2">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-sm font-medium">{t("ide.createPod.selectConfigFile")}</p>
          <p className="text-xs text-muted-foreground">
            {t("ide.createPod.selectConfigFileHint")}
          </p>
        </div>
        <label className="inline-flex cursor-pointer items-center gap-1 text-xs font-medium text-primary hover:underline">
          <Upload className="h-3.5 w-3.5" />
          {busy ? t("common.saving") : t("ide.createPod.uploadConfigFile")}
          <input
            type="file"
            accept=".json,application/json"
            className="sr-only"
            disabled={busy}
            onChange={(event) => {
              void upload(event.target.files?.[0]);
              event.currentTarget.value = "";
            }}
          />
        </label>
      </div>
      {busy && <Spinner size="sm" />}
      {error && <p role="alert" className="text-xs text-destructive">{error}</p>}
      {selectedId !== undefined && (
        <div className="flex flex-wrap gap-1.5">
          <span className="inline-flex items-center gap-1 rounded-md border border-border px-2 py-1 text-xs">
            <FileJson className="h-3 w-3 text-primary" />
            {allBundles.find((item) => item.id === selectedId)?.name ?? `#${selectedId}`}
            <button type="button" onClick={() => select(selectedId)} aria-label={t("common.delete")}>
              <X className="h-3 w-3 text-muted-foreground" />
            </button>
          </span>
        </div>
      )}
      {allBundles.length > 0 && (
        <div className="surface-card max-h-32 overflow-y-auto">
          {allBundles.map((bundle) => (
            <label key={bundle.id} className="flex items-center gap-2 border-b border-border px-2 py-1.5 last:border-b-0">
              <input
                type="checkbox"
                checked={selectedId === bundle.id}
                onChange={() => select(bundle.id)}
              />
              <FileJson className="h-4 w-4 text-muted-foreground" />
              <span className="min-w-0 flex-1 truncate text-sm">{bundle.name}</span>
            </label>
          ))}
        </div>
      )}
    </section>
  );
}
