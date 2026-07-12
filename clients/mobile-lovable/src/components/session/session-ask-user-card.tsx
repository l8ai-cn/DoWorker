import { Check, MessageSquare } from "lucide-react";
import { useState } from "react";
import type { AgentEvent } from "@/lib/session-types";
import { cn } from "@/lib/utils";

export function AskUserCard({ event }: { event: AgentEvent }) {
  const form = event.form;
  const [values, setValues] = useState<Record<string, string | boolean>>(() => {
    const init: Record<string, string | boolean> = {};
    for (const f of form?.fields ?? []) {
      if (f.type === "checkbox") init[f.name] = f.defaultValue === "true";
      else init[f.name] = f.defaultValue ?? "";
    }
    return init;
  });
  const [submitted, setSubmitted] = useState<Record<string, string | boolean> | null>(
    event.answer ?? null,
  );
  if (!form) return null;

  const set = (k: string, v: string | boolean) => setValues((s) => ({ ...s, [k]: v }));

  if (submitted) {
    return (
      <div className="stream-in rounded-2xl border border-success/40 bg-success/5 p-3">
        <p className="flex items-center gap-1.5 text-[11px] font-semibold text-success">
          <Check className="h-3 w-3" />
          已回答 · {event.ts}
        </p>
        <p className="mt-1 text-[12px] font-medium">{form.title}</p>
        <dl className="mt-2 space-y-0.5 text-[11.5px]">
          {form.fields.map((f) => (
            <div key={f.name} className="flex gap-2">
              <dt className="shrink-0 text-muted-foreground">{f.label}：</dt>
              <dd className="min-w-0 flex-1 break-all font-mono text-foreground/90">
                {String(submitted[f.name] ?? "—") || "—"}
              </dd>
            </div>
          ))}
        </dl>
        <button
          onClick={() => setSubmitted(null)}
          className="mt-2 text-[10.5px] text-primary hover:underline"
        >
          修改回答
        </button>
      </div>
    );
  }

  return (
    <div className="stream-in rounded-2xl border border-primary/40 bg-gradient-to-br from-primary/10 to-transparent p-3.5">
      <p className="flex items-center gap-1.5 text-[10.5px] font-semibold uppercase tracking-wider text-primary">
        <MessageSquare className="h-3 w-3" />
        Agent 询问 · {event.ts}
      </p>
      <p className="mt-1.5 text-[13.5px] font-semibold leading-tight">{form.title}</p>
      {form.description && (
        <p className="mt-1 text-[11.5px] text-muted-foreground">{form.description}</p>
      )}

      <div className="mt-3 space-y-3">
        {form.fields.map((f) => (
          <div key={f.name}>
            <label className="mb-1 block text-[11px] font-medium text-foreground/80">
              {f.label}
              {f.required && <span className="ml-0.5 text-warning">*</span>}
            </label>
            {f.type === "text" && (
              <input
                type="text"
                value={String(values[f.name] ?? "")}
                placeholder={f.placeholder}
                onChange={(e) => set(f.name, e.target.value)}
                className="w-full rounded-lg bg-surface px-2.5 py-1.5 text-[12.5px] outline-none ring-1 ring-border/40 focus:ring-primary/50"
              />
            )}
            {f.type === "number" && (
              <input
                type="number"
                value={String(values[f.name] ?? "")}
                placeholder={f.placeholder}
                onChange={(e) => set(f.name, e.target.value)}
                className="w-full rounded-lg bg-surface px-2.5 py-1.5 text-[12.5px] outline-none ring-1 ring-border/40 focus:ring-primary/50"
              />
            )}
            {f.type === "textarea" && (
              <textarea
                rows={2}
                value={String(values[f.name] ?? "")}
                placeholder={f.placeholder}
                onChange={(e) => set(f.name, e.target.value)}
                className="w-full resize-none rounded-lg bg-surface px-2.5 py-1.5 text-[12.5px] outline-none ring-1 ring-border/40 focus:ring-primary/50"
              />
            )}
            {f.type === "select" && (
              <select
                value={String(values[f.name] ?? "")}
                onChange={(e) => set(f.name, e.target.value)}
                className="w-full rounded-lg bg-surface px-2.5 py-1.5 text-[12.5px] outline-none ring-1 ring-border/40 focus:ring-primary/50"
              >
                {f.options?.map((o) => (
                  <option key={o} value={o}>{o}</option>
                ))}
              </select>
            )}
            {f.type === "radio" && (
              <div className="flex flex-wrap gap-1.5">
                {f.options?.map((o) => (
                  <button
                    key={o}
                    onClick={() => set(f.name, o)}
                    className={cn(
                      "rounded-full px-2.5 py-1 text-[11.5px] font-mono transition",
                      values[f.name] === o
                        ? "bg-primary text-primary-foreground"
                        : "bg-surface text-foreground/70 ring-1 ring-border/40 hover:ring-primary/40",
                    )}
                  >
                    {o}
                  </button>
                ))}
              </div>
            )}
            {f.type === "checkbox" && (
              <button
                onClick={() => set(f.name, !values[f.name])}
                className="flex items-center gap-2 text-[12px] text-foreground/85"
              >
                <span
                  className={cn(
                    "flex h-4 w-4 items-center justify-center rounded border transition",
                    values[f.name]
                      ? "border-primary bg-primary text-primary-foreground"
                      : "border-border bg-surface",
                  )}
                >
                  {values[f.name] && <Check className="h-3 w-3" />}
                </span>
                {f.placeholder ?? "启用"}
              </button>
            )}
          </div>
        ))}
      </div>

      <div className="mt-3 flex items-center gap-2">
        <button
          onClick={() => setSubmitted(values)}
          className="flex-1 rounded-full bg-primary py-2 text-[12.5px] font-semibold text-primary-foreground transition active:scale-[0.98]"
        >
          {form.submitLabel ?? "提交回答"}
        </button>
        <button
          onClick={() => setSubmitted({ __skipped: true })}
          className="rounded-full bg-surface px-3 py-2 text-[11.5px] text-muted-foreground ring-1 ring-border/40"
        >
          跳过
        </button>
      </div>
    </div>
  );
}
