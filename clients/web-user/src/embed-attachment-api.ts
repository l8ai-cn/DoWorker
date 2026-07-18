import type { AgentAttachmentReference } from "@do-worker/agent-ui";

type EmbeddedRequest = (path: string, init?: RequestInit) => Promise<Response>;

export async function uploadEmbeddedAttachment(
  request: EmbeddedRequest,
  sessionPath: string,
  file: File,
): Promise<AgentAttachmentReference> {
  const body = new FormData();
  body.append("file", file);
  const response = await request(`${sessionPath}/resources/files`, {
    body,
    method: "POST",
  });
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
    name: result.name || file.name,
    mediaType: file.type || "application/octet-stream",
    bytes: result.metadata?.bytes ?? file.size,
  };
}
