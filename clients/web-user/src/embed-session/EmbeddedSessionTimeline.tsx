import { useMemo, useState, type FormEvent } from "react";
import { ArrowUpIcon, Loader2Icon, ShieldAlertIcon } from "lucide-react";

import { BlockRenderer } from "@/components/blocks/BlockRenderer";
import { Conversation, ConversationContent, ConversationEmptyState } from "@/components/ai-elements/conversation";
import { Message, MessageContent, MessageResponse } from "@/components/ai-elements/message";
import { Button } from "@/components/ui/button";
import { buildBubbles, type Bubble } from "@/lib/renderItems";
import type { EmbedSessionClient } from "@/embed-session-api";
import { useEmbeddedSessionTimeline } from "./useEmbeddedSessionTimeline";

export function EmbeddedSessionTimeline({ client }: { client: EmbedSessionClient }) {
  const { state, sendMessage } = useEmbeddedSessionTimeline(client);
  const bubbles = useMemo(
    () => buildBubbles(state.blocks, state.activeResponse),
    [state.blocks, state.activeResponse],
  );
  const canWrite = client.sendMessage !== undefined;

  if (state.isLoading) {
    return <EmbeddedStatus label="Loading embedded session…" />;
  }
  if (state.error && state.session === null) {
    return <EmbeddedError error={state.error.message} />;
  }

  return (
    <div className="flex min-h-screen flex-col bg-background text-foreground">
      <header className="flex min-h-12 items-center justify-between border-b px-4 text-sm">
        <span className="min-w-0 truncate font-medium">{state.session?.title ?? "Agent session"}</span>
        <span className="ml-3 shrink-0 text-muted-foreground">{statusLabel(state.status)}</span>
      </header>
      <Conversation className="flex-1">
        <ConversationContent className="mx-auto w-full max-w-3xl gap-4 px-4 py-5">
          {bubbles.length === 0 ? (
            <ConversationEmptyState
              title="No messages yet"
              description="The agent session has not produced any messages."
              className="min-h-64"
            />
          ) : (
            bubbles.map((bubble) => <EmbeddedBubble key={bubbleKey(bubble)} bubble={bubble} status={state.status} />)
          )}
          {state.status === "running" && <WorkingRow />}
        </ConversationContent>
      </Conversation>
      {state.error && <div className="border-t px-4 py-2 text-sm text-destructive">{state.error.message}</div>}
      <EmbeddedComposer canWrite={canWrite} sending={state.isSending} onSend={sendMessage} />
    </div>
  );
}

function EmbeddedBubble({ bubble, status }: { bubble: Bubble; status: "idle" | "launching" | "running" | "waiting" | "failed" }) {
  if (bubble.kind === "user") {
    const text = bubble.content
      .filter((block) => block.type === "input_text")
      .map((block) => block.text)
      .join("");
    return (
      <Message from="user">
        <MessageContent>{text && <MessageResponse>{text}</MessageResponse>}</MessageContent>
      </Message>
    );
  }
  if (bubble.kind === "assistant") {
    const elicitations = bubble.items.filter((item) => item.kind === "elicitation");
    const renderable = bubble.items.filter((item) => item.kind !== "elicitation");
    return (
      <Message from="assistant" className="max-w-full">
        <MessageContent className="w-full">
          {renderable.length > 0 && <BlockRenderer items={renderable} sessionStatus={status} />}
          {elicitations.map((item) => (
            <div key={item.elicitationId} className="flex items-center gap-2 rounded border px-3 py-2 text-sm text-muted-foreground">
              <ShieldAlertIcon className="size-4 shrink-0" />
              <span>{item.message || "Agent approval is required in the host workspace."}</span>
            </div>
          ))}
        </MessageContent>
      </Message>
    );
  }
  if (bubble.kind === "compaction_loading") return <WorkingRow label="Compacting conversation…" />;
  if (bubble.kind === "compaction") return <WorkingRow label="Conversation compacted" />;
  return (
    <div className="text-center text-xs text-muted-foreground">
      Agent selected {bubble.model}.
    </div>
  );
}

function EmbeddedComposer({
  canWrite,
  sending,
  onSend,
}: {
  canWrite: boolean;
  sending: boolean;
  onSend: (text: string) => Promise<void>;
}) {
  const [value, setValue] = useState("");
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!value.trim() || sending || !canWrite) return;
    const text = value;
    setValue("");
    await onSend(text);
  };
  return (
    <form className="flex gap-2 border-t p-3" onSubmit={submit}>
      <textarea
        aria-label="Message the agent"
        className="min-h-11 flex-1 resize-none rounded border bg-background px-3 py-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed"
        disabled={!canWrite || sending}
        onChange={(event) => setValue(event.target.value)}
        placeholder={canWrite ? "Message the agent" : "Read-only session"}
        value={value}
      />
      <Button aria-label="Send message" disabled={!canWrite || sending || !value.trim()} size="icon" title="Send message" type="submit">
        {sending ? <Loader2Icon className="size-4 animate-spin" /> : <ArrowUpIcon className="size-4" />}
      </Button>
    </form>
  );
}

function EmbeddedStatus({ label }: { label: string }) {
  return <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">{label}</div>;
}

function EmbeddedError({ error }: { error: string }) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-2 px-6 text-center">
      <h1 className="font-medium">Embedded session unavailable</h1>
      <p className="text-sm text-muted-foreground">{error}</p>
    </div>
  );
}

function WorkingRow({ label = "Agent is working…" }: { label?: string }) {
  return <div className="flex items-center gap-2 text-sm text-muted-foreground"><Loader2Icon className="size-4 animate-spin" />{label}</div>;
}

function statusLabel(status: "idle" | "launching" | "running" | "waiting" | "failed"): string {
  if (status === "running" || status === "waiting") return "Working";
  if (status === "launching") return "Starting";
  if (status === "failed") return "Failed";
  return "Ready";
}

function bubbleKey(bubble: Bubble): string {
  if (bubble.kind === "user") return `user-${bubble.itemId}`;
  if (bubble.kind === "assistant") return `assistant-${bubble.stableId}`;
  return `${bubble.kind}-${bubble.itemId}`;
}
