// Connect-RPC adapter for proto.pod.v1.PodService.
//
// Encodes requests via @bufbuild/protobuf .toBinary(), passes the Uint8Array
// to the wasm bridge (binary in / binary out per conventions §2.5), decodes
// responses via .fromBinary(). No JSON intermediate.
//
// Returns snake_case web shapes (PodData, PodConnectionInfo) so call sites
// don't have to switch wire-camelCase off the proto generated types — same
// pattern as runnerConnect.ts during the dual-track migration window.

import {
  CreatePodRequestSchema,
  CreatePodResponseSchema,
  GetMobileAccessDescriptorRequestSchema,
  GetPodConnectionRequestSchema,
  ListPodsByTicketRequestSchema,
  ListPodsByTicketResponseSchema,
  PodConnectionInfoSchema,
  MobileAccessDescriptorSchema,
  SendPodPromptRequestSchema,
  SendPodPromptResponseSchema,
  TerminatePodRequestSchema,
  TerminatePodResponseSchema,
  UpdatePodAliasRequestSchema,
  UpdatePodAliasResponseSchema,
  UpdatePodPerpetualRequestSchema,
  UpdatePodPerpetualResponseSchema,
  UpdatePodPreviewConfigRequestSchema,
  UpdatePodPreviewConfigResponseSchema,
} from "@proto/pod/v1/pod_pb";
import { create, toBinary, fromBinary } from "@bufbuild/protobuf";
// Shared proto->PodData projection. Aliased to the historical fromProtoPod
// name and re-exported for cross-file use (stores/pod, podProtoMap, facade).
import { podToCache as fromProtoPod } from "@/lib/api/projections";
import { getPodService } from "@/lib/wasm-core";
import type { PodData } from "@/lib/api/facade/pod";
import { workerDraftToProto } from "./podWorkerDraftProto";
import type { WorkerSpecDraft } from "./podWorkerCreationTypes";

export { fromProtoPod };

export interface PodConnectionInfo {
  relay_url: string;
  token: string;
  pod_key: string;
}

export interface MobileAccessDescriptor {
  canonical_url: string;
  pod_key: string;
  status: string;
  interaction_mode: string;
  console_available: boolean;
  preview_available: boolean;
  relay_available: boolean;
  preview_path?: string;
}

export interface CreatePodInput {
  agent_slug: string;
  runner_id?: number;
  ticket_slug?: string;
  alias?: string;
  agentfile_layer?: string;
  automation_level?: string;
  repository_id?: number;
  cols?: number;
  rows?: number;
  source_pod_key?: string;
  resume_agent_session?: boolean;
  perpetual?: boolean;
  model_resource_id?: number;
  token_budget?: number;
  worker_spec?: WorkerSpecDraft;
}

export async function createPod(
  orgSlug: string,
  input: CreatePodInput,
): Promise<{ pod: PodData; warning?: string }> {
  const req = create(CreatePodRequestSchema, {
    orgSlug,
    agentSlug: input.agent_slug,
    runnerId: input.runner_id === undefined ? undefined : BigInt(input.runner_id),
    ticketSlug: input.ticket_slug,
    alias: input.alias,
    agentfileLayer: input.agentfile_layer,
    automationLevel: input.automation_level,
    repositoryId: input.repository_id === undefined ? undefined : BigInt(input.repository_id),
    cols: input.cols ?? 0,
    rows: input.rows ?? 0,
    sourcePodKey: input.source_pod_key,
    resumeAgentSession: input.resume_agent_session,
    perpetual: input.perpetual,
    modelResourceId:
      input.model_resource_id === undefined ? undefined : BigInt(input.model_resource_id),
    tokenBudget: input.token_budget === undefined ? undefined : BigInt(input.token_budget),
    workerSpec: input.worker_spec ? workerDraftToProto(input.worker_spec) : undefined,
  });
  const bytes = toBinary(CreatePodRequestSchema, req);
  const respBytes = await getPodService().create_pod_connect(bytes);
  const resp = fromBinary(CreatePodResponseSchema, new Uint8Array(respBytes));
  return { pod: fromProtoPod(resp.pod!), warning: resp.warning };
}

export async function wakePod(
  orgSlug: string,
  sourcePodKey: string,
): Promise<{ pod: PodData; warning?: string }> {
  return createPod(orgSlug, {
    agent_slug: "",
    source_pod_key: sourcePodKey,
    resume_agent_session: true,
  });
}

export async function terminatePod(orgSlug: string, podKey: string): Promise<string> {
  const req = create(TerminatePodRequestSchema, { orgSlug, podKey });
  const bytes = toBinary(TerminatePodRequestSchema, req);
  const respBytes = await getPodService().terminate_pod_connect(bytes);
  return fromBinary(TerminatePodResponseSchema, new Uint8Array(respBytes)).message;
}

export async function updatePodAlias(
  orgSlug: string,
  podKey: string,
  alias: string | null,
): Promise<string> {
  const req = create(UpdatePodAliasRequestSchema, {
    orgSlug,
    podKey,
    // alias is `optional string` — undefined = no change, "" = clear.
    alias: alias === null ? "" : alias,
  });
  const bytes = toBinary(UpdatePodAliasRequestSchema, req);
  const respBytes = await getPodService().update_pod_alias_connect(bytes);
  return fromBinary(UpdatePodAliasResponseSchema, new Uint8Array(respBytes)).message;
}

export async function updatePodPerpetual(
  orgSlug: string,
  podKey: string,
  perpetual: boolean,
): Promise<string> {
  const req = create(UpdatePodPerpetualRequestSchema, { orgSlug, podKey, perpetual });
  const bytes = toBinary(UpdatePodPerpetualRequestSchema, req);
  const respBytes = await getPodService().update_pod_perpetual_connect(bytes);
  return fromBinary(UpdatePodPerpetualResponseSchema, new Uint8Array(respBytes)).message;
}

export async function updatePodPreviewConfig(
  orgSlug: string,
  podKey: string,
  previewPort: number,
  previewPath: string,
): Promise<PodData> {
  const req = create(UpdatePodPreviewConfigRequestSchema, {
    orgSlug,
    podKey,
    previewPort,
    previewPath,
  });
  const bytes = toBinary(UpdatePodPreviewConfigRequestSchema, req);
  const respBytes = await getPodService().update_pod_preview_config_connect(bytes);
  const response = fromBinary(
    UpdatePodPreviewConfigResponseSchema,
    new Uint8Array(respBytes),
  );
  return fromProtoPod(response.pod!);
}

export async function getMobileAccessDescriptor(
  orgSlug: string,
  podKey: string,
): Promise<MobileAccessDescriptor> {
  const req = create(GetMobileAccessDescriptorRequestSchema, { orgSlug, podKey });
  const bytes = toBinary(GetMobileAccessDescriptorRequestSchema, req);
  const respBytes = await getPodService().get_mobile_access_descriptor_connect(bytes);
  const response = fromBinary(MobileAccessDescriptorSchema, new Uint8Array(respBytes));
  return {
    canonical_url: response.canonicalUrl,
    pod_key: response.podKey,
    status: response.status,
    interaction_mode: response.interactionMode,
    console_available: response.consoleAvailable,
    preview_available: response.previewAvailable,
    relay_available: response.relayAvailable,
    preview_path: response.previewPath,
  };
}

export async function getPodConnection(
  orgSlug: string,
  podKey: string,
): Promise<PodConnectionInfo> {
  const req = create(GetPodConnectionRequestSchema, { orgSlug, podKey });
  const bytes = toBinary(GetPodConnectionRequestSchema, req);
  const respBytes = await getPodService().get_pod_connection_connect(bytes);
  const c = fromBinary(PodConnectionInfoSchema, new Uint8Array(respBytes));
  return {
    relay_url: c.relayUrl,
    token: c.token,
    pod_key: c.podKey,
  };
}

export async function sendPodPrompt(
  orgSlug: string,
  podKey: string,
  prompt: string,
): Promise<string> {
  const req = create(SendPodPromptRequestSchema, { orgSlug, podKey, prompt });
  const bytes = toBinary(SendPodPromptRequestSchema, req);
  const respBytes = await getPodService().send_pod_prompt_connect(bytes);
  return fromBinary(SendPodPromptResponseSchema, new Uint8Array(respBytes)).status;
}

export async function listPodsByTicket(
  orgSlug: string,
  ticketId: number,
): Promise<{ items: PodData[]; total: number }> {
  const req = create(ListPodsByTicketRequestSchema, { orgSlug, ticketId: BigInt(ticketId) });
  const bytes = toBinary(ListPodsByTicketRequestSchema, req);
  const respBytes = await getPodService().list_pods_by_ticket_connect(bytes);
  const resp = fromBinary(ListPodsByTicketResponseSchema, new Uint8Array(respBytes));
  return { items: resp.items.map(fromProtoPod), total: Number(resp.total) };
}

export { getPod, listPods, listPodsRaw } from "./podQueryConnect";
