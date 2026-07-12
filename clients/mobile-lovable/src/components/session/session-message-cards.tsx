import {
  Brain,
  Check,
  ChevronDown,
  CircleAlert,
  ListChecks,
  Loader2,
  MessageSquare,
  User,
} from "lucide-react";
import { useState } from "react";
import { Lightbox } from "@/components/lightbox";
import { Markdown } from "@/components/markdown";
import type { AgentEvent, PlanItem } from "@/lib/session-types";
import { dedupeRepeatedText } from "@/lib/text-normalizer";
import { cn } from "@/lib/utils";

export function PhaseCard({ event }: { event: AgentEvent }) {
  return (
    <div className="stream-in my-3 flex items-center gap-3">
      <div className="h-px flex-1 bg-gradient-to-r from-transparent to-primary/40" />
      <div className="flex items-center gap-2 rounded-full border border-primary/40 bg-gradient-to-br from-primary/20 to-accent/10 px-3 py-1.5 shadow-[0_0_20px_-8px] shadow-primary/50">
        <span className="text-base leading-none">{event.phaseEmoji ?? "▶"}</span>
        <div className="text-left">
          <p className="text-[9.5px] font-semibold uppercase tracking-wider text-primary">
            Phase {event.phaseIndex}/{event.phaseTotal}
          </p>
          <p className="text-[12.5px] font-semibold leading-tight">{event.title}</p>
        </div>
      </div>
      <div className="h-px flex-1 bg-gradient-to-l from-transparent to-primary/40" />
    </div>
  );
}

export function UserBubble({ event }: { event: AgentEvent }) {
  const text = event.markdown ?? event.detail;
  return (
    <div className="stream-in flex items-start justify-end gap-2.5 pl-10 pt-1">
      <div className="max-w-[88%] space-y-1.5 rounded-2xl rounded-tr-md bg-primary/12 px-3 py-2.5 ring-1 ring-primary/20">
        {text && (
          <p className="whitespace-pre-wrap text-[13px] leading-[1.55] text-foreground">{text}</p>
        )}
        {event.attachments?.map((a, i) =>
          a.kind === "image" && a.src ? (
            <figure key={i} className="overflow-hidden rounded-lg ring-1 ring-primary/20">
              <Lightbox
                src={a.src}
                alt={a.name}
                caption={a.note}
                thumbClassName="max-h-56 w-full object-cover"
              />
              {a.note && (
                <figcaption className="bg-background/60 px-2 py-1 font-mono text-[10px] text-muted-foreground">
                  {a.note}
                </figcaption>
              )}
            </figure>
          ) : (
            <div key={i} className="rounded-md bg-background/40 px-2 py-1 font-mono text-[10.5px]">
              📎 {a.name}
            </div>
          ),
        )}

        <p className="text-right font-mono text-[9.5px] text-muted-foreground">{event.ts}</p>
      </div>
      <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary/15 ring-1 ring-primary/25">
        <User className="h-3.5 w-3.5 text-primary" />
      </div>
    </div>
  );
}

export function AgentBubble({ event }: { event: AgentEvent }) {
  const body = event.markdown ? dedupeRepeatedText(event.markdown) : (event.detail ?? event.title);
  return (
    <div className="stream-in flex items-start gap-2.5 pr-6">
      <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-accent/15 ring-1 ring-accent/25">
        <MessageSquare className="h-3.5 w-3.5 text-accent" />
      </div>
      <div className="min-w-0 max-w-[88%] space-y-1.5 rounded-2xl rounded-tl-md bg-card px-3 py-2.5 ring-1 ring-border/50">
        {event.markdown ? (
          <Markdown className="text-[13px] leading-[1.55]">{body}</Markdown>
        ) : (
          body && (
            <p className="whitespace-pre-wrap text-[13px] leading-[1.55] text-foreground">{body}</p>
          )
        )}
        {event.images?.map((img, i) => (
          <figure key={i} className="overflow-hidden rounded-lg ring-1 ring-border/50">
            <Lightbox
              src={img.src}
              alt={img.alt ?? ""}
              caption={img.caption}
              thumbClassName="w-full object-cover"
            />
            {img.caption && (
              <figcaption className="bg-surface/60 px-2.5 py-1.5 text-[10.5px] text-muted-foreground">
                {img.caption}
              </figcaption>
            )}
          </figure>
        ))}

        <p className="font-mono text-[9.5px] text-muted-foreground/80">{event.ts}</p>
      </div>
    </div>
  );
}

export function ThoughtBlock({ event }: { event: AgentEvent }) {
  const [open, setOpen] = useState(false);
  return (
    <button
      onClick={() => setOpen((o) => !o)}
      className="stream-in flex w-full items-start gap-2 rounded-lg px-2 py-1.5 text-left hover:bg-surface/60"
    >
      <Brain className="mt-0.5 h-3.5 w-3.5 shrink-0 text-event-thinking" />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
          <span className="italic">Thinking</span>
          <span className="font-mono text-[9.5px]">· {event.ts}</span>
          <ChevronDown className={cn("h-3 w-3 transition-transform", open && "rotate-180")} />
        </div>
        {open && (
          <p className="mt-1 whitespace-pre-wrap border-l-2 border-event-thinking/40 pl-2 text-[12px] italic leading-relaxed text-muted-foreground">
            {event.markdown ?? event.detail}
          </p>
        )}
      </div>
    </button>
  );
}

export function PlanCard({ event }: { event: AgentEvent }) {
  const items = event.plan ?? [];
  const done = items.filter((i) => i.status === "completed").length;
  return (
    <div className="stream-in rounded-xl border border-info/30 bg-info/5 p-3">
      <div className="mb-2 flex items-center justify-between">
        <div className="flex items-center gap-1.5">
          <ListChecks className="h-3.5 w-3.5 text-info" />
          <span className="text-[12px] font-semibold text-info">{event.title}</span>
        </div>
        <span className="font-mono text-[10px] text-muted-foreground">{done}/{items.length}</span>
      </div>
      <ul className="space-y-1.5">
        {items.map((item, i) => <PlanRow key={i} item={item} />)}
      </ul>
    </div>
  );
}

export function PlanRow({ item }: { item: PlanItem }) {
  return (
    <li className="flex items-start gap-2 text-[12px]">
      <span className={cn(
        "mt-0.5 flex h-3.5 w-3.5 shrink-0 items-center justify-center rounded-full ring-1",
        item.status === "completed" && "bg-success/20 ring-success/50",
        item.status === "in_progress" && "bg-primary/20 ring-primary/50",
        item.status === "pending" && "bg-transparent ring-border",
      )}>
        {item.status === "completed" && <Check className="h-2.5 w-2.5 text-success" />}
        {item.status === "in_progress" && <Loader2 className="h-2.5 w-2.5 animate-spin text-primary" />}
      </span>
      <span className={cn(
        "leading-snug",
        item.status === "completed" && "text-muted-foreground line-through",
        item.status === "in_progress" && "text-foreground",
        item.status === "pending" && "text-muted-foreground",
      )}>{item.text}</span>
    </li>
  );
}

export function ErrorCard({ event }: { event: AgentEvent }) {
  return (
    <div className="stream-in flex items-start gap-2 rounded-xl border border-destructive/40 bg-destructive/5 p-3">
      <CircleAlert className="mt-0.5 h-4 w-4 shrink-0 text-destructive" />
      <div className="min-w-0 flex-1">
        <p className="text-[12.5px] font-semibold text-destructive">{event.title}</p>
        {event.detail && <p className="mt-0.5 text-[12px] text-muted-foreground">{event.detail}</p>}
      </div>
      <span className="font-mono text-[9.5px] text-muted-foreground">{event.ts}</span>
    </div>
  );
}
