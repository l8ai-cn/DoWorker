import { useCallback, useState } from "react";
import type { AgentData, CredentialField } from "@/lib/api";
import {
  createEnvBundle,
  deleteEnvBundle,
  listEnvBundles,
  setPrimaryEnvBundle,
  updateEnvBundle,
} from "@/lib/api/facade/envBundleConnect";
import { getAgentConfigSchema } from "@/lib/api/facade/agentConnect";
import { useCurrentOrg } from "@/stores/auth";
import type { CredentialProfileViewModel } from "../_shared/credentialViewModel";
import { toCredentialProfile } from "./envBundleWire";
import type { CredentialBundleFormData } from "./types";
import type { AgentConfigMessages } from "./useAgentConfigMessages";

export function useCredentialBundles(
  msgs: AgentConfigMessages,
  t: (key: string) => string,
) {
  const [credentialFields, setCredentialFields] = useState<CredentialField[]>([]);
  const [credentialBundles, setCredentialBundles] = useState<CredentialProfileViewModel[]>([]);
  const currentOrg = useCurrentOrg();

  const loadCredentialBundles = useCallback(async (agent: AgentData) => {
    if (!currentOrg) {
      setCredentialFields([]);
      setCredentialBundles([]);
      return;
    }
    try {
      const [schema, bundles] = await Promise.all([
        getAgentConfigSchema(currentOrg.slug, agent.slug),
        listEnvBundles({ kind: "credential", agentSlug: agent.slug }),
      ]);
      setCredentialFields(schema.credential_fields ?? []);
      setCredentialBundles(bundles.items.map((bundle) => toCredentialProfile(bundle, agent.slug)));
    } catch (error) {
      setCredentialFields([]);
      setCredentialBundles([]);
      msgs.reportError(error, t, "settings.agentConfig.loadFailed");
      throw error;
    }
  }, [currentOrg, msgs, t]);

  const handleSetCredentialPrimary = useCallback(async (id: number) => {
    try {
      await setPrimaryEnvBundle(BigInt(id));
      setCredentialBundles((bundles) =>
        bundles.map((bundle) => ({ ...bundle, is_default: bundle.id === id })),
      );
      msgs.reportSuccess(t("settings.agentConfig.credentialBundles.defaultSet"));
    } catch (error) {
      msgs.reportError(error, t, "settings.agentConfig.credentialBundles.failedToSetDefault");
    }
  }, [msgs, t]);

  const handleClearCredentialPrimary = useCallback(async () => {
    const current = credentialBundles.find((bundle) => bundle.is_default);
    if (!current) return;
    try {
      await updateEnvBundle(BigInt(current.id), { kindPrimary: false });
      setCredentialBundles((bundles) =>
        bundles.map((bundle) => ({ ...bundle, is_default: false })),
      );
      msgs.reportSuccess(t("settings.agentConfig.credentialBundles.defaultSet"));
    } catch (error) {
      msgs.reportError(error, t, "settings.agentConfig.credentialBundles.failedToSetDefault");
    }
  }, [credentialBundles, msgs, t]);

  const handleDeleteCredentialBundle = useCallback(async (id: number) => {
    try {
      await deleteEnvBundle(BigInt(id));
      setCredentialBundles((bundles) => bundles.filter((bundle) => bundle.id !== id));
      msgs.reportSuccess(t("settings.agentConfig.credentialBundles.deleted"));
    } catch (error) {
      msgs.reportError(error, t, "settings.agentConfig.credentialBundles.failedToDelete");
    }
  }, [msgs, t]);

  const handleSaveCredentialBundle = useCallback(async (
    data: CredentialBundleFormData,
    editing: CredentialProfileViewModel | null,
    agent: AgentData,
  ) => {
    if (editing) {
      await updateEnvBundle(BigInt(editing.id), {
        name: data.name,
        description: data.description || undefined,
        hasData: Object.keys(data.data).length > 0,
        data: Object.keys(data.data).length > 0 ? data.data : undefined,
      });
      msgs.reportSuccess(t("settings.agentConfig.credentialBundles.updated"));
    } else {
      await createEnvBundle({
        agentSlug: agent.slug,
        name: data.name,
        description: data.description || undefined,
        kind: "credential",
        data: data.data,
      });
      msgs.reportSuccess(t("settings.agentConfig.credentialBundles.created"));
    }
    await loadCredentialBundles(agent);
  }, [loadCredentialBundles, msgs, t]);

  return {
    credentialFields,
    credentialBundles,
    loadCredentialBundles,
    handleSetCredentialPrimary,
    handleClearCredentialPrimary,
    handleDeleteCredentialBundle,
    handleSaveCredentialBundle,
  };
}
