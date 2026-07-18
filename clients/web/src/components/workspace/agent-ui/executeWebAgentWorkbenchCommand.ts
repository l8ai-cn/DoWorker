import { toBinary } from "@bufbuild/protobuf";
import {
  createAgentCommandEnvelope,
} from "@do-worker/agent-ui";
import { CommandEnvelopeSchema } from "@proto/agent_workbench/v2/command_pb";
import type { SessionSnapshot } from "@proto/agent_workbench/v2/session_pb";

import type {
  WebAgentWorkbenchRuntimeDeps,
} from "./webAgentWorkbenchRuntimeTypes";

export async function executeWebAgentWorkbenchCommand(input: {
  command: Parameters<typeof createAgentCommandEnvelope>[0]["command"];
  commandId: string;
  deps: WebAgentWorkbenchRuntimeDeps;
  sessionId: string;
  snapshot: SessionSnapshot | null | undefined;
}): Promise<void> {
  if (input.deps.state.projectionStatus(input.sessionId) !== "ready") {
    throw new Error("agent_workbench_session_not_ready");
  }
  if (!input.snapshot) throw new Error("agent_workbench_snapshot_missing");
  const envelope = await createAgentCommandEnvelope({
    command: input.command,
    commandId: input.commandId,
    expectedRevision: input.snapshot.revision,
    issuedAt: new Date().toISOString(),
    sessionId: input.sessionId,
    streamEpoch: input.snapshot.streamEpoch,
  });
  const access = input.deps.getAccess();
  await input.deps.service.executeCommandConnect(
    access.orgSlug,
    access.bearerToken,
    toBinary(CommandEnvelopeSchema, envelope),
  );
}
