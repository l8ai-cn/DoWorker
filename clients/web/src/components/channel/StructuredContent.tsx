"use client";

import type { MessageContent, Block } from "@/lib/viewModels/channelMessage";
import { cn } from "@/lib/utils";
import { classifyMediaUrl } from "@/lib/media/url";
import { LightboxImage } from "@/components/media/MediaLightbox";
import { VideoEmbed } from "@/components/media/VideoEmbed";
import { HtmlPreviewCard } from "@/components/media/HtmlPreviewCard";
import { RenderInline } from "./StructuredInline";
import { RenderTable } from "./StructuredTable";

interface StructuredContentProps {
  content: MessageContent;
  className?: string;
}

const SUPPORTED_SCHEMA_VERSION = 1;

export function StructuredContent({ content, className }: StructuredContentProps) {
  if (!content.blocks?.length) return null;

  if (content.schema_version && content.schema_version > SUPPORTED_SCHEMA_VERSION) {
    return (
      <p className="text-sm text-muted-foreground italic">
        This message uses a newer format. Please update your client.
      </p>
    );
  }

  return (
    <div className={cn("prose prose-sm max-w-none [&_li>p:only-child]:m-0", className)}>
      {content.blocks.map((block, i) => (
        <RenderBlock key={i} block={block} />
      ))}
    </div>
  );
}

// Returns the URL when the paragraph consists solely of one link (plus
// optional whitespace text), which is the signal to upgrade it to a media
// embed.
function singleLinkUrl(block: Block): string | null {
  if (!block.elements?.length || block.children?.length) return null;
  let url: string | null = null;
  for (const el of block.elements) {
    if (el.type === "link") {
      if (url) return null;
      if (!el.url) return null;
      url = el.url;
    } else if (el.type === "text") {
      if ((el.text ?? "").trim() !== "") return null;
    } else if (el.type !== "linebreak") {
      return null;
    }
  }
  return url;
}

function RenderBlock({ block }: { block: Block }) {
  switch (block.type) {
    case "paragraph": {
      const linkUrl = singleLinkUrl(block);
      if (linkUrl) {
        const kind = classifyMediaUrl(linkUrl);
        if (kind === "image") {
          return (
            <LightboxImage
              src={linkUrl}
              className="my-1 max-w-[320px]"
              imgClassName="max-h-[240px]"
            />
          );
        }
        if (kind === "html") {
          return <HtmlPreviewCard src={linkUrl} className="max-w-xl" />;
        }
        if (kind !== "link") {
          return <VideoEmbed url={linkUrl} kind={kind} className="my-2" />;
        }
      }
      return (
        <>
          <p>
            {block.elements?.map((el, i) => (
              <RenderInline key={i} element={el} />
            ))}
          </p>
          <BlockChildren block={block} />
        </>
      );
    }
    case "code_block":
      if (block.language?.toLowerCase() === "html" && block.text?.trim()) {
        return <HtmlPreviewCard html={block.text} />;
      }
      return (
        <pre className="p-3 bg-muted rounded-md text-sm overflow-x-auto">
          <code>{block.text}</code>
        </pre>
      );
    case "heading": {
      const Tag = `h${Math.min(block.level ?? 1, 3)}` as "h1" | "h2" | "h3";
      return (
        <>
          <Tag>
            {block.elements?.map((el, i) => (
              <RenderInline key={i} element={el} />
            ))}
          </Tag>
          <BlockChildren block={block} />
        </>
      );
    }
    case "quote":
      return (
        <blockquote>
          {block.elements?.map((el, i) => (
            <RenderInline key={i} element={el} />
          ))}
          <BlockChildren block={block} />
        </blockquote>
      );
    case "list": {
      const Tag = block.ordered ? "ol" : "ul";
      return (
        <>
          <Tag>
            {block.items?.map((item, i) => (
              <li key={i}>
                {item.map((child, j) => (
                  <RenderBlock key={j} block={child} />
                ))}
              </li>
            ))}
          </Tag>
          <BlockChildren block={block} />
        </>
      );
    }
    case "table":
      return <RenderTable block={block} />;
    default:
      return null;
  }
}

function BlockChildren({ block }: { block: Block }) {
  if (!block.children?.length) return null;
  return <>{block.children.map((child, i) => <RenderBlock key={i} block={child} />)}</>;
}

export default StructuredContent;
