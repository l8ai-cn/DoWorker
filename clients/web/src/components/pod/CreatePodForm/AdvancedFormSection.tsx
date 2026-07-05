"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Spinner } from "@/components/ui/spinner";
import { ConfigForm } from "@/components/ide/ConfigForm";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { CredentialBundleSelect } from "./CredentialBundleSelect";
import { EnvBundleMultiSelect } from "./EnvBundleMultiSelect";
import { SkillMultiSelect } from "./SkillMultiSelect";
import { RepositorySelect, BranchInput } from "./RepositorySelect";
import { AdvancedOptions } from "./AdvancedOptions";
import { AgentfileLayerEditor } from "./AgentfileLayerEditor";
import { KnowledgeBaseMountSelect } from "./KnowledgeBaseMountSelect";
import { PodLifecycleSection } from "./PodLifecycleSection";
import type { CreatePodFormState } from "../hooks";
import type { RepositoryData, ConfigField } from "@/lib/api";

interface AdvancedFormSectionProps {
  form: CreatePodFormState;
  repositories: RepositoryData[];
  configFields: ConfigField[];
  loadingConfig: boolean;
  configValues: Record<string, unknown>;
  handleConfigChange: (key: string, value: unknown) => void;
  showPerpetual?: boolean;
}

export function AdvancedFormSection({
  form,
  repositories,
  configFields,
  loadingConfig,
  configValues,
  handleConfigChange,
  showPerpetual,
}: AdvancedFormSectionProps) {
  const t = useTranslations();

  const hideFormSections = form.rawLayerMode;

  return (
    <AdvancedOptions t={t}>
      {/* Pod Alias (optional display name) — always visible */}
      <div>
        <label htmlFor="pod-alias" className="block text-sm font-medium mb-1">
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

      {/* Perpetual Pod (auto-restart on exit) — workspace only */}
      {showPerpetual && (
        <div className="flex items-center justify-between">
          <div>
            <label htmlFor="pod-perpetual" className="text-sm font-medium">
              {t("ide.createPod.perpetual")}
            </label>
            <p className="text-xs text-muted-foreground">
              {t("ide.createPod.perpetualDescription")}
            </p>
          </div>
          <Switch
            id="pod-perpetual"
            checked={form.perpetual}
            onCheckedChange={form.setPerpetual}
          />
        </div>
      )}

      <PodLifecycleSection
        destroyPolicy={form.destroyPolicy}
        destroyAfterMinutes={form.destroyAfterMinutes}
        onPolicyChange={form.setDestroyPolicy}
        onAfterChange={form.setDestroyAfterMinutes}
      />

      {/* Form-mode-only sections (hidden when source mode is ON) */}
      {!hideFormSections && (
        <>
          {/* API Credential — single-select dropdown */}
          <CredentialBundleSelect
            bundles={form.envBundles.filter((b) => b.kind === "credential")}
            selectedBundleName={form.selectedCredentialName}
            onSelect={form.setSelectedCredentialName}
            loading={form.loadingBundles}
            t={t}
          />

          {/* Runtime EnvBundle — ordered multi-select */}
          <EnvBundleMultiSelect
            bundles={form.envBundles.filter((b) => b.kind === "runtime")}
            selectedBundleNames={form.selectedRuntimeBundleNames}
            onChange={form.setSelectedRuntimeBundleNames}
            loading={form.loadingBundles}
            t={t}
          />

          {/* Repository Select */}
          <RepositorySelect
            repositories={repositories}
            selectedRepositoryId={form.selectedRepository}
            onSelect={form.setSelectedRepository}
            t={t}
          />

          {/* Branch Input */}
          {form.selectedRepository && (
            <BranchInput
              value={form.selectedBranch}
              onChange={form.setSelectedBranch}
              error={form.validationErrors.branch}
              t={t}
            />
          )}

          <KnowledgeBaseMountSelect
            selectedMounts={form.selectedKnowledgeMounts}
            onChange={form.setSelectedKnowledgeMounts}
          />

          <SkillMultiSelect
            skills={form.repoSkills}
            selectedSlugs={form.selectedSkillSlugs}
            onChange={form.setSelectedSkillSlugs}
            loading={form.loadingSkills}
            repositorySelected={Boolean(form.selectedRepository)}
            t={t}
          />

          {/* Agent Configuration Section */}
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
                <label className="block text-sm font-medium mb-2">
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

      {/* AgentFile Layer Editor — always visible */}
      <AgentfileLayerEditor
        generatedLayer={form.agentfileLayer}
        rawMode={form.rawLayerMode}
        rawText={form.rawLayerText}
        onRawModeChange={form.setRawLayerMode}
        onRawTextChange={form.setRawLayerText}
        configFields={configFields}
        repositories={repositories}
        envBundles={form.envBundles}
        t={t}
      />
    </AdvancedOptions>
  );
}
