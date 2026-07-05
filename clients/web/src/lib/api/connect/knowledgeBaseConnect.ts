// Connect-RPC adapter for proto.knowledgebase.v1.KnowledgeBaseService.
//
// Encodes requests via @bufbuild/protobuf .toBinary(), passes the Uint8Array
// to the wasm bridge (binary in / binary out — conventions §2.5), decodes
// responses via .fromBinary(). Returns snake_case web shapes.

import {
  CreateKnowledgeBaseRequestSchema,
  DeleteKnowledgeBaseRequestSchema,
  GetKnowledgeBaseFileRequestSchema,
  GetKnowledgeBaseRequestSchema,
  KnowledgeBaseFileSchema,
  KnowledgeBaseSchema,
  ListAgentMountsRequestSchema,
  ListAgentMountsResponseSchema,
  ListKnowledgeBaseDirRequestSchema,
  ListKnowledgeBaseDirResponseSchema,
  ListKnowledgeBasesRequestSchema,
  ListKnowledgeBasesResponseSchema,
  SetAgentMountsRequestSchema,
  UpdateKnowledgeBaseRequestSchema,
  type KnowledgeBase as ProtoKnowledgeBase,
} from "@proto/knowledgebase/v1/knowledgebase_pb";
import { create, toBinary, fromBinary } from "@bufbuild/protobuf";
import { getKnowledgeBaseService } from "@/lib/wasm-core";
import type {
  KbAgentMount,
  KbDirEntry,
  KbFile,
  KnowledgeBase,
} from "../facade/knowledgeBaseApi";

export function fromProtoKnowledgeBase(p: ProtoKnowledgeBase): KnowledgeBase {
  return {
    id: Number(p.id),
    slug: p.slug,
    name: p.name,
    description: p.description,
    http_clone_url: p.httpCloneUrl,
    default_branch: p.defaultBranch,
    source_type: p.sourceType,
    sync_status: p.syncStatus,
    sync_error: p.syncError,
    last_synced_at: p.lastSyncedAt,
    created_at: p.createdAt,
    updated_at: p.updatedAt,
  };
}

export async function listKnowledgeBases(orgSlug: string): Promise<KnowledgeBase[]> {
  const req = create(ListKnowledgeBasesRequestSchema, { orgSlug });
  const bytes = toBinary(ListKnowledgeBasesRequestSchema, req);
  const respBytes = await getKnowledgeBaseService().listKnowledgeBasesConnect(bytes);
  const resp = fromBinary(ListKnowledgeBasesResponseSchema, new Uint8Array(respBytes));
  return resp.items.map(fromProtoKnowledgeBase);
}

export async function getKnowledgeBase(orgSlug: string, slug: string): Promise<KnowledgeBase> {
  const req = create(GetKnowledgeBaseRequestSchema, { orgSlug, slug });
  const bytes = toBinary(GetKnowledgeBaseRequestSchema, req);
  const respBytes = await getKnowledgeBaseService().getKnowledgeBaseConnect(bytes);
  return fromProtoKnowledgeBase(fromBinary(KnowledgeBaseSchema, new Uint8Array(respBytes)));
}

export async function createKnowledgeBase(
  orgSlug: string,
  input: { name: string; description?: string; sourceType?: string; sourceConfigJson?: string },
): Promise<KnowledgeBase> {
  const req = create(CreateKnowledgeBaseRequestSchema, {
    orgSlug,
    name: input.name,
    description: input.description,
    sourceType: input.sourceType,
    sourceConfigJson: input.sourceConfigJson,
  });
  const bytes = toBinary(CreateKnowledgeBaseRequestSchema, req);
  const respBytes = await getKnowledgeBaseService().createKnowledgeBaseConnect(bytes);
  return fromProtoKnowledgeBase(fromBinary(KnowledgeBaseSchema, new Uint8Array(respBytes)));
}

export async function updateKnowledgeBase(
  orgSlug: string,
  slug: string,
  input: { name?: string; description?: string },
): Promise<KnowledgeBase> {
  const req = create(UpdateKnowledgeBaseRequestSchema, {
    orgSlug,
    slug,
    name: input.name,
    description: input.description,
  });
  const bytes = toBinary(UpdateKnowledgeBaseRequestSchema, req);
  const respBytes = await getKnowledgeBaseService().updateKnowledgeBaseConnect(bytes);
  return fromProtoKnowledgeBase(fromBinary(KnowledgeBaseSchema, new Uint8Array(respBytes)));
}

export async function deleteKnowledgeBase(orgSlug: string, slug: string): Promise<void> {
  const req = create(DeleteKnowledgeBaseRequestSchema, { orgSlug, slug });
  const bytes = toBinary(DeleteKnowledgeBaseRequestSchema, req);
  await getKnowledgeBaseService().deleteKnowledgeBaseConnect(bytes);
}

export async function listKbAgentMounts(orgSlug: string, slug: string): Promise<KbAgentMount[]> {
  const req = create(ListAgentMountsRequestSchema, { orgSlug, slug });
  const bytes = toBinary(ListAgentMountsRequestSchema, req);
  const respBytes = await getKnowledgeBaseService().listAgentMountsConnect(bytes);
  const resp = fromBinary(ListAgentMountsResponseSchema, new Uint8Array(respBytes));
  return resp.mounts.map((m) => ({ agent_slug: m.agentSlug, mode: m.mode === "rw" ? "rw" : "ro" }));
}

export async function setKbAgentMounts(
  orgSlug: string,
  slug: string,
  mounts: KbAgentMount[],
): Promise<void> {
  const req = create(SetAgentMountsRequestSchema, {
    orgSlug,
    slug,
    mounts: mounts.map((m) => ({ agentSlug: m.agent_slug, mode: m.mode })),
  });
  const bytes = toBinary(SetAgentMountsRequestSchema, req);
  await getKnowledgeBaseService().setAgentMountsConnect(bytes);
}

export async function getKbFile(orgSlug: string, slug: string, path: string): Promise<KbFile> {
  const req = create(GetKnowledgeBaseFileRequestSchema, { orgSlug, slug, path });
  const bytes = toBinary(GetKnowledgeBaseFileRequestSchema, req);
  const respBytes = await getKnowledgeBaseService().getKnowledgeBaseFileConnect(bytes);
  const resp = fromBinary(KnowledgeBaseFileSchema, new Uint8Array(respBytes));
  return { path: resp.path, content: resp.content, size: Number(resp.size) };
}

export async function listKbDir(
  orgSlug: string,
  slug: string,
  path: string,
): Promise<KbDirEntry[]> {
  const req = create(ListKnowledgeBaseDirRequestSchema, { orgSlug, slug, path });
  const bytes = toBinary(ListKnowledgeBaseDirRequestSchema, req);
  const respBytes = await getKnowledgeBaseService().listKnowledgeBaseDirConnect(bytes);
  const resp = fromBinary(ListKnowledgeBaseDirResponseSchema, new Uint8Array(respBytes));
  return resp.entries.map((e) => ({
    name: e.name,
    path: e.path,
    type: e.type === "dir" ? "dir" : "file",
    size: Number(e.size),
  }));
}
