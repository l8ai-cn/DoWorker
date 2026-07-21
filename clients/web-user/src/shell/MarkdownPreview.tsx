import ReactMarkdown, {
  defaultUrlTransform,
  type Components,
  type Options,
  type UrlTransform,
} from "react-markdown";
import remarkGfm from "remark-gfm";
import remarkEmoji from "remark-emoji";
import rehypeRaw from "rehype-raw";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import { rehypeGithubAlerts } from "rehype-github-alerts";
import { markdownImageSource } from "@agent-cloud/agent-ui";

const MARKDOWN_REMARK_PLUGINS = [remarkGfm, remarkEmoji];
const ALERT_CLASS = /^markdown-alert(-\w+)?$/;
const ALERT_TITLE_CLASS = /^markdown-alert-title$/;

const MARKDOWN_SANITIZE_SCHEMA = {
  ...defaultSchema,
  tagNames: defaultSchema.tagNames?.filter(
    (tagName) => tagName !== "picture" && tagName !== "source",
  ),
  attributes: {
    ...defaultSchema.attributes,
    div: [...(defaultSchema.attributes?.div ?? []), ["className", ALERT_CLASS]],
    p: [...(defaultSchema.attributes?.p ?? []), ["className", ALERT_TITLE_CLASS]],
  },
  protocols: {
    ...defaultSchema.protocols,
    src: [...(defaultSchema.protocols?.src ?? []), "blob", "data"],
  },
};

const MARKDOWN_REHYPE_PLUGINS: Options["rehypePlugins"] = [
  rehypeRaw,
  rehypeGithubAlerts,
  [rehypeSanitize, MARKDOWN_SANITIZE_SCHEMA],
];

function imageDimension(value: string | number | undefined): string | undefined {
  if (typeof value === "number") return `${value}px`;
  return typeof value === "string" && /^\d+$/.test(value) ? `${value}px` : undefined;
}

function blockedImageLabel(alt: string | undefined, src: string | undefined): string {
  return alt?.trim() || src || "";
}

const markdownUrlTransform: UrlTransform = (url, key, node) => {
  if (node.tagName === "img" && key === "src") return url;
  return defaultUrlTransform(url);
};

const MARKDOWN_COMPONENTS: Components = {
  img({
    node: _node,
    src,
    srcSet: _srcSet,
    sizes: _sizes,
    alt,
    width,
    height,
    style,
    ...props
  }) {
    const safeSource = markdownImageSource(src);
    if (!safeSource) {
      const label = blockedImageLabel(alt, src);
      if (/^https?:\/\//i.test(src ?? "")) {
        return (
          <a href={src} target="_blank" rel="noopener noreferrer" referrerPolicy="no-referrer">
            {label}
          </a>
        );
      }
      return <span>{label}</span>;
    }

    const sized = {
      ...style,
      width: imageDimension(width) ?? style?.width,
      height: imageDimension(height) ?? style?.height,
    };
    return <img {...props} src={safeSource} alt={alt} style={sized} />;
  },
};

export function MarkdownPreview({ content }: { content: string }) {
  return (
    <div className="markdown-preview prose prose-sm h-full max-w-none overflow-auto px-6 py-4 dark:prose-invert">
      <ReactMarkdown
        remarkPlugins={MARKDOWN_REMARK_PLUGINS}
        rehypePlugins={MARKDOWN_REHYPE_PLUGINS}
        components={MARKDOWN_COMPONENTS}
        urlTransform={markdownUrlTransform}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
