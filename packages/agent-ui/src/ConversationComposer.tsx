import { ArrowUp, LoaderCircle, Square } from "lucide-react";
import {
  useState,
  type FormEvent,
  type KeyboardEvent,
} from "react";

import { ComposerCapabilityBar } from "./ComposerCapabilityBar";
import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import {
  commandQuery,
  ComposerCommandMenu,
  parseAgentCommand,
} from "./ComposerCommandMenu";
import type {
  AgentSessionRuntime,
  AgentSessionSnapshot,
} from "./contracts";

export function ConversationComposer({
  onError,
  runtime,
  snapshot,
}: {
  onError: (error: unknown) => void;
  runtime: AgentSessionRuntime;
  snapshot: AgentSessionSnapshot;
}) {
  const [value, setValue] = useState("");
  const [sending, setSending] = useState(false);
  const text = useAgentWorkspaceText();
  const isRunning = snapshot.status === "running" || snapshot.status === "waiting";
  const hasDraft = value.trim().length > 0;
  const showInterrupt = isRunning && snapshot.capabilities.interrupt;

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    const message = value.trim();
    if (!message || sending || isRunning || !snapshot.capabilities.sendMessage) return;
    const parsedCommand = parseAgentCommand(message, snapshot.commands ?? []);
    if (parsedCommand?.command.requiresArgument && !parsedCommand.arguments) {
      onError(new Error(text.requiresArgument(parsedCommand.command.label)));
      return;
    }
    setSending(true);
    try {
      if (parsedCommand && runtime.sendSlashCommand) {
        await runtime.sendSlashCommand(
          snapshot.sessionId,
          crypto.randomUUID(),
          {
            name: parsedCommand.command.name,
            arguments: parsedCommand.arguments,
          },
        );
      } else {
        await runtime.sendMessage(snapshot.sessionId, crypto.randomUUID(), {
          text: message,
        });
      }
      setValue("");
    } catch (cause) {
      onError(cause);
    } finally {
      setSending(false);
    }
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (
      event.key !== "Enter" ||
      event.shiftKey ||
      event.nativeEvent.isComposing
    ) {
      return;
    }
    event.preventDefault();
    event.currentTarget.form?.requestSubmit();
  };

  return (
    <form className="shrink-0 px-3 pb-3 pt-2" onSubmit={submit}>
      <div className="relative mx-auto w-full max-w-4xl rounded-lg border border-border bg-card shadow-sm transition-colors focus-within:border-ring">
        <ComposerCommandMenu
          commands={snapshot.commands ?? []}
          onSelect={(command) =>
            setValue(`${command.label}${command.requiresArgument ? " " : ""}`)
          }
          query={commandQuery(value)}
        />
        <textarea
          aria-label={text.messageAgent}
          className="min-h-24 max-h-56 w-full resize-none bg-transparent px-4 pb-2 pt-3 text-sm leading-6 outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-60"
          disabled={!snapshot.capabilities.sendMessage || sending}
          onChange={(event) => setValue(event.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={
            snapshot.capabilities.sendMessage
              ? text.composerPlaceholder(snapshot.agentLabel)
              : text.readOnly
          }
          rows={3}
          value={value}
        />
        <div className="flex min-h-11 items-end justify-between gap-2 px-2 pb-2">
          <ComposerCapabilityBar
            onError={onError}
            runtime={runtime}
            snapshot={snapshot}
          />
          {showInterrupt ? (
            <button
              aria-label={text.stopAgent}
              className="flex size-10 shrink-0 items-center justify-center rounded-full bg-destructive text-destructive-foreground outline-none hover:opacity-90 focus-visible:ring-2 focus-visible:ring-ring"
              onClick={() =>
                void runtime
                  .interrupt(snapshot.sessionId, crypto.randomUUID())
                  .catch(onError)
              }
              title={text.stopAgent}
              type="button"
            >
              <Square className="size-3.5 fill-current" />
            </button>
          ) : (
            <button
              aria-label={text.sendMessage}
              className="flex size-10 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground outline-none hover:opacity-90 focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-30"
              disabled={
                !hasDraft ||
                sending ||
                isRunning ||
                !snapshot.capabilities.sendMessage
              }
              title={text.sendMessage}
              type="submit"
            >
              {sending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : (
                <ArrowUp className="size-4" />
              )}
            </button>
          )}
        </div>
      </div>
    </form>
  );
}
