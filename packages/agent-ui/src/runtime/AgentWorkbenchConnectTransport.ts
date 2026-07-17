import {
  Code,
  ConnectError,
  createClient,
  type Client,
  type Interceptor,
} from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import type {
  CommandEnvelope,
  CommandReceipt,
} from "@do-worker/proto/agent_workbench/v2/command_pb";
import {
  AgentWorkbenchService,
} from "@do-worker/proto/agent_workbench/v2/service_pb";
import type {
  SessionCursor,
  SessionDeltaBatch,
  SessionSnapshot,
} from "@do-worker/proto/agent_workbench/v2/session_pb";

export interface AgentWorkbenchSessionTransport {
  execute(
    command: CommandEnvelope,
    signal?: AbortSignal,
  ): Promise<CommandReceipt>;
  getSnapshot(signal?: AbortSignal): Promise<SessionSnapshot>;
  streamDeltas(
    cursor: SessionCursor,
    signal: AbortSignal,
  ): AsyncIterable<SessionDeltaBatch>;
}

export interface AgentWorkbenchConnectTransportOptions {
  baseUrl: string;
  fetch?: typeof globalThis.fetch;
  getAccessToken: () => Promise<string> | string;
  orgSlug: string;
  replayLimit?: number;
  sessionId: string;
}

export class AgentWorkbenchConnectTransport
  implements AgentWorkbenchSessionTransport
{
  private readonly client: Client<typeof AgentWorkbenchService>;
  private readonly orgSlug: string;
  private readonly replayLimit: number;
  private readonly sessionId: string;

  constructor(options: AgentWorkbenchConnectTransportOptions) {
    this.orgSlug = required(options.orgSlug, "agent_workbench_org_missing");
    this.sessionId = required(
      options.sessionId,
      "agent_workbench_session_missing",
    );
    this.replayLimit = options.replayLimit ?? 256;
    if (this.replayLimit < 1) {
      throw new Error("agent_workbench_replay_limit_invalid");
    }
    this.client = createClient(
      AgentWorkbenchService,
      createConnectTransport({
        baseUrl: required(options.baseUrl, "agent_workbench_base_url_missing"),
        useBinaryFormat: true,
        binaryOptions: { writeUnknownFields: false },
        fetch: options.fetch,
        interceptors: [bearerTokenInterceptor(options.getAccessToken)],
      }),
    );
  }

  getSnapshot(signal?: AbortSignal): Promise<SessionSnapshot> {
    return this.client.getSessionSnapshot(
      { orgSlug: this.orgSlug, sessionId: this.sessionId },
      { signal },
    );
  }

  streamDeltas(
    cursor: SessionCursor,
    signal: AbortSignal,
  ): AsyncIterable<SessionDeltaBatch> {
    if (cursor.sessionId !== this.sessionId) {
      throw new Error("agent_workbench_cursor_session_mismatch");
    }
    return this.client.streamSessionDeltas(
      {
        orgSlug: this.orgSlug,
        cursor,
        replayLimit: this.replayLimit,
      },
      { signal },
    );
  }

  execute(
    command: CommandEnvelope,
    signal?: AbortSignal,
  ): Promise<CommandReceipt> {
    if (command.sessionId !== this.sessionId) {
      throw new Error("agent_workbench_command_session_mismatch");
    }
    return this.client.executeCommand(
      { orgSlug: this.orgSlug, command },
      { signal },
    );
  }
}

export function isAgentWorkbenchCursorRejected(error: unknown): boolean {
  return (
    error instanceof ConnectError &&
    (error.code === Code.FailedPrecondition || error.code === Code.OutOfRange)
  );
}

function bearerTokenInterceptor(
  getAccessToken: AgentWorkbenchConnectTransportOptions["getAccessToken"],
): Interceptor {
  return (next) => async (request) => {
    const token = (await getAccessToken()).trim();
    if (!token) throw new Error("agent_workbench_access_token_missing");
    request.header.set("Authorization", `Bearer ${token}`);
    return next(request);
  };
}

function required(value: string, error: string): string {
  const normalized = value.trim();
  if (!normalized) throw new Error(error);
  return normalized.replace(/\/+$/, "");
}
