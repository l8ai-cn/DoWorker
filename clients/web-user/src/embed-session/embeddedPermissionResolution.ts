import type {
  AgentPermissionResolution,
} from "@do-worker/agent-ui";

import type { EmbedSessionClient } from "@/embed-session-api";
import {
  resolvePermissionBlock,
  type EmbeddedRuntimeState,
} from "./embeddedRuntimeState";

export async function resolveEmbeddedPermission(
  client: EmbedSessionClient,
  state: EmbeddedRuntimeState,
  permissionId: string,
  result: AgentPermissionResolution,
): Promise<EmbeddedRuntimeState> {
  if (!client.resolvePermission) throw new Error("Approval is not permitted");
  await client.resolvePermission(permissionId, result);
  return {
    ...state,
    blocks: resolvePermissionBlock(state.blocks, permissionId, result),
  };
}
