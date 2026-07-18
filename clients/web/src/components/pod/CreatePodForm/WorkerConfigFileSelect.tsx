"use client";

import { useState } from "react";
import { FileJson, Upload, X } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import { createEnvBundle } from "@/lib/api/facade/envBundleConnect";
import type {
  WorkerConfigDocumentBinding,
  WorkerConfigDocumentRequirement,
} from "@/lib/api/facade/podConnect";
import type { EnvBundleSummary } from "@/lib/viewModels/envBundleSummary";
import { CONFIG_BUNDLE_JSON_KEY } from "@/components/settings/AgentConfigPage/configBundleKeys";

interface WorkerConfigFileSelectProps {
  agentSlug: string;
  requirements: WorkerConfigDocumentRequirement[];
  bundles: EnvBundleSummary[];
  bindings: WorkerConfigDocumentBinding[];
  onChange: (bindings: WorkerConfigDocumentBinding[]) => void;
  t: (key: string) => string;
}

const MAX_FILE_BYTES = 1024 * 1024;

export function WorkerConfigFileSelect({
  agentSlug,
  requirements,
  bundles,
  bindings,
  onChange,
  t,
}: WorkerConfigFileSelectProps) {
  const [uploaded, setUploaded] = useState<EnvBundleSummary[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const allBundles = [...bundles, ...uploaded.filter((item) =>
    !bundles.some((bundle) => bundle.id === item.id),
  )];

  const select = (documentID: string, bundleID: number) => {
    const selected = bindings.find(
      (binding) => binding.document_id === documentID,
    );
    const next = bindings.filter(
      (binding) => binding.document_id !== documentID,
    );
    if (selected?.config_bundle_id !== bundleID) {
      next.push({ document_id: documentID, config_bundle_id: bundleID });
    }
    onChange(next);
  };

  const upload = async (
    requirement: WorkerConfigDocumentRequirement,
    file: File | undefined,
  ) => {
    if (!file) return;
    setError(null);
    if (requirement.format !== "json") {
      setError(`${requirement.document_id}: unsupported format ${requirement.format}`);
      return;
    }
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
      select(requirement.document_id, summary.id);
    } catch (uploadError) {
      setError(uploadError instanceof Error
        ? uploadError.message
        : t("ide.createPod.configFileUploadFailed"));
    } finally {
      setBusy(false);
    }
  };

  if (requirements.length === 0) return null;

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-sm font-medium">{t("ide.createPod.selectConfigFile")}</p>
          <p className="text-xs text-muted-foreground">
            {t("ide.createPod.selectConfigFileHint")}
          </p>
        </div>
      </div>
      {busy && <Spinner size="sm" />}
      {error && <p role="alert" className="text-xs text-destructive">{error}</p>}
      {requirements.map((requirement) => {
        const selectedID = bindings.find(
          (binding) => binding.document_id === requirement.document_id,
        )?.config_bundle_id;
        return (
          <section
            key={requirement.document_id}
            className="space-y-2 border-l-2 border-border pl-3"
          >
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="text-sm font-medium">{requirement.document_id}</p>
                <p className="text-xs text-muted-foreground">
                  {requirement.format} · {requirement.target_path}
                </p>
              </div>
              <label className="inline-flex cursor-pointer items-center gap-1 text-xs font-medium text-primary hover:underline">
                <Upload className="h-3.5 w-3.5" />
                {busy ? t("common.saving") : t("ide.createPod.uploadConfigFile")}
                <input
                  type="file"
                  accept=".json,application/json"
                  className="sr-only"
                  disabled={busy || requirement.format !== "json"}
                  onChange={(event) => {
                    void upload(requirement, event.target.files?.[0]);
                    event.currentTarget.value = "";
                  }}
                />
              </label>
            </div>
            {selectedID !== undefined && (
              <div className="flex flex-wrap gap-1.5">
                <span className="inline-flex items-center gap-1 rounded-md border border-border px-2 py-1 text-xs">
                  <FileJson className="h-3 w-3 text-primary" />
                  {allBundles.find((item) => item.id === selectedID)?.name ?? `#${selectedID}`}
                  <button
                    type="button"
                    onClick={() => select(requirement.document_id, selectedID)}
                    aria-label={t("common.delete")}
                  >
                    <X className="h-3 w-3 text-muted-foreground" />
                  </button>
                </span>
              </div>
            )}
            {allBundles.length > 0 && (
              <div className="surface-card max-h-32 overflow-y-auto">
                {allBundles.map((bundle) => {
                  const usedByOtherDocument = bindings.some(
                    (binding) =>
                      binding.document_id !== requirement.document_id &&
                      binding.config_bundle_id === bundle.id,
                  );
                  return (
                    <label
                      key={bundle.id}
                      className="flex items-center gap-2 border-b border-border px-2 py-1.5 last:border-b-0"
                    >
                      <input
                        type="checkbox"
                        checked={selectedID === bundle.id}
                        disabled={usedByOtherDocument}
                        onChange={() => select(requirement.document_id, bundle.id)}
                      />
                      <FileJson className="h-4 w-4 text-muted-foreground" />
                      <span className="min-w-0 flex-1 truncate text-sm">{bundle.name}</span>
                    </label>
                  );
                })}
              </div>
            )}
          </section>
        );
      })}
    </section>
  );
}
