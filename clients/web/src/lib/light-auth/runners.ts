// Runner CRUD over Connect-RPC JSON. GetRunnerAuthStatus is public (no JWT
// — the auth key IS the auth); AuthorizeRunner / CreateRunnerToken /
// ListRunners require the user's bearer token.

import { lightConnect } from "./api-fetch";
import type { RunnerAuthStatus } from "@/lib/viewModels/runner";

interface ConnectRunnerAuthStatus {
  status: string;
  nodeId?: string;
  expiresAt?: string;
}

export async function lightGetRunnerAuthStatus(
  authKey: string,
): Promise<RunnerAuthStatus> {
  const resp = await lightConnect<{ authKey: string }, ConnectRunnerAuthStatus>(
    "proto.runner_api.v1.RunnerPublicService",
    "GetRunnerAuthStatus",
    { authKey },
  );
  return {
    status: resp.status as RunnerAuthStatus["status"],
    node_id: resp.nodeId,
    expires_at: resp.expiresAt,
  };
}

export interface LightAuthorizeRunnerInput {
  organizationSlug: string;
  authKey: string;
  clusterId: string;
  nodeId?: string;
}

interface ConnectAuthorizeRunnerResponse {
  runnerId?: number | string;
  nodeId?: string;
  message?: string;
}

export async function lightAuthorizeRunner(
  input: LightAuthorizeRunnerInput,
): Promise<{ runner_id?: number; node_id?: string; message?: string }> {
  if (!/^[1-9]\d*$/.test(input.clusterId)) {
    throw new Error("invalid execution cluster id");
  }
  const resp = await lightConnect<
    { orgSlug: string; authKey: string; nodeId: string; clusterId: string },
    ConnectAuthorizeRunnerResponse
  >(
    "proto.runner_api.v1.RunnerService",
    "AuthorizeRunner",
    {
      orgSlug: input.organizationSlug,
      authKey: input.authKey,
      nodeId: input.nodeId ?? "",
      clusterId: input.clusterId,
    },
    { authenticated: true },
  );
  return {
    runner_id: resp.runnerId !== undefined ? Number(resp.runnerId) : undefined,
    node_id: resp.nodeId,
    message: resp.message,
  };
}
