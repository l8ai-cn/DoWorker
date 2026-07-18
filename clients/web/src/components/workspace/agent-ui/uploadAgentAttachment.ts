import type { AgentAttachmentReference } from "@do-worker/agent-ui";

import { getApiBaseUrl } from "@/lib/env";

export async function uploadAgentAttachment(input: {
  bearerToken: string;
  file: File;
  orgSlug: string;
  sessionId: string;
}): Promise<AgentAttachmentReference> {
  const body = new FormData();
  body.append("file", input.file);
  const base = getApiBaseUrl().replace(/\/$/, "");
  const response = await fetch(
    `${base}/v1/sessions/${encodeURIComponent(input.sessionId)}/resources/files`,
    {
      body,
      headers: {
        Authorization: `Bearer ${input.bearerToken}`,
        "X-Organization-Slug": input.orgSlug,
      },
      method: "POST",
    },
  );
  if (!response.ok) {
    throw new Error(`agent_attachment_upload_failed:${response.status}`);
  }
  const result = (await response.json()) as {
    id?: string;
    metadata?: { bytes?: number };
    name?: string;
  };
  if (!result.id) throw new Error("agent_attachment_upload_invalid");
  return {
    id: result.id,
    name: result.name || input.file.name,
    mediaType: input.file.type || "application/octet-stream",
    bytes: result.metadata?.bytes ?? input.file.size,
  };
}
