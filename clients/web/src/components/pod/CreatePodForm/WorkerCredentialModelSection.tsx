"use client";

import { KeyRound } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import { ConfigForm } from "@/components/ide/ConfigForm";
import { CredentialBundleSelect } from "./CredentialBundleSelect";
import { EnvBundleMultiSelect } from "./EnvBundleMultiSelect";
import type { ConfigField, EnvBundleSummary } from "@/lib/api";

interface WorkerCredentialModelSectionProps {
  agentSlug: string | null;
  envBundles: EnvBundleSummary[];
  loadingBundles: boolean;
  selectedCredentialName: string;
  onSelectCredential: (name: string) => void;
  selectedRuntimeBundleNames: string[];
  onSelectRuntimeBundles: (names: string[]) => void;
  configFields: ConfigField[];
  configValues: Record<string, unknown>;
  loadingConfig: boolean;
  onConfigChange: (key: string, value: unknown) => void;
  rawLayerMode: boolean;
  t: (key: string) => string;
}

/** API credentials, runtime env bundles, and model/plugin config for the selected image. */
export function WorkerCredentialModelSection({
  agentSlug,
  envBundles,
  loadingBundles,
  selectedCredentialName,
  onSelectCredential,
  selectedRuntimeBundleNames,
  onSelectRuntimeBundles,
  configFields,
  configValues,
  loadingConfig,
  onConfigChange,
  rawLayerMode,
  t,
}: WorkerCredentialModelSectionProps) {
  if (rawLayerMode || !agentSlug) return null;

  return (
    <div
      className="space-y-4 border-t border-border pt-4"
      data-testid="worker-credential-model-section"
    >
      <div className="flex gap-2">
        <KeyRound className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
        <div>
          <p className="text-sm font-medium">{t("ide.createPod.credentialModelTitle")}</p>
          <p className="text-xs leading-5 text-muted-foreground">
            {t("ide.createPod.credentialModelDescription")}
          </p>
        </div>
      </div>

      <CredentialBundleSelect
        bundles={envBundles.filter((b) => b.kind === "credential")}
        selectedBundleName={selectedCredentialName}
        onSelect={onSelectCredential}
        loading={loadingBundles}
        t={t}
      />

      <EnvBundleMultiSelect
        bundles={envBundles.filter((b) => b.kind === "runtime")}
        selectedBundleNames={selectedRuntimeBundleNames}
        onChange={onSelectRuntimeBundles}
        loading={loadingBundles}
        t={t}
      />

      {loadingConfig ? (
        <div className="flex items-center justify-center py-4">
          <Spinner size="sm" className="mr-2" />
          <span className="text-sm text-muted-foreground">
            {t("ide.createPod.loadingPlugins")}
          </span>
        </div>
      ) : (
        configFields.length > 0 && (
          <div>
            <label className="mb-2 block text-sm font-medium">
              {t("ide.createPod.pluginConfig")}
            </label>
            <ConfigForm
              fields={configFields}
              values={configValues}
              onChange={onConfigChange}
              agentSlug={agentSlug}
            />
          </div>
        )
      )}
    </div>
  );
}
