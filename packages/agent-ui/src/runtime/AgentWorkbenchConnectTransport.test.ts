import { fromBinary, toBinary } from "@bufbuild/protobuf";
import { describe, expect, it, vi } from "vitest";

import {
  GetSessionSnapshotRequestSchema,
} from "@do-worker/proto/agent_workbench/v2/service_pb";
import { SessionSnapshotSchema } from "@do-worker/proto/agent_workbench/v2/session_pb";
import { SessionStatus } from "@do-worker/proto/agent_workbench/v2/session_state_pb";
import { AgentWorkbenchConnectTransport } from "./AgentWorkbenchConnectTransport";

describe("AgentWorkbenchConnectTransport", () => {
  it("uses the exact org/session scope and refreshes authorization per RPC", async () => {
    const tokens = ["token-1", "token-2"];
    const getAccessToken = vi.fn(async () => tokens.shift() ?? "");
    const fetch = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      const request = fromBinary(
        GetSessionSnapshotRequestSchema,
        requestBytes(init?.body),
      );
      expect(request).toMatchObject({
        orgSlug: "demo-org",
        sessionId: "session-1",
      });
      return new Response(
        toBinary(
          SessionSnapshotSchema,
          {
            $typeName: "proto.agent_workbench.v2.SessionSnapshot",
            sessionId: "session-1",
            streamEpoch: "epoch-1",
            revision: 1n,
            latestSequence: 0n,
            history: [],
            commandReceipts: [],
            permissionRequests: [],
            grants: [],
            resources: [],
            artifacts: [],
            status: SessionStatus.IDLE,
          },
        ),
        { headers: { "Content-Type": "application/proto" } },
      );
    });
    const transport = new AgentWorkbenchConnectTransport({
      baseUrl: "http://api.test",
      orgSlug: "demo-org",
      sessionId: "session-1",
      getAccessToken,
      fetch,
    });

    await transport.getSnapshot();
    await transport.getSnapshot();

    expect(getAccessToken).toHaveBeenCalledTimes(2);
    expect(fetch).toHaveBeenCalledTimes(2);
    expect(requestAuthorization(fetch, 0)).toBe("Bearer token-1");
    expect(requestAuthorization(fetch, 1)).toBe("Bearer token-2");
  });

  it("rejects an empty token before issuing the request", async () => {
    const fetch = vi.fn();
    const transport = new AgentWorkbenchConnectTransport({
      baseUrl: "http://api.test",
      orgSlug: "demo-org",
      sessionId: "session-1",
      getAccessToken: async () => "",
      fetch,
    });

    await expect(transport.getSnapshot()).rejects.toThrow(
      "agent_workbench_access_token_missing",
    );
    expect(fetch).not.toHaveBeenCalled();
  });
});

function requestBytes(body: BodyInit | null | undefined): Uint8Array {
  if (body instanceof Uint8Array) return body;
  if (body instanceof ArrayBuffer) return new Uint8Array(body);
  throw new Error("unexpected_request_body");
}

function requestAuthorization(
  fetch: ReturnType<typeof vi.fn>,
  index: number,
): string | null {
  const init = fetch.mock.calls[index]?.[1] as RequestInit | undefined;
  return new Headers(init?.headers).get("Authorization");
}
