import type { AgentSessionRuntime } from "@agent-cloud/agent-ui";

import type { WebAgentWorkbenchRuntimeDeps } from "./webAgentWorkbenchRuntimeTypes";

export function createWebAgentWorkbenchAttachmentUploader(
  deps: WebAgentWorkbenchRuntimeDeps,
  assertSession: (sessionId: string) => void,
): AgentSessionRuntime["uploadAttachment"] {
  const upload = deps.uploadAttachment;
  if (!upload) return undefined;
  return (sessionId, file) => {
    assertSession(sessionId);
    return upload({ access: deps.getAccess(), file, sessionId });
  };
}
