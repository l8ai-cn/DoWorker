"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Spinner } from "@/components/ui/spinner";
import { ConfigForm } from "@/components/ide/ConfigForm";
import { Input } from "@/components/ui/input";
import { CredentialBundleSelect } from "./CredentialBundleSelect";
import { EnvBundleMultiSelect } from "./EnvBundleMultiSelect";
import { PodLifecycleSection } from "./PodLifecycleSection";
import type { CreatePodFormState } from "../hooks";
import type { ConfigField } from "@/lib/api";

interface AdvancedFormSectionProps {
  form: CreatePodFormState;
  configFields: ConfigField[];
  loadingConfig: boolean;
  configValues: Record<string, unknown>;
  handleConfigChange: (key: string, value: unknown) => void;
}

export function AdvancedFormSection({
  form,
  configFields,
  loadingConfig,
  configValues,
  handleConfigChange,
}: AdvancedFormSectionProps) {
  const t = useTranslations();

  return (
    <div className="space-y-4">
      <div>
        <label htmlFor="pod-alias" className="mb-1 block text-sm font-medium">
          {t("ide.createPod.alias")}
        </label>
        <Input
          id="pod-alias"
          value={form.alias}
          onChange={(e) => form.setAlias(e.target.value)}
          placeholder={t("ide.createPod.aliasPlaceholder")}
          maxLength={100}
        />
      </div>

      <PodLifecycleSection
        destroyPolicy={form.destroyPolicy}
        destroyAfterMinutes={form.destroyAfterMinutes}
        onPolicyChange={form.setDestroyPolicy}
        onAfterChange={form.setDestroyAfterMinutes}
      />

      {!form.rawLayerMode && (
        <>
          <CredentialBundleSelect
            bundles={form.envBundles.filter((b) => b.kind === "credential")}
            selectedBundleName={form.selectedCredentialName}
            onSelect={form.setSelectedCredentialName}
            loading={form.loadingBundles}
            t={t}
          />

          <EnvBundleMultiSelect
            bundles={form.envBundles.filter((b) => b.kind === "runtime")}
            selectedBundleNames={form.selectedRuntimeBundleNames}
            onChange={form.setSelectedRuntimeBundleNames}
            loading={form.loadingBundles}
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
                  onChange={handleConfigChange}
                  agentSlug={form.selectedAgentSlug}
                />
              </div>
            )
          )}
        </>
      )}
    </div>
  );
}
