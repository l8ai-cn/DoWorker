import { useCallback, useState } from "react";
import type { AgentData, ConfigFile } from "@/lib/api";
import {
  listEnvBundles,
  createEnvBundle,
  updateEnvBundle,
  deleteEnvBundle,
  setPrimaryEnvBundle,
} from "@/lib/api/facade/envBundleConnect";
import { getAgentConfigSchema } from "@/lib/api/facade/agentConnect";
import { useCurrentOrg } from "@/stores/auth";
import type { ConfigFileBundleViewModel, ConfigFileFormData } from "./types";
import type { AgentConfigMessages } from "./useAgentConfigMessages";
import { toConfigFileBundle } from "./envBundleWire";
import { CONFIG_BUNDLE_JSON_KEY } from "./configBundleKeys";

export function useConfigFileBundles(
  msgs: AgentConfigMessages,
  t: (key: string) => string
) {
  const [configFileSpecs, setConfigFileSpecs] = useState<ConfigFile[]>([]);
  const [configFileBundles, setConfigFileBundles] = useState<ConfigFileBundleViewModel[]>([]);
  const currentOrg = useCurrentOrg();

  const loadConfigFileBundles = useCallback(async (agent: AgentData) => {
    const orgSlug = currentOrg?.slug;
    if (orgSlug) {
      const schema = await getAgentConfigSchema(orgSlug, agent.slug).catch(() => ({
        config_files: [] as ConfigFile[],
      }));
      setConfigFileSpecs(schema.config_files ?? []);
    } else {
      setConfigFileSpecs([]);
    }

    try {
      const res = await listEnvBundles({ kind: "config", agentSlug: agent.slug }).catch(
        () => ({ items: [] })
      );
      setConfigFileBundles((res.items ?? []).map((b) => toConfigFileBundle(b, agent.slug)));
    } catch (err) {
      msgs.reportError(err, t, "settings.agentConfig.loadFailed");
    }
  }, [currentOrg, msgs, t]);

  const handleSetConfigPrimary = useCallback(async (id: number) => {
    try {
      msgs.setError(null);
      await setPrimaryEnvBundle(BigInt(id));
      setConfigFileBundles((prev) => prev.map((b) => ({ ...b, is_default: b.id === id })));
      msgs.reportSuccess(t("settings.agentConfig.configFiles.defaultSet"));
    } catch (err) {
      msgs.reportError(err, t, "settings.agentConfig.configFiles.failedToSetDefault");
    }
  }, [msgs, t]);

  const handleClearConfigPrimary = useCallback(async () => {
    try {
      msgs.setError(null);
      const current = configFileBundles.find((b) => b.is_default);
      if (current) {
        await updateEnvBundle(BigInt(current.id), { kindPrimary: false });
      }
      setConfigFileBundles((prev) => prev.map((b) => ({ ...b, is_default: false })));
      msgs.reportSuccess(t("settings.agentConfig.configFiles.defaultSet"));
    } catch (err) {
      msgs.reportError(err, t, "settings.agentConfig.configFiles.failedToSetDefault");
    }
  }, [configFileBundles, msgs, t]);

  const handleDeleteConfigFileBundle = useCallback(async (id: number) => {
    try {
      msgs.setError(null);
      await deleteEnvBundle(BigInt(id));
      setConfigFileBundles((prev) => prev.filter((b) => b.id !== id));
      msgs.reportSuccess(t("settings.agentConfig.configFiles.deleted"));
    } catch (err) {
      msgs.reportError(err, t, "settings.agentConfig.configFiles.failedToDelete");
    }
  }, [msgs, t]);

  const handleSaveConfigFileBundle = useCallback(
    async (data: ConfigFileFormData, editing: ConfigFileBundleViewModel | null, agent: AgentData) => {
      const payload = { [CONFIG_BUNDLE_JSON_KEY]: data.jsonContent };
      if (editing) {
        await updateEnvBundle(BigInt(editing.id), {
          name: data.name,
          description: data.description || undefined,
          hasData: true,
          data: payload,
        });
        msgs.reportSuccess(t("settings.agentConfig.configFiles.updated"));
      } else {
        await createEnvBundle({
          agentSlug: agent.slug,
          name: data.name,
          description: data.description || undefined,
          kind: "config",
          data: payload,
        });
        msgs.reportSuccess(t("settings.agentConfig.configFiles.created"));
      }
      await loadConfigFileBundles(agent);
    },
    [loadConfigFileBundles, msgs, t]
  );

  return {
    configFileSpecs,
    configFileBundles,
    loadConfigFileBundles,
    handleSetConfigPrimary,
    handleClearConfigPrimary,
    handleDeleteConfigFileBundle,
    handleSaveConfigFileBundle,
  };
}
