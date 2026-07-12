import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  GetMobileAccessDescriptorRequestSchema,
  GetPodConnectionRequestSchema,
  MobileAccessDescriptorSchema,
  PodConnectionInfoSchema,
} from "@do-worker/proto/pod/v1/pod_pb";
import { readOrgSlug } from "./auth-store";
import { getMobilePodService } from "./mobile-wasm";

export type MobileWorkerDescriptor = {
  canonicalUrl: string;
  podKey: string;
  status: string;
  interactionMode: string;
  consoleAvailable: boolean;
  previewAvailable: boolean;
  relayAvailable: boolean;
  previewPath?: string;
};

export type MobilePodConnection = {
  relayUrl: string;
  token: string;
  podKey: string;
};

function orgSlug(): string {
  const slug = readOrgSlug();
  if (!slug) throw new Error("当前登录未选择组织");
  return slug;
}

export async function getMobileWorkerDescriptor(
  podKey: string,
): Promise<MobileWorkerDescriptor> {
  const request = create(GetMobileAccessDescriptorRequestSchema, {
    orgSlug: orgSlug(),
    podKey,
  });
  const responseBytes = await (await getMobilePodService())
    .get_mobile_access_descriptor_connect(toBinary(GetMobileAccessDescriptorRequestSchema, request));
  const response = fromBinary(
    MobileAccessDescriptorSchema,
    new Uint8Array(responseBytes),
  );
  return {
    canonicalUrl: response.canonicalUrl,
    podKey: response.podKey,
    status: response.status,
    interactionMode: response.interactionMode,
    consoleAvailable: response.consoleAvailable,
    previewAvailable: response.previewAvailable,
    relayAvailable: response.relayAvailable,
    previewPath: response.previewPath,
  };
}

export async function getMobilePodConnection(
  podKey: string,
): Promise<MobilePodConnection> {
  const request = create(GetPodConnectionRequestSchema, {
    orgSlug: orgSlug(),
    podKey,
  });
  const responseBytes = await (await getMobilePodService())
    .get_pod_connection_connect(toBinary(GetPodConnectionRequestSchema, request));
  const response = fromBinary(PodConnectionInfoSchema, new Uint8Array(responseBytes));
  if (!response.relayUrl || !response.token || !response.podKey) {
    throw new Error("Worker Relay 连接信息无效");
  }
  return {
    relayUrl: response.relayUrl,
    token: response.token,
    podKey: response.podKey,
  };
}
