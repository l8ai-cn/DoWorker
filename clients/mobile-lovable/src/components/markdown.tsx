import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { cn } from "@/lib/utils";
import { Lightbox } from "@/components/lightbox";


export function Markdown({ children, className }: { children: string; className?: string }) {
  return (
    <div className={cn("space-y-2 text-[12.5px] leading-relaxed text-foreground", className)}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          p: ({ children }) => <p className="whitespace-pre-wrap">{children}</p>,
          h1: ({ children }) => <h1 className="mt-3 text-[15px] font-semibold">{children}</h1>,
          h2: ({ children }) => <h2 className="mt-3 text-[13.5px] font-semibold text-foreground">{children}</h2>,
          h3: ({ children }) => <h3 className="mt-2 text-[12.5px] font-semibold uppercase tracking-wider text-muted-foreground">{children}</h3>,
          strong: ({ children }) => <strong className="font-semibold text-foreground">{children}</strong>,
          em: ({ children }) => <em className="italic text-foreground/85">{children}</em>,
          a: ({ children, href }) => (
            <a href={href} target="_blank" rel="noreferrer" className="text-primary underline underline-offset-2 hover:text-primary/80">
              {children}
            </a>
          ),
          ul: ({ children }) => <ul className="ml-4 list-disc space-y-1 marker:text-muted-foreground">{children}</ul>,
          ol: ({ children }) => <ol className="ml-4 list-decimal space-y-1 marker:text-muted-foreground">{children}</ol>,
          li: ({ children }) => <li className="[&>p]:inline">{children}</li>,
          blockquote: ({ children }) => (
            <blockquote className="border-l-2 border-primary/40 bg-primary/5 py-1 pl-3 text-foreground/85">
              {children}
            </blockquote>
          ),
          code: ({ className, children }) => {
            const isBlock = /language-/.test(className ?? "");
            if (isBlock) {
              return (
                <pre className="overflow-x-auto rounded-lg bg-[#0a0d12] p-2.5 font-mono text-[11px] leading-relaxed ring-1 ring-border/40">
                  <code>{children}</code>
                </pre>
              );
            }
            return (
              <code className="rounded bg-surface-2 px-1 py-0.5 font-mono text-[11px] text-foreground">
                {children}
              </code>
            );
          },
          table: ({ children }) => (
            <div className="overflow-x-auto rounded-lg ring-1 ring-border/40">
              <table className="w-full border-collapse text-[11.5px]">{children}</table>
            </div>
          ),
          thead: ({ children }) => <thead className="bg-surface">{children}</thead>,
          th: ({ children }) => <th className="border-b border-border/60 px-2.5 py-1.5 text-left font-semibold">{children}</th>,
          td: ({ children }) => <td className="border-b border-border/30 px-2.5 py-1.5 text-foreground/85">{children}</td>,
          hr: () => <hr className="border-border/50" />,
          img: ({ src, alt }) =>
            typeof src === "string" ? (
              <Lightbox
                src={src}
                alt={alt ?? ""}
                thumbClassName="rounded-lg ring-1 ring-border/40"
              />
            ) : null,

        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  );
}
