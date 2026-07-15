"use client";

import { useState, useCallback } from "react";
import { CenteredSpinner } from "@/components/ui/spinner";
import { AlertMessage } from "@/components/ui/alert-message";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { useTranslations } from "next-intl";
import { AlertCircle } from "lucide-react";
import { useAgentConfig } from "./useAgentConfig";
import { AgentConfigHeader } from "./AgentConfigHeader";
import { useAgentConfigDeleteActions } from "./useAgentConfigDeleteActions";
import { RuntimeConfigSection } from "./RuntimeConfigSection";
import { RuntimeBundlesSection } from "./RuntimeBundlesSection";
import { RuntimeBundleDialog } from "./RuntimeBundleDialog";
import { CredentialBundlesSection } from "./CredentialBundlesSection";
import { CredentialBundleDialog } from "./CredentialBundleDialog";
import { ConfigFilesSection } from "./ConfigFilesSection";
import { ConfigFileDialog } from "./ConfigFileDialog";
import type {
  AgentConfigPageProps,
  RuntimeBundleViewModel,
  ConfigFileBundleViewModel,
} from "./types";
import type { CredentialProfileViewModel } from "../_shared/credentialViewModel";

export function AgentConfigPage({ agentSlug }: AgentConfigPageProps) {
  const t = useTranslations();

  const [showRuntimeDialog, setShowRuntimeDialog] = useState(false);
  const [editingRuntime, setEditingRuntime] = useState<RuntimeBundleViewModel | null>(null);
  const [showCredentialDialog, setShowCredentialDialog] = useState(false);
  const [editingCredential, setEditingCredential] = useState<CredentialProfileViewModel | null>(null);
  const [showConfigDialog, setShowConfigDialog] = useState(false);
  const [editingConfig, setEditingConfig] = useState<ConfigFileBundleViewModel | null>(null);

  const {
    loading,
    savingConfig,
    agent,
    configFields,
    configValues,
    credentialFields,
    credentialBundles,
    runtimeBundles,
    configFileSpecs,
    configFileBundles,
    error,
    success,
    handleConfigChange,
    handleSaveConfig,
    handleSetCredentialPrimary,
    handleClearCredentialPrimary,
    handleDeleteCredentialBundle,
    handleSaveCredentialBundle,
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

  const {
    dialogProps,
    deleteRuntime,
    deleteCredential,
    deleteConfigFile,
  } = useAgentConfigDeleteActions({
    deleteRuntime: handleDeleteRuntimeBundle,
    deleteCredential: handleDeleteCredentialBundle,
    deleteConfigFile: handleDeleteConfigFileBundle,
    t,
  });

  const handleOpenAddRuntime = useCallback(() => {
    setEditingRuntime(null);
    setShowRuntimeDialog(true);
  }, []);

  const handleOpenEditRuntime = useCallback((b: RuntimeBundleViewModel) => {
    setEditingRuntime(b);
    setShowRuntimeDialog(true);
  }, []);

  const handleOpenAddCredential = useCallback(() => {
    setEditingCredential(null);
    setShowCredentialDialog(true);
  }, []);

  const handleOpenEditCredential = useCallback((bundle: CredentialProfileViewModel) => {
    setEditingCredential(bundle);
    setShowCredentialDialog(true);
  }, []);

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
      <AgentConfigHeader agent={agent} />
      {error && <AlertMessage type="error" message={error} onDismiss={() => setError(null)} />}
      {success && <AlertMessage type="success" message={success} onDismiss={() => setSuccess(null)} />}
      <RuntimeConfigSection
        configFields={configFields}
        configValues={configValues}
        agentSlug={agentSlug}
        saving={savingConfig}
        onChange={handleConfigChange}
        onSave={handleSaveConfig}
        t={t}
      />

      {credentialFields.length > 0 && (
        <CredentialBundlesSection
          bundles={credentialBundles}
          onSetDefault={handleSetCredentialPrimary}
          onClearDefault={handleClearCredentialPrimary}
          onEdit={handleOpenEditCredential}
          onDelete={deleteCredential}
          onAdd={handleOpenAddCredential}
          t={t}
        />
      )}

      <RuntimeBundlesSection
        bundles={runtimeBundles}
        onSetDefault={handleSetRuntimePrimary}
        onClearDefault={handleClearRuntimePrimary}
        onEdit={handleOpenEditRuntime}
        onDelete={deleteRuntime}
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
        onDelete={deleteConfigFile}
        onAdd={() => {
          setEditingConfig(null);
          setShowConfigDialog(true);
        }}
        t={t}
      />

      <RuntimeBundleDialog
        open={showRuntimeDialog}
        onOpenChange={setShowRuntimeDialog}
        editingBundle={editingRuntime}
        onSubmit={handleSaveRuntimeBundle}
        t={t}
      />

      {credentialFields.length > 0 && (
        <CredentialBundleDialog
          open={showCredentialDialog}
          onOpenChange={setShowCredentialDialog}
          agentSlug={agentSlug}
          credentialFields={credentialFields}
          editing={editingCredential}
          onSubmit={handleSaveCredentialBundle}
          t={t}
        />
      )}

      <ConfigFileDialog
        open={showConfigDialog}
        onOpenChange={setShowConfigDialog}
        editing={editingConfig}
        fileSpecs={configFileSpecs}
        onSubmit={handleSaveConfigFileBundle}
        t={t}
      />

      <ConfirmDialog {...dialogProps} />
    </div>
  );
}

export default AgentConfigPage;
