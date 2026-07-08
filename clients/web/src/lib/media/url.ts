// Shared URL classification and safety helpers for inline media previews
// in chat surfaces (ACP activity stream + channel messages). Mirrors the
// provider detection logic used by the blockstore renderers so both systems
// embed the same providers the same way.

export { isSafeURL, sanitizeURL } from "@/lib/blockstore/urlGuard";
import { isSafeURL } from "@/lib/blockstore/urlGuard";

export type MediaKind =
  | "image"
  | "video"
  | "audio"
  | "html"
  | "youtube"
  | "vimeo"
  | "loom"
  | "figma"
  | "codesandbox"
  | "link";

export const IMAGE_EXTS = ["jpg", "jpeg", "png", "gif", "webp", "svg", "bmp", "avif"];
export const VIDEO_EXTS = ["mp4", "webm", "mov", "m4v", "ogv"];
export const AUDIO_EXTS = ["mp3", "wav", "m4a", "flac", "ogg", "aac"];
export const HTML_EXTS = ["html", "htm"];

export function extOf(url: string): string {
  try {
    const clean = url.split("?")[0].split("#")[0];
    const lastSegment = clean.slice(clean.lastIndexOf("/") + 1);
    const dot = lastSegment.lastIndexOf(".");
    if (dot < 0) return "";
    return lastSegment.slice(dot + 1).toLowerCase();
  } catch {
    return "";
  }
}

// isRelativePath: no scheme at all (same-origin resource). Scheme-relative
// (//host) and custom schemes (javascript:, data:) are NOT relative.
function isRelativePath(url: string): boolean {
  if (!url || url.startsWith("//")) return false;
  return !/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(url);
}

// classifyMediaUrl maps a URL to a preview kind. Absolute URLs must be
// http(s); same-origin relative paths are classified by extension only.
// Unsafe or unrecognized URLs come back as plain "link" so callers fall
// through to a regular anchor.
export function classifyMediaUrl(url: string): MediaKind {
  if (!isSafeURL(url)) {
    if (!isRelativePath(url)) return "link";
    const ext = extOf(url);
    if (IMAGE_EXTS.includes(ext)) return "image";
    if (VIDEO_EXTS.includes(ext)) return "video";
    if (AUDIO_EXTS.includes(ext)) return "audio";
    if (HTML_EXTS.includes(ext)) return "html";
    return "link";
  }

  if (/youtu\.be\/|youtube\.com\//.test(url)) return "youtube";
  if (/vimeo\.com\//.test(url)) return "vimeo";
  if (/loom\.com\/share\//.test(url)) return "loom";
  if (/figma\.com\//.test(url)) return "figma";
  if (/codesandbox\.io\//.test(url)) return "codesandbox";

  const ext = extOf(url);
  if (IMAGE_EXTS.includes(ext)) return "image";
  if (VIDEO_EXTS.includes(ext)) return "video";
  if (AUDIO_EXTS.includes(ext)) return "audio";
  if (HTML_EXTS.includes(ext)) return "html";
  return "link";
}

// buildEmbedURL converts a provider page URL into its iframe embed src.
// Returns null when the URL doesn't match the provider's expected shape,
// in which case callers should fall back to a plain link.
export function buildEmbedURL(url: string, kind: MediaKind): string | null {
  switch (kind) {
    case "youtube": {
      const m = url.match(/(?:v=|youtu\.be\/|shorts\/|embed\/)([\w-]{6,})/);
      return m ? `https://www.youtube.com/embed/${m[1]}` : null;
    }
    case "vimeo": {
      const m = url.match(/vimeo\.com\/(\d+)/);
      return m ? `https://player.vimeo.com/video/${m[1]}` : null;
    }
    case "loom": {
      const m = url.match(/loom\.com\/share\/([\w-]+)/);
      return m ? `https://www.loom.com/embed/${m[1]}` : null;
    }
    case "figma":
      return `https://www.figma.com/embed?embed_host=agentsmesh&url=${encodeURIComponent(url)}`;
    case "codesandbox":
      return url.includes("/s/") ? url.replace("/s/", "/embed/") : null;
    default:
      return null;
  }
}

// isSafeRenderableSrc accepts absolute http(s) URLs and same-origin relative
// paths — the set of values safe to place in a media element's src.
export function isSafeRenderableSrc(url: string): boolean {
  return isSafeURL(url) || isRelativePath(url);
}

// isSafeImageSrc additionally allows inline data:image/* URIs (used by the
// chat markdown renderer so agent-emitted base64 screenshots display).
export function isSafeImageSrc(src: string): boolean {
  if (!src || typeof src !== "string") return false;
  if (/^data:image\/(png|jpe?g|gif|webp|svg\+xml|bmp|avif);/i.test(src)) return true;
  return isSafeRenderableSrc(src);
}
