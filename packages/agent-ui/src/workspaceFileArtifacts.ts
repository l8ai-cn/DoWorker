import type { AgentArtifactItem } from "./agentArtifactContracts";

export function workspaceFileArtifacts(
  sourceId: string,
  changes: unknown,
): AgentArtifactItem[] {
  if (!Array.isArray(changes)) return [];
  return changes.flatMap((change, index) => {
    const path = changedPath(change);
    const mimeType = path ? deliverableMimeType(path) : undefined;
    if (
      !path ||
      !mimeType ||
      hasIgnoredPathSegment(path) ||
      changedStatus(change) === "deleted"
    ) {
      return [];
    }
    return [{
      id: `${sourceId}:artifact:${index}`,
      kind: "artifact",
      artifactId: `workspace:${path}`,
      filename: path.split("/").pop() || path,
      mimeType,
      status: "completed",
    }];
  });
}

function changedPath(value: unknown): string | null {
  if (!value || typeof value !== "object") return null;
  const path = (value as { path?: unknown }).path;
  return typeof path === "string" && path.trim() ? path : null;
}

function changedStatus(value: unknown): string | null {
  if (!value || typeof value !== "object") return null;
  const status = (value as { status?: unknown }).status;
  return typeof status === "string" ? status.toLowerCase() : null;
}

function hasIgnoredPathSegment(path: string): boolean {
  return path
    .split("/")
    .some((segment) => segment.startsWith(".") || segment === "node_modules");
}

function fileExtension(path: string): string {
  return path.split(".").pop()?.toLowerCase() ?? "";
}

function deliverableMimeType(path: string): string | undefined {
  const extension = fileExtension(path);
  return (
    deliverableTypes[extension] ??
    (hasDeliverableRoot(path) ? textDeliverableTypes[extension] : undefined)
  );
}

function hasDeliverableRoot(path: string): boolean {
  return ["artifacts", "deliverables", "output", "outputs"].includes(
    path.split("/")[0],
  );
}

const deliverableTypes: Record<string, string> = {
  aac: "audio/aac",
  avif: "image/avif",
  flac: "audio/flac",
  gif: "image/gif",
  htm: "text/html",
  html: "text/html",
  jpeg: "image/jpeg",
  jpg: "image/jpeg",
  m4a: "audio/mp4",
  m4v: "video/x-m4v",
  mov: "video/quicktime",
  mp3: "audio/mpeg",
  mp4: "video/mp4",
  docx: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  pdf: "application/pdf",
  png: "image/png",
  ppt: "application/vnd.ms-powerpoint",
  pptx: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
  svg: "image/svg+xml",
  wav: "audio/wav",
  webm: "video/webm",
  webp: "image/webp",
  xls: "application/vnd.ms-excel",
  xlsx: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
};

const textDeliverableTypes: Record<string, string> = {
  cjs: "text/javascript",
  css: "text/css",
  csv: "text/csv",
  go: "text/x-go",
  js: "text/javascript",
  jsx: "text/javascript",
  json: "application/json",
  md: "text/markdown",
  mjs: "text/javascript",
  py: "text/x-python",
  rs: "text/x-rust",
  scss: "text/x-scss",
  sh: "text/x-shellscript",
  toml: "text/plain",
  ts: "text/typescript",
  tsx: "text/typescript",
  txt: "text/plain",
  yaml: "text/yaml",
  yml: "text/yaml",
};
