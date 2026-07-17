import type { NodeViewRenderer } from "@tiptap/core";
import { markdownImageSource } from "@do-worker/agent-ui";
import {
  fetchFileContent,
  fileContentToBlob,
} from "@/hooks/useFileContent";
import {
  isWorkspaceRelativeSrc,
  resolveWorkspacePath,
} from "./workspaceImagePaths";

function applyImageAttribute(
  image: HTMLImageElement,
  key: string,
  value: unknown,
) {
  if ((key === "width" || key === "height") && /^\d+$/.test(String(value))) {
    image.style[key] = `${value}px`;
    image.removeAttribute(key);
    return;
  }
  image.setAttribute(key, String(value));
}

function syncImageAttributes(
  image: HTMLImageElement,
  attributes: Record<string, unknown>,
) {
  for (const [key, value] of Object.entries(attributes)) {
    if (key === "src") continue;
    if (value == null) {
      image.removeAttribute(key);
      if (key === "width" || key === "height") image.style[key] = "";
    } else {
      applyImageAttribute(image, key, value);
    }
  }
}

function remoteImageLabel(alt: unknown): string {
  const label = String(alt || "image");
  return document.documentElement.lang.toLowerCase().startsWith("zh")
    ? `加载图片：${label}`
    : `Load image: ${label}`;
}

function isUserLoadableRemoteSource(src: string): boolean {
  return /^(https?:|\/\/)/i.test(src);
}

function addRemoteImageAction(
  container: HTMLElement,
  image: HTMLImageElement,
  src: string,
  alt: unknown,
): HTMLSpanElement | null {
  if (!isUserLoadableRemoteSource(src)) return null;
  const action = document.createElement("span");
  action.dataset.remoteImageAction = "true";
  action.role = "button";
  action.tabIndex = 0;
  action.className =
    "cursor-pointer text-sm text-primary underline underline-offset-2";
  action.textContent = remoteImageLabel(alt);
  const load = (event: Event) => {
    event.preventDefault();
    event.stopPropagation();
    image.src = src;
    action.remove();
  };
  action.addEventListener("click", load);
  action.addEventListener("keydown", (event) => {
    if (event.key === "Enter" || event.key === " ") load(event);
  });
  container.appendChild(action);
  return action;
}

export function createWorkspaceImageNodeView(
  conversationId: string,
  filePath: string,
): NodeViewRenderer {
  return ({ node: initialNode, HTMLAttributes }) => {
    let node = initialNode;
    let cancelled = false;
    let objectUrl: string | null = null;
    const container = document.createElement("span");
    container.contentEditable = "false";
    const image = document.createElement("img");
    container.appendChild(image);
    syncImageAttributes(image, HTMLAttributes);
    const src: string = node.attrs.src ?? "";
    let remoteAction: HTMLSpanElement | null = null;

    const safeInlineSource = markdownImageSource(src);
    if (safeInlineSource) {
      image.src = safeInlineSource;
    } else if (src && isWorkspaceRelativeSrc(src)) {
      const pathPart = src.split(/[?#]/)[0];
      const resolved =
        pathPart === "" ? "" : resolveWorkspacePath(filePath, src);
      if (resolved) {
        fetchFileContent(conversationId, resolved)
          .then((data) => {
            if (cancelled) return;
            objectUrl = URL.createObjectURL(fileContentToBlob(data));
            image.src = objectUrl;
          })
          .catch(() => undefined);
      }
    } else if (src) {
      remoteAction = addRemoteImageAction(
        container,
        image,
        src,
        node.attrs.alt,
      );
    }

    return {
      dom: container,
      update(newNode) {
        if (newNode.type !== node.type || newNode.attrs.src !== src) return false;
        syncImageAttributes(image, newNode.attrs);
        if (remoteAction) {
          remoteAction.textContent = remoteImageLabel(newNode.attrs.alt);
        }
        node = newNode;
        return true;
      },
      destroy() {
        cancelled = true;
        if (objectUrl) URL.revokeObjectURL(objectUrl);
      },
    };
  };
}
