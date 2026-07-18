import type { AgentAttachmentReference } from "@do-worker/agent-ui";

import { getApiBaseUrl } from "@/lib/env";
import type { WebAgentWorkbenchAttachmentUploadInput } from "./webAgentWorkbenchRuntimeTypes";

export async function uploadWebAgentWorkbenchAttachment(
  input: WebAgentWorkbenchAttachmentUploadInput,
  fetcher: typeof globalThis.fetch = globalThis.fetch,
): Promise<AgentAttachmentReference> {
  const mediaType = input.file.type.trim();
  if (!mediaType) throw new Error("agent_attachment_upload_unsupported");

  const body = new FormData();
  body.append("file", input.file, input.file.name);
  const base = getApiBaseUrl().replace(/\/$/, "");
  const response = await fetcher(
    `${base}/v1/sessions/${encodeURIComponent(input.sessionId)}/resources/files`,
    {
      body,
      cache: "no-store",
      headers: {
        Authorization: `Bearer ${input.access.bearerToken}`,
        "X-Organization-Slug": input.access.orgSlug,
      },
      method: "POST",
    },
  );
  const payload = await response.json().catch(() => null);
  if (!response.ok) throw uploadError(response.status, payload);
  return attachmentReference(payload, mediaType);
}

function attachmentReference(
  payload: unknown,
  mediaType: string,
): AgentAttachmentReference {
  const wire = object(payload);
  const metadata = object(wire?.metadata);
  const id = string(wire?.id);
  const name = string(wire?.name);
  const bytes = metadata?.bytes;
  if (!id || !name || typeof bytes !== "number" || bytes < 0) {
    throw new Error("agent_attachment_upload_invalid_response");
  }
  return { bytes, id, mediaType, name };
}

function uploadError(status: number, payload: unknown) {
  const message = string(object(payload)?.error);
  if (status === 400 && message?.startsWith("file type not allowed")) {
    return new Error("agent_attachment_upload_unsupported");
  }
  return new Error("agent_attachment_upload_failed");
}

function object(value: unknown): Record<string, unknown> | undefined {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : undefined;
}

function string(value: unknown): string | undefined {
  return typeof value === "string" && value.trim() ? value : undefined;
}
