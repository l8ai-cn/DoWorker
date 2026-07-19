import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  GetMobileAccessDescriptorRequestSchema,
  GetPodConnectionRequestSchema,
  ListPodsByTicketRequestSchema,
  ListPodsByTicketResponseSchema,
  MobileAccessDescriptorSchema,
  PodConnectionInfoSchema,
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
import type { PodData } from "@/lib/api/facade/pod";
import { podToCache } from "@/lib/api/projections";
import { getPodService } from "@/lib/wasm-core";

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

export async function terminatePod(orgSlug: string, podKey: string): Promise<string> {
  const request = create(TerminatePodRequestSchema, { orgSlug, podKey });
  const response = await getPodService().terminate_pod_connect(
    toBinary(TerminatePodRequestSchema, request),
  );
  return fromBinary(TerminatePodResponseSchema, new Uint8Array(response)).message;
}

export async function updatePodAlias(
  orgSlug: string,
  podKey: string,
  alias: string | null,
): Promise<string> {
  const request = create(UpdatePodAliasRequestSchema, { orgSlug, podKey, alias: alias ?? "" });
  const response = await getPodService().update_pod_alias_connect(
    toBinary(UpdatePodAliasRequestSchema, request),
  );
  return fromBinary(UpdatePodAliasResponseSchema, new Uint8Array(response)).message;
}

export async function updatePodPerpetual(
  orgSlug: string,
  podKey: string,
  perpetual: boolean,
): Promise<string> {
  const request = create(UpdatePodPerpetualRequestSchema, { orgSlug, podKey, perpetual });
  const response = await getPodService().update_pod_perpetual_connect(
    toBinary(UpdatePodPerpetualRequestSchema, request),
  );
  return fromBinary(UpdatePodPerpetualResponseSchema, new Uint8Array(response)).message;
}

export async function updatePodPreviewConfig(
  orgSlug: string,
  podKey: string,
  previewPort: number,
  previewPath: string,
): Promise<PodData> {
  const request = create(UpdatePodPreviewConfigRequestSchema, {
    orgSlug,
    podKey,
    previewPort,
    previewPath,
  });
  const response = await getPodService().update_pod_preview_config_connect(
    toBinary(UpdatePodPreviewConfigRequestSchema, request),
  );
  return podToCache(
    fromBinary(UpdatePodPreviewConfigResponseSchema, new Uint8Array(response)).pod!,
  );
}

export async function getMobileAccessDescriptor(
  orgSlug: string,
  podKey: string,
): Promise<MobileAccessDescriptor> {
  const request = create(GetMobileAccessDescriptorRequestSchema, { orgSlug, podKey });
  const response = await getPodService().get_mobile_access_descriptor_connect(
    toBinary(GetMobileAccessDescriptorRequestSchema, request),
  );
  const descriptor = fromBinary(MobileAccessDescriptorSchema, new Uint8Array(response));
  return {
    canonical_url: descriptor.canonicalUrl,
    pod_key: descriptor.podKey,
    status: descriptor.status,
    interaction_mode: descriptor.interactionMode,
    console_available: descriptor.consoleAvailable,
    preview_available: descriptor.previewAvailable,
    relay_available: descriptor.relayAvailable,
    preview_path: descriptor.previewPath,
  };
}

export async function getPodConnection(
  orgSlug: string,
  podKey: string,
): Promise<PodConnectionInfo> {
  const request = create(GetPodConnectionRequestSchema, { orgSlug, podKey });
  const response = await getPodService().get_pod_connection_connect(
    toBinary(GetPodConnectionRequestSchema, request),
  );
  const connection = fromBinary(PodConnectionInfoSchema, new Uint8Array(response));
  return { relay_url: connection.relayUrl, token: connection.token, pod_key: connection.podKey };
}

export async function sendPodPrompt(
  orgSlug: string,
  podKey: string,
  prompt: string,
): Promise<string> {
  const request = create(SendPodPromptRequestSchema, { orgSlug, podKey, prompt });
  const response = await getPodService().send_pod_prompt_connect(
    toBinary(SendPodPromptRequestSchema, request),
  );
  return fromBinary(SendPodPromptResponseSchema, new Uint8Array(response)).status;
}

export async function listPodsByTicket(
  orgSlug: string,
  ticketId: number,
): Promise<{ items: PodData[]; total: number }> {
  const request = create(ListPodsByTicketRequestSchema, { orgSlug, ticketId: BigInt(ticketId) });
  const response = await getPodService().list_pods_by_ticket_connect(
    toBinary(ListPodsByTicketRequestSchema, request),
  );
  const result = fromBinary(ListPodsByTicketResponseSchema, new Uint8Array(response));
  return { items: result.items.map(podToCache), total: Number(result.total) };
}
