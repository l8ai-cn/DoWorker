"use client";

import { useCallback, useEffect, useState } from "react";
import type { AgentData } from "@/lib/api";
import { listAgents } from "@/lib/api/facade/agentConnect";
import { useCurrentOrg } from "@/stores/auth";
import type {
  AgentConfigState,
  AgentConfigActions,
  RuntimeBundleFormData,
  RuntimeBundleViewModel,
  CredentialBundleFormData,
  ConfigFileFormData,
  ConfigFileBundleViewModel,
} from "./types";
import { useAgentConfigMessages } from "./useAgentConfigMessages";
import { useRuntimeBundles } from "./useRuntimeBundles";
import { useCredentialBundles } from "./useCredentialBundles";
import { useConfigFileBundles } from "./useConfigFileBundles";
import { useAgentRuntimeConfig } from "./useAgentRuntimeConfig";
import type { CredentialProfileViewModel } from "../_shared/credentialViewModel";

export function useAgentConfig(
  agentSlug: string,
  t: (key: string) => string
): AgentConfigState & AgentConfigActions {
  const [loading, setLoading] = useState(true);
  const [agent, setAgent] = useState<AgentData | null>(null);
  const currentOrg = useCurrentOrg();

  const msgs = useAgentConfigMessages();
  const runtime = useRuntimeBundles(msgs, t);
  const credentials = useCredentialBundles(msgs, t);
  const configFiles = useConfigFileBundles(msgs, t);
  const cfg = useAgentRuntimeConfig(msgs, t);

  const loadData = useCallback(async () => {
    if (!currentOrg) {
      setLoading(false);
      return;
    }
    setLoading(true);
    msgs.setError(null);

    try {
      const agentsRes = await listAgents(currentOrg.slug);
      const allAgents: AgentData[] = [
        ...agentsRes.builtin_agents,
        ...agentsRes.custom_agents,
        ...agentsRes.agents,
      ];
      const found = allAgents.find((a) => a.slug === agentSlug);
      if (!found) {
        msgs.setError(t("settings.agentConfig.agentNotFound"));
        setAgent(null);
        return;
      }
      setAgent(found);
      await Promise.all([
        credentials.loadCredentialBundles(found),
        runtime.loadRuntimeBundles(found),
        configFiles.loadConfigFileBundles(found),
        cfg.loadRuntimeConfig(found),
      ]);
    } catch (err) {
      msgs.reportError(err, t, "settings.agentConfig.loadFailed");
    } finally {
      setLoading(false);
    }
  }, [agentSlug, t, credentials, runtime, configFiles, cfg, msgs, currentOrg]);

  useEffect(() => {
    loadData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [agentSlug, currentOrg?.slug]);

  const handleSaveRuntimeBundle = useCallback(
    (data: RuntimeBundleFormData, editingBundle: RuntimeBundleViewModel | null) => {
      if (!agent) return Promise.resolve();
      return runtime.handleSaveRuntimeBundle(data, editingBundle, agent);
    },
    [agent, runtime]
  );

  const handleSaveCredentialBundle = useCallback(
    (data: CredentialBundleFormData, editing: CredentialProfileViewModel | null) => {
      if (!agent) return Promise.resolve();
      return credentials.handleSaveCredentialBundle(data, editing, agent);
    },
    [agent, credentials]
  );

  const handleSaveConfigFileBundle = useCallback(
    (data: ConfigFileFormData, editing: ConfigFileBundleViewModel | null) => {
      if (!agent) return Promise.resolve();
      return configFiles.handleSaveConfigFileBundle(data, editing, agent);
    },
    [agent, configFiles]
  );

  const handleSaveConfig = useCallback(() => {
    if (!agent) return Promise.resolve();
    return cfg.handleSaveConfig(agent);
  }, [agent, cfg]);

  return {
    loading,
    savingConfig: cfg.savingConfig,
    agent,
    configFields: cfg.configFields,
    configValues: cfg.configValues,
    credentialFields: credentials.credentialFields,
    credentialBundles: credentials.credentialBundles,
    runtimeBundles: runtime.runtimeBundles,
    configFileSpecs: configFiles.configFileSpecs,
    configFileBundles: configFiles.configFileBundles,
    error: msgs.error,
    success: msgs.success,
    handleSetRuntimePrimary: runtime.handleSetRuntimePrimary,
    handleClearRuntimePrimary: runtime.handleClearRuntimePrimary,
    handleDeleteRuntimeBundle: runtime.handleDeleteRuntimeBundle,
    handleSaveRuntimeBundle,
    handleSetConfigPrimary: configFiles.handleSetConfigPrimary,
    handleClearConfigPrimary: configFiles.handleClearConfigPrimary,
    handleDeleteConfigFileBundle: configFiles.handleDeleteConfigFileBundle,
    handleSaveConfigFileBundle,
    handleConfigChange: cfg.handleConfigChange,
    handleSaveConfig,
    handleSetCredentialPrimary: credentials.handleSetCredentialPrimary,
    handleClearCredentialPrimary: credentials.handleClearCredentialPrimary,
    handleDeleteCredentialBundle: credentials.handleDeleteCredentialBundle,
    handleSaveCredentialBundle,
    setError: msgs.setError,
    setSuccess: msgs.setSuccess,
    loadData,
  };
}
