import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

export function MarkdownMessage({ text }: { text: string }) {
  return (
    <div className="min-w-0 text-sm leading-6">
      <ReactMarkdown
        components={{
          a: ({ children, ...props }) => (
            <a
              {...props}
              className="font-medium text-primary underline underline-offset-4"
              rel="noreferrer"
              target="_blank"
            >
              {children}
            </a>
          ),
          blockquote: ({ children }) => (
            <blockquote className="my-3 border-l-2 border-border pl-3 text-muted-foreground">
              {children}
            </blockquote>
          ),
          code: ({ children, className }) =>
            className ? (
              <code className={`${className} font-mono text-[13px]`}>{children}</code>
            ) : (
              <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-[0.9em]">
                {children}
              </code>
            ),
          h1: ({ children }) => (
            <h1 className="mb-2 mt-5 text-xl font-semibold first:mt-0">{children}</h1>
          ),
          h2: ({ children }) => (
            <h2 className="mb-2 mt-5 text-lg font-semibold first:mt-0">{children}</h2>
          ),
          h3: ({ children }) => (
            <h3 className="mb-1.5 mt-4 text-base font-semibold first:mt-0">{children}</h3>
          ),
          li: ({ children }) => <li className="ml-5 list-disc pl-1">{children}</li>,
          ol: ({ children }) => <ol className="my-3 space-y-1">{children}</ol>,
          p: ({ children }) => <p className="my-2 first:mt-0 last:mb-0">{children}</p>,
          pre: ({ children }) => (
            <pre className="my-3 overflow-x-auto rounded-md border border-border bg-muted/50 p-3 text-foreground">
              {children}
            </pre>
          ),
          table: ({ children }) => (
            <div className="my-3 overflow-x-auto">
              <table className="w-full border-collapse text-left text-sm">{children}</table>
            </div>
          ),
          td: ({ children }) => (
            <td className="border border-border px-2 py-1.5">{children}</td>
          ),
          th: ({ children }) => (
            <th className="border border-border bg-muted px-2 py-1.5 font-medium">
              {children}
            </th>
          ),
          ul: ({ children }) => <ul className="my-3 space-y-1">{children}</ul>,
        }}
        remarkPlugins={[remarkGfm]}
      >
        {text}
      </ReactMarkdown>
    </div>
  );
}
