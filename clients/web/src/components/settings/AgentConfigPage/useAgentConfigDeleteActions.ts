import { useCallback } from "react";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";

interface AgentConfigDeleteActions {
  dialogProps: ReturnType<typeof useConfirmDialog>["dialogProps"];
  deleteRuntime: (id: number) => Promise<void>;
  deleteCredential: (id: number) => Promise<void>;
  deleteConfigFile: (id: number) => Promise<void>;
}

interface AgentConfigDeleteDeps {
  deleteRuntime: (id: number) => Promise<void>;
  deleteCredential: (id: number) => Promise<void>;
  deleteConfigFile: (id: number) => Promise<void>;
  t: (key: string) => string;
}

export function useAgentConfigDeleteActions({
  deleteRuntime,
  deleteCredential,
  deleteConfigFile,
  t,
}: AgentConfigDeleteDeps): AgentConfigDeleteActions {
  const { dialogProps, confirm } = useConfirmDialog();
  const remove = useCallback(async (description: string, action: (id: number) => Promise<void>, id: number) => {
    const confirmed = await confirm({
      title: t("common.confirmDelete"),
      description,
      variant: "destructive",
      confirmText: t("common.delete"),
      cancelText: t("common.cancel"),
    });
    if (confirmed) await action(id);
  }, [confirm, t]);
  return {
    dialogProps,
    deleteRuntime: (id) => remove(t("settings.agentConfig.runtimeBundles.confirmDelete"), deleteRuntime, id),
    deleteCredential: (id) => remove(t("settings.agentConfig.credentialBundles.confirmDelete"), deleteCredential, id),
    deleteConfigFile: (id) => remove(t("settings.agentConfig.configFiles.confirmDelete"), deleteConfigFile, id),
  };
}
