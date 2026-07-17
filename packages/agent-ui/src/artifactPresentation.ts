export type ArtifactKind =
  | "image"
  | "video"
  | "pdf"
  | "presentation"
  | "html"
  | "code"
  | "text"
  | "file";

export function artifactPresentation(
  mimeType: string | null,
  filename: string,
): { kind: ArtifactKind; label: string } {
  const mime = mimeType?.toLowerCase() ?? "";
  const extension = filename.split(".").pop()?.toLowerCase() ?? "";
  if (mime === "text/html" || extension === "html" || extension === "htm") {
    return { kind: "html", label: "HTML document" };
  }
  if (mime === "image/svg+xml" || extension === "svg") {
    return { kind: "file", label: "SVG document" };
  }
  if (mime.startsWith("image/") || imageExtensions.has(extension)) {
    return { kind: "image", label: "Image" };
  }
  if (mime.startsWith("video/") || videoExtensions.has(extension)) {
    return { kind: "video", label: "Video" };
  }
  if (mime === "application/pdf" || extension === "pdf") {
    return { kind: "pdf", label: "PDF" };
  }
  if (
    mime.includes("presentation") ||
    extension === "ppt" ||
    extension === "pptx"
  ) {
    return { kind: "presentation", label: "PowerPoint" };
  }
  if (codeMimeTypes.has(mime) || codeExtensions.has(extension)) {
    return { kind: "code", label: "Code file" };
  }
  if (mime.startsWith("text/") || textExtensions.has(extension)) {
    return { kind: "text", label: "Text file" };
  }
  return { kind: "file", label: mimeType || "File" };
}

const imageExtensions = new Set(["avif", "gif", "jpeg", "jpg", "png", "webp"]);
const videoExtensions = new Set(["mov", "mp4", "m4v", "webm"]);
const codeMimeTypes = new Set([
  "application/javascript",
  "application/json",
  "application/typescript",
  "text/css",
  "text/javascript",
  "text/typescript",
]);
const codeExtensions = new Set([
  "cjs",
  "css",
  "go",
  "js",
  "jsx",
  "json",
  "mjs",
  "py",
  "rs",
  "scss",
  "sh",
  "ts",
  "tsx",
]);
const textExtensions = new Set(["md", "txt", "toml", "yaml", "yml"]);
