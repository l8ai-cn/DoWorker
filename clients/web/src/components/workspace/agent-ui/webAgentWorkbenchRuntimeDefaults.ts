import {
  getAgentWorkbenchService,
  getAgentWorkbenchState,
  getAuthManager,
} from "@/lib/wasm-core";
import { getAgentWorkbenchStreamBaseUrl } from "@/lib/env";
import { readCurrentOrg } from "@/stores/auth";
import type {
  WebAgentWorkbenchRuntimeDeps,
} from "./webAgentWorkbenchRuntimeTypes";

export const defaultWebAgentWorkbenchRuntimeDeps:
  WebAgentWorkbenchRuntimeDeps = {
    getAccess: () => {
      const bearerToken = getAuthManager().get_token();
      const orgSlug = readCurrentOrg()?.slug;
      if (!bearerToken || !orgSlug) {
        throw new Error("agent_workbench_access_missing");
      }
      return { bearerToken, orgSlug };
    },
    service: {
      executeCommandConnect: (...args) =>
        getAgentWorkbenchService().executeCommandConnect(...args),
      getSessionSnapshotConnect: (...args) =>
        getAgentWorkbenchService().getSessionSnapshotConnect(...args),
      streamSessionDeltasConnect: (
        orgSlug,
        bearerToken,
        sessionId,
        replayLimit,
        onCommit,
        onError,
        onClose,
      ) =>
        getAgentWorkbenchService().streamSessionDeltasConnect(
          orgSlug,
          bearerToken,
          getAgentWorkbenchStreamBaseUrl(),
          sessionId,
          replayLimit,
          onCommit,
          onError,
          onClose,
        ),
    },
    sleep: (milliseconds) =>
      new Promise((resolve) => setTimeout(resolve, milliseconds)),
    state: {
      projectionStatus: (sessionId) =>
        getAgentWorkbenchState().projectionStatus(sessionId),
      resyncReason: (sessionId) =>
        getAgentWorkbenchState().resyncReason(sessionId),
      revision: (sessionId) =>
        getAgentWorkbenchState().revision(sessionId),
      snapshotBytes: (sessionId) =>
        getAgentWorkbenchState().snapshotBytes(sessionId),
    },
  };
