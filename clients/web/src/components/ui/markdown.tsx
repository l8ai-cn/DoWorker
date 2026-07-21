"use client";

import { useMemo } from "react";
import ReactMarkdown, {
  defaultUrlTransform,
  type Components,
  type ExtraProps,
  type UrlTransform,
} from "react-markdown";
import remarkGfm from "remark-gfm";
import { markdownImageSource } from "@agent-cloud/agent-ui";
import { cn } from "@/lib/utils";
import { LightboxImage } from "@/components/media/MediaLightbox";
import { HtmlPreviewCard } from "@/components/media/HtmlPreviewCard";

interface MarkdownProps {
  content: string;
  className?: string;
  compact?: boolean;
  highlightMentions?: boolean;
  enableMedia?: boolean;
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
            <span key={i} className="text-primary font-medium bg-primary/10 rounded px-0.5">
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

type HastElement = NonNullable<ExtraProps["node"]>;
type ElementContent = HastElement["children"][number];
function hastText(node: ElementContent | HastElement): string {
  if (node.type === "text") return node.value;
  if (node.type === "element") return node.children.map(hastText).join("");
  return "";
}

function codeLanguage(codeEl: HastElement): string | null {
  const cls = codeEl.properties?.className;
  const classes = Array.isArray(cls) ? cls.map(String) : typeof cls === "string" ? [cls] : [];
  const lang = classes.find((c) => c.startsWith("language-"));
  return lang ? lang.slice("language-".length).toLowerCase() : null;
}

const mediaUrlTransform: UrlTransform = (url, key, node) =>
  node.tagName === "img" && key === "src" ? url : defaultUrlTransform(url);

function buildMarkdownComponents(
  withMentions: boolean,
  enableMedia: boolean,
  streaming: boolean,
): Components {
  return {
    img({ src, alt }) {
      const source = typeof src === "string" ? src : undefined;
      const inlineSource = markdownImageSource(source);
      if (inlineSource) {
        return (
          <LightboxImage
            src={inlineSource}
            alt={typeof alt === "string" ? alt : undefined}
            className="my-1 max-w-md"
            imgClassName="max-h-80"
          />
        );
      }
      const label = typeof alt === "string" && alt.trim() ? alt : source;
      if (source && /^https?:\/\//i.test(source)) {
        return (
          <a
            href={source}
            target="_blank"
            rel="noopener noreferrer"
            referrerPolicy="no-referrer"
          >
            {label}
          </a>
        );
      }
      return <span>{label}</span>;
    },
    p({ children }) {
      return <p>{withMentions ? processMentions(children) : children}</p>;
    },
    ...(enableMedia
      ? {
          pre({ node, children }) {
            const first = node?.children?.[0];
            if (
              first &&
              first.type === "element" &&
              first.tagName === "code" &&
              codeLanguage(first) === "html"
            ) {
              return <HtmlPreviewCard html={hastText(first)} streaming={streaming} />;
            }
            return <pre>{children}</pre>;
          },
        }
      : {}),
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
  const components = useMemo(
    () => buildMarkdownComponents(highlightMentions, enableMedia, mediaStreaming),
    [enableMedia, highlightMentions, mediaStreaming],
  );

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
        urlTransform={mediaUrlTransform}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}

export default Markdown;
