"use client";

import { KeyRound } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import { ConfigForm } from "@/components/ide/ConfigForm";
import { EnvBundleMultiSelect } from "./EnvBundleMultiSelect";
import { WorkerModelResourceSelect } from "./WorkerModelResourceSelect";
import type { ConfigField, EnvBundleSummary } from "@/lib/api";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";

interface WorkerCredentialModelSectionProps {
  agentSlug: string | null;
  modelResources: EffectiveResource[];
  selectedModelResourceId: number | null;
  onSelectModelResource: (id: number | null) => void;
  loadingModelResources: boolean;
  modelResourceError: string | null;
  modelResourceValidationError?: string;
  envBundles: EnvBundleSummary[];
  loadingBundles: boolean;
  bundleLoadError: string | null;
  runtimeBundleValidationError?: string;
  selectedRuntimeBundleNames: string[];
  onSelectRuntimeBundles: (names: string[]) => void;
  configFields: ConfigField[];
  configValues: Record<string, unknown>;
  loadingConfig: boolean;
  onConfigChange: (key: string, value: unknown) => void;
  rawLayerMode: boolean;
  t: (key: string) => string;
}

/** Model resource, runtime env bundles, and image config for the selected Worker. */
export function WorkerCredentialModelSection({
  agentSlug,
  modelResources,
  selectedModelResourceId,
  onSelectModelResource,
  loadingModelResources,
  modelResourceError,
  modelResourceValidationError,
  envBundles,
  loadingBundles,
  bundleLoadError,
  runtimeBundleValidationError,
  selectedRuntimeBundleNames,
  onSelectRuntimeBundles,
  configFields,
  configValues,
  loadingConfig,
  onConfigChange,
  rawLayerMode,
  t,
}: WorkerCredentialModelSectionProps) {
  if (!agentSlug) return null;

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

      <WorkerModelResourceSelect
        resources={modelResources}
        selectedResourceId={selectedModelResourceId}
        onSelect={onSelectModelResource}
        loading={loadingModelResources}
        error={modelResourceError}
        validationError={modelResourceValidationError}
        t={t}
      />

      {!rawLayerMode && (
        <>
          <EnvBundleMultiSelect
            bundles={envBundles}
            selectedBundleNames={selectedRuntimeBundleNames}
            onChange={onSelectRuntimeBundles}
            loading={loadingBundles}
            error={bundleLoadError}
            validationError={runtimeBundleValidationError}
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
        </>
      )}
    </div>
  );
}
