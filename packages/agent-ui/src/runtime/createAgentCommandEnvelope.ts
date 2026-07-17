import {
  create,
  toBinary,
  type MessageInitShape,
} from "@bufbuild/protobuf";

import {
  CommandEnvelopeSchema,
  type CommandEnvelope,
} from "@do-worker/proto/agent_workbench/v2/command_pb";

export interface CreateAgentCommandEnvelopeInput {
  command: MessageInitShape<typeof CommandEnvelopeSchema>["command"];
  commandId: string;
  expectedRevision?: bigint;
  issuedAt: string;
  sessionId: string;
  streamEpoch: string;
}

export async function createAgentCommandEnvelope(
  input: CreateAgentCommandEnvelopeInput,
): Promise<CommandEnvelope> {
  const command = create(CommandEnvelopeSchema, {
    sessionId: input.sessionId,
    streamEpoch: input.streamEpoch,
    commandId: input.commandId,
    payloadDigest: "",
    expectedRevision: input.expectedRevision,
    issuedAt: input.issuedAt,
    command: input.command,
  });
  const bytes = toBinary(CommandEnvelopeSchema, command, {
    writeUnknownFields: false,
  });
  const digest = await globalThis.crypto.subtle.digest("SHA-256", bytes);
  command.payloadDigest = `sha256:${hex(new Uint8Array(digest))}`;
  return command;
}

function hex(bytes: Uint8Array): string {
  return Array.from(bytes, (value) => value.toString(16).padStart(2, "0")).join(
    "",
  );
}
