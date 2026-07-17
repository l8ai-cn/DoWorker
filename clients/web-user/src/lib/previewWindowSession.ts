import { parsePreviewSessionUrl } from "./previewSessionUrl";

const SESSION_PATH_PATTERN =
  /^\/preview\/([a-z0-9]+(?:-[a-z0-9]+)*)\/__session$/;

export function readPreviewWindowSessionUrl(
  hash: string,
  publicOrigin: string,
): string {
  try {
    const encodedUrl = hash.startsWith("#") ? hash.slice(1) : hash;
    const rawUrl = decodeURIComponent(encodedUrl);
    const candidate = new URL(rawUrl);
    const podKey = SESSION_PATH_PATTERN.exec(candidate.pathname)?.[1];
    if (!podKey) {
      throw new Error("missing Pod key");
    }
    return parsePreviewSessionUrl(rawUrl, podKey, publicOrigin).href;
  } catch {
    throw new Error("预览地址无效");
  }
}
