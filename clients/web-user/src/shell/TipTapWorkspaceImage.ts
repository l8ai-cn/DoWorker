import type { AnyExtension } from "@tiptap/core";
import { Image } from "@tiptap/extension-image";
import { Link } from "@tiptap/extension-link";
import { createWorkspaceImageNodeView } from "./WorkspaceImageNodeView";

export {
  isWorkspaceRelativeSrc,
  resolveWorkspacePath,
} from "./workspaceImagePaths";

export const ImageAwareLink = Link.extend({
  parseMarkdown: (token, helpers) => {
    const attrs = { href: token.href, title: token.title || null };
    const content = helpers.parseInline(token.tokens || []);
    if (content.length > 0 && content.every((node) => node.type === "image")) {
      return content.map((node) => ({
        ...node,
        marks: [...(node.marks ?? []), { type: "link", attrs }],
      }));
    }
    return helpers.applyMark("link", content, attrs);
  },
});

function escapeHtmlAttribute(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/"/g, "&quot;")
    .replace(/</g, "&lt;");
}

function escapeMarkdownLabel(value: string): string {
  return value.replace(/([\\[\]])/g, "\\$1");
}

function markdownDestination(destination: string): string {
  let needsBrackets = /[\s<>]/.test(destination);
  let depth = 0;
  for (const character of destination) {
    if (character === "(") depth += 1;
    else if (character === ")") depth -= 1;
    if (depth < 0) break;
  }
  needsBrackets ||= depth !== 0;
  return needsBrackets
    ? `<${destination.replace(/([\\<>])/g, "\\$1")}>`
    : destination;
}

function markdownTitle(title: string | null | undefined): string {
  return title ? ` "${title.replace(/([\\"])/g, "\\$1")}"` : "";
}

export function createWorkspaceImageExtension(
  conversationId: string,
  filePath: string,
): AnyExtension {
  return Image.extend({
    marks: "link",
    addAttributes() {
      return {
        ...this.parent?.(),
        valign: { default: null },
        align: { default: null },
      };
    },
    renderMarkdown(node) {
      const src: string = node.attrs?.src ?? "";
      const alt: string = node.attrs?.alt ?? "";
      const title: string = node.attrs?.title ?? "";
      const width = node.attrs?.width;
      const height = node.attrs?.height;
      const valign = node.attrs?.valign;
      const align = node.attrs?.align;
      const needsHtml =
        width != null ||
        height != null ||
        valign != null ||
        align != null;
      let image: string;

      if (needsHtml) {
        const attributes = [`src="${escapeHtmlAttribute(src)}"`];
        if (node.attrs?.alt != null) {
          attributes.push(`alt="${escapeHtmlAttribute(alt)}"`);
        }
        if (title) attributes.push(`title="${escapeHtmlAttribute(title)}"`);
        if (width != null) {
          attributes.push(`width="${escapeHtmlAttribute(String(width))}"`);
        }
        if (height != null) {
          attributes.push(`height="${escapeHtmlAttribute(String(height))}"`);
        }
        if (valign != null) {
          attributes.push(`valign="${escapeHtmlAttribute(String(valign))}"`);
        }
        if (align != null) {
          attributes.push(`align="${escapeHtmlAttribute(String(align))}"`);
        }
        image = `<img ${attributes.join(" ")} />`;
      } else {
        image = `![${escapeMarkdownLabel(alt)}](${markdownDestination(src)}${markdownTitle(title)})`;
      }

      const linkMark = (node.marks ?? []).find(
        (mark) => (typeof mark === "string" ? mark : mark.type) === "link",
      );
      const href =
        typeof linkMark === "object" ? linkMark.attrs?.href : undefined;
      if (!href) return image;
      const linkTitle =
        typeof linkMark === "object" ? linkMark.attrs?.title : undefined;
      return `[${image}](${markdownDestination(String(href))}${markdownTitle(
        linkTitle as string | null,
      )})`;
    },
    addNodeView() {
      return createWorkspaceImageNodeView(conversationId, filePath);
    },
  }).configure({ inline: true });
}
