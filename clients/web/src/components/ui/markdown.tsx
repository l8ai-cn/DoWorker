"use client";

import { useMemo } from "react";
import ReactMarkdown, {
  defaultUrlTransform,
  type Components,
  type ExtraProps,
} from "react-markdown";
import remarkGfm from "remark-gfm";
import { cn } from "@/lib/utils";
import { classifyMediaUrl, isSafeImageSrc } from "@/lib/media/url";
import { LightboxImage } from "@/components/media/MediaLightbox";
import { VideoEmbed } from "@/components/media/VideoEmbed";
import { HtmlPreviewCard } from "@/components/media/HtmlPreviewCard";

interface MarkdownProps {
  content: string;
  className?: string;
  compact?: boolean;
  highlightMentions?: boolean;
  /**
   * Upgrades media content in the rendered markdown: images open in a
   * lightbox (data:image/* allowed), single-link paragraphs pointing at
   * media (video / YouTube / .html / …) become inline embeds, and fenced
   * ```html code blocks get a sandboxed live preview.
   */
  enableMedia?: boolean;
  /** While true, html previews stay on the code tab (streaming output). */
  mediaStreaming?: boolean;
}

const remarkPlugins = [remarkGfm];

function TextWithMentions({ children }: { children: string }) {
  const mentionRegex = /(@[\w.\-]+)/g;
  const parts = children.split(mentionRegex);

  return (
    <>
      {parts.map((part, i) => {
        if (mentionRegex.test(part)) {
          mentionRegex.lastIndex = 0;
          return (
            <span
              key={i}
              className="text-primary font-medium bg-primary/10 rounded px-0.5"
            >
              {part}
            </span>
          );
        }
        mentionRegex.lastIndex = 0;
        return part;
      })}
    </>
  );
}

const mentionComponents: Components = {
  p({ children }) {
    return <p>{processMentions(children)}</p>;
  },
  li({ children }) {
    return <li>{processMentions(children)}</li>;
  },
  td({ children }) {
    return <td>{processMentions(children)}</td>;
  },
  th({ children }) {
    return <th>{processMentions(children)}</th>;
  },
};

function processMentions(children: React.ReactNode): React.ReactNode {
  if (!children) return children;
  if (typeof children === "string") {
    return <TextWithMentions>{children}</TextWithMentions>;
  }
  if (Array.isArray(children)) {
    return children.map((child, i) => {
      if (typeof child === "string") {
        return <TextWithMentions key={i}>{child}</TextWithMentions>;
      }
      return child;
    });
  }
  return children;
}

// ---------- media helpers ----------

// hast node types derived from react-markdown's props so we don't depend on
// the (non-hoisted) @types/hast package directly.
type HastElement = NonNullable<ExtraProps["node"]>;
type ElementContent = HastElement["children"][number];

function hastText(node: ElementContent | HastElement): string {
  if (node.type === "text") return node.value;
  if (node.type === "element") return node.children.map(hastText).join("");
  return "";
}

// Returns the href when the paragraph consists solely of one link (plus
// optional whitespace), which is the signal to upgrade it to an embed.
function singleLinkHref(node: HastElement | undefined): string | null {
  if (!node?.children?.length) return null;
  let href: string | null = null;
  for (const child of node.children) {
    if (child.type === "text") {
      if (child.value.trim() !== "") return null;
    } else if (child.type === "element" && child.tagName === "a") {
      if (href) return null;
      const h = child.properties?.href;
      if (typeof h !== "string" || !h) return null;
      href = h;
    } else {
      return null;
    }
  }
  return href;
}

function codeLanguage(codeEl: HastElement): string | null {
  const cls = codeEl.properties?.className;
  const classes = Array.isArray(cls) ? cls.map(String) : typeof cls === "string" ? [cls] : [];
  const lang = classes.find((c) => c.startsWith("language-"));
  return lang ? lang.slice("language-".length).toLowerCase() : null;
}

function mediaUrlTransform(url: string): string {
  return isSafeImageSrc(url) ? url : defaultUrlTransform(url);
}

function buildMediaComponents(withMentions: boolean, streaming: boolean): Components {
  return {
    img({ src, alt }) {
      if (typeof src === "string" && isSafeImageSrc(src)) {
        return (
          <LightboxImage
            src={src}
            alt={typeof alt === "string" ? alt : undefined}
            className="my-1 max-w-md"
            imgClassName="max-h-80"
          />
        );
      }
      return null;
    },
    p({ node, children }) {
      const href = singleLinkHref(node);
      if (href) {
        const kind = classifyMediaUrl(href);
        if (kind === "image") {
          return (
            <LightboxImage src={href} className="my-1 max-w-md" imgClassName="max-h-80" />
          );
        }
        if (kind === "html") {
          return <HtmlPreviewCard src={href} />;
        }
        if (kind !== "link") {
          return <VideoEmbed url={href} kind={kind} className="my-2" />;
        }
      }
      return <p>{withMentions ? processMentions(children) : children}</p>;
    },
    pre({ node, children }) {
      const first = node?.children?.[0];
      if (first && first.type === "element" && first.tagName === "code") {
        if (codeLanguage(first) === "html") {
          return <HtmlPreviewCard html={hastText(first)} streaming={streaming} />;
        }
      }
      return <pre>{children}</pre>;
    },
    ...(withMentions
      ? {
          li({ children }) {
            return <li>{processMentions(children)}</li>;
          },
          td({ children }) {
            return <td>{processMentions(children)}</td>;
          },
          th({ children }) {
            return <th>{processMentions(children)}</th>;
          },
        }
      : {}),
  };
}

export function Markdown({
  content,
  className,
  compact = false,
  highlightMentions = false,
  enableMedia = false,
  mediaStreaming = false,
}: MarkdownProps) {
  const components = useMemo<Components | undefined>(() => {
    if (enableMedia) return buildMediaComponents(highlightMentions, mediaStreaming);
    if (highlightMentions) return mentionComponents;
    return undefined;
  }, [enableMedia, highlightMentions, mediaStreaming]);

  return (
    <div
      className={cn(
        "prose max-w-none",
        compact && "prose-sm",
        compact && "[&_p]:my-1 [&_ul]:my-1 [&_ol]:my-1 [&_li]:my-0.5 [&_h1]:text-base [&_h2]:text-sm [&_h3]:text-xs",
        className
      )}
    >
      <ReactMarkdown
        remarkPlugins={remarkPlugins}
        components={components}
        urlTransform={enableMedia ? mediaUrlTransform : undefined}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}

export default Markdown;
