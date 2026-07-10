"use client";

import { useState, useCallback } from "react";
import { CenteredSpinner } from "@/components/ui/spinner";
import { AlertMessage } from "@/components/ui/alert-message";
import { ConfirmDialog, useConfirmDialog } from "@/components/ui/confirm-dialog";
import { useTranslations } from "next-intl";
import { Bot, AlertCircle } from "lucide-react";
import { useAgentConfig } from "./useAgentConfig";
import { RuntimeConfigSection } from "./RuntimeConfigSection";
import { RuntimeBundlesSection } from "./RuntimeBundlesSection";
import { RuntimeBundleDialog } from "./RuntimeBundleDialog";
import { ConfigFilesSection } from "./ConfigFilesSection";
import { ConfigFileDialog } from "./ConfigFileDialog";
import type {
  AgentConfigPageProps,
  RuntimeBundleViewModel,
  ConfigFileBundleViewModel,
} from "./types";

export function AgentConfigPage({ agentSlug }: AgentConfigPageProps) {
  const t = useTranslations();

  const [showRuntimeDialog, setShowRuntimeDialog] = useState(false);
  const [editingRuntime, setEditingRuntime] = useState<RuntimeBundleViewModel | null>(null);
  const [showConfigDialog, setShowConfigDialog] = useState(false);
  const [editingConfig, setEditingConfig] = useState<ConfigFileBundleViewModel | null>(null);

  const {
    loading,
    savingConfig,
    agent,
    configFields,
    configValues,
    runtimeBundles,
    configFileSpecs,
    configFileBundles,
    error,
    success,
    handleConfigChange,
    handleSaveConfig,
    handleSetRuntimePrimary,
    handleClearRuntimePrimary,
    handleDeleteRuntimeBundle,
    handleSaveRuntimeBundle,
    handleSetConfigPrimary,
    handleClearConfigPrimary,
    handleDeleteConfigFileBundle,
    handleSaveConfigFileBundle,
    setError,
    setSuccess,
  } = useAgentConfig(agentSlug, t);

  const { dialogProps, confirm } = useConfirmDialog();

  const handleOpenAddRuntime = useCallback(() => {
    setEditingRuntime(null);
    setShowRuntimeDialog(true);
  }, []);

  const handleOpenEditRuntime = useCallback((b: RuntimeBundleViewModel) => {
    setEditingRuntime(b);
    setShowRuntimeDialog(true);
  }, []);

  const handleDeleteRuntimeWithConfirm = useCallback(async (id: number) => {
    const confirmed = await confirm({
      title: t("common.confirmDelete"),
      description: t("settings.agentConfig.runtimeBundles.confirmDelete"),
      variant: "destructive",
      confirmText: t("common.delete"),
      cancelText: t("common.cancel"),
    });
    if (confirmed) {
      await handleDeleteRuntimeBundle(id);
    }
  }, [confirm, handleDeleteRuntimeBundle, t]);

  const handleDeleteConfigWithConfirm = useCallback(async (id: number) => {
    const confirmed = await confirm({
      title: t("common.confirmDelete"),
      description: t("settings.agentConfig.configFiles.confirmDelete"),
      variant: "destructive",
      confirmText: t("common.delete"),
      cancelText: t("common.cancel"),
    });
    if (confirmed) {
      await handleDeleteConfigFileBundle(id);
    }
  }, [confirm, handleDeleteConfigFileBundle, t]);

  if (loading) {
    return <CenteredSpinner className="py-12" />;
  }

  if (!agent) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <AlertCircle className="w-12 h-12 text-muted-foreground mb-4" />
        <p className="text-muted-foreground">{error || t("settings.agentConfig.agentNotFound")}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Bot className="w-8 h-8 text-primary" />
        <div>
          <h2 className="text-xl font-semibold">{agent.name}</h2>
          {agent.description && (
            <p className="text-sm text-muted-foreground">{agent.description}</p>
          )}
        </div>
      </div>

      {/* Error/Success messages */}
      {error && <AlertMessage type="error" message={error} onDismiss={() => setError(null)} />}
      {success && <AlertMessage type="success" message={success} onDismiss={() => setSuccess(null)} />}

      {/* Runtime Config Section */}
      <RuntimeConfigSection
        configFields={configFields}
        configValues={configValues}
        agentSlug={agentSlug}
        saving={savingConfig}
        onChange={handleConfigChange}
        onSave={handleSaveConfig}
        t={t}
      />

      {/* Runtime EnvBundles Section — plaintext KV preferences attached
          to this agent (model overrides, log levels, etc.). */}
      <RuntimeBundlesSection
        bundles={runtimeBundles}
        onSetDefault={handleSetRuntimePrimary}
        onClearDefault={handleClearRuntimePrimary}
        onEdit={handleOpenEditRuntime}
        onDelete={handleDeleteRuntimeWithConfirm}
        onAdd={handleOpenAddRuntime}
        t={t}
      />

      <ConfigFilesSection
        bundles={configFileBundles}
        fileSpecs={configFileSpecs}
        onSetDefault={handleSetConfigPrimary}
        onClearDefault={handleClearConfigPrimary}
        onEdit={(b) => {
          setEditingConfig(b);
          setShowConfigDialog(true);
        }}
        onDelete={handleDeleteConfigWithConfirm}
        onAdd={() => {
          setEditingConfig(null);
          setShowConfigDialog(true);
        }}
        t={t}
      />

      {/* Add/Edit Runtime Bundle Dialog */}
      <RuntimeBundleDialog
        open={showRuntimeDialog}
        onOpenChange={setShowRuntimeDialog}
        editingBundle={editingRuntime}
        onSubmit={handleSaveRuntimeBundle}
        t={t}
      />

      <ConfigFileDialog
        open={showConfigDialog}
        onOpenChange={setShowConfigDialog}
        editing={editingConfig}
        fileSpecs={configFileSpecs}
        onSubmit={handleSaveConfigFileBundle}
        t={t}
      />

      {/* Confirm Dialog */}
      <ConfirmDialog {...dialogProps} />
    </div>
  );
}

export default AgentConfigPage;
