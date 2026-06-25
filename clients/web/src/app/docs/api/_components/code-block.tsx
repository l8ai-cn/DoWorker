"use client";

export function JsonBlock({
  children,
  tone,
}: {
  children: string;
  tone?: "success";
}) {
  return (
    <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 font-mono text-sm overflow-x-auto">
      <pre className={tone === "success" ? "text-success" : undefined}>
        {children}
      </pre>
    </div>
  );
}
