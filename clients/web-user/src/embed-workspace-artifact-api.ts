import {
  workspaceFileArtifacts,
  type AgentArtifactItem,
} from "@do-worker/agent-ui";
import { readEmbeddedJson } from "./embed-session-response-parsers";

type EmbeddedRequest = (path: string, init?: RequestInit) => Promise<Response>;

export async function loadEmbeddedWorkspaceArtifact(
  request: EmbeddedRequest,
  sessionPath: string,
  path: string,
): Promise<Blob> {
  if (!path) throw new Error("Workspace artifact path is empty");
  const response = await request(
    `${sessionPath}/resources/environments/workspace/filesystem/${encodePath(path)}`,
  );
  const body = await workspaceFileResponse(response);
  const mimeType =
    typeof body.content_type === "string" ? body.content_type : "";
  if (body.encoding === "base64") {
    const decoded = atob(body.content);
    const bytes = Uint8Array.from(decoded, (char) => char.charCodeAt(0));
    return new Blob([bytes], { type: mimeType });
  }
  if (body.encoding !== undefined && body.encoding !== "utf-8") {
    throw new Error("Workspace artifact encoding is unsupported");
  }
  return new Blob([body.content], { type: mimeType });
}

export async function listEmbeddedWorkspaceArtifacts(
  request: EmbeddedRequest,
  sessionPath: string,
): Promise<AgentArtifactItem[]> {
  const response = await request(
    `${sessionPath}/resources/environments/workspace/changes`,
  );
  const body = (await readEmbeddedJson(response)) as { data?: unknown };
  return workspaceFileArtifacts("workspace-discovery", body.data);
}

async function workspaceFileResponse(response: Response): Promise<{
  content: string;
  content_type?: unknown;
  encoding?: unknown;
}> {
  const body = (await readEmbeddedJson(response)) as {
    content?: unknown;
    content_type?: unknown;
    encoding?: unknown;
    truncated?: unknown;
  };
  if (body.truncated === true) {
    throw new Error("Workspace artifact exceeds the preview size limit");
  }
  if (typeof body.content !== "string") {
    throw new Error("Workspace artifact response is invalid");
  }
  return { ...body, content: body.content };
}

function encodePath(path: string): string {
  return path.split("/").map(encodeURIComponent).join("/");
}
