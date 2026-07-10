"use client";

import { useCallback, useMemo, useState } from "react";
import { Send, StopCircle } from "lucide-react";
import { useTranslations } from "next-intl";
import { WorkerSlashDropdown } from "@/components/shared/WorkerSlashDropdown";
import { useWorkerSlashComposer } from "@/hooks/useWorkerSlashComposer";
import { relayPool } from "@/stores/relayConnection";
import { useAcpSessionField } from "@/stores/acpSession";
import { AcpPermissionModeSelector } from "./AcpPermissionModeSelector";

interface AcpPromptInputProps {
  podKey: string;
}

export function AcpPromptInput({ podKey }: AcpPromptInputProps) {
  const tRoot = useTranslations();
  const tPrompt = (key: string) => tRoot(`acp.promptInput.${key}`);
  const [prompt, setPrompt] = useState("");
  const [sending, setSending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const sessionState = useAcpSessionField(podKey, (s) => s.state);
  const isProcessing = sessionState === "processing" || sessionState === "waiting_permission";

  const slash = useWorkerSlashComposer(tRoot);
  const matches = useMemo(
    () => slash.matchesFor(prompt, prompt.length),
    [slash, prompt],
  );

  const submitPrompt = useCallback(
    (raw: string) => {
      const resolved = slash.resolveSubmit(raw);
      if (!resolved) return;
      if (!relayPool.isConnected(podKey)) {
        setError(tPrompt("notConnected"));
        return;
      }
      setSending(true);
      setError(null);
      try {
        relayPool.sendAcpCommand(podKey, { type: "prompt", prompt: resolved.prompt });
        setPrompt("");
        slash.setVisible(false);
      } finally {
        setSending(false);
      }
    },
    [podKey, slash, tPrompt],
  );

  const handleSend = useCallback(() => {
    if (!prompt.trim() || sending || isProcessing) return;
    submitPrompt(prompt);
  }, [prompt, sending, isProcessing, submitPrompt]);

  const handleCancel = useCallback(() => {
    if (!relayPool.isConnected(podKey)) {
      setError(tPrompt("notConnected"));
      return;
    }
    setError(null);
    relayPool.sendAcpCommand(podKey, { type: "interrupt" });
  }, [podKey, tPrompt]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    slash.handleKeyDown(
      e,
      prompt,
      matches,
      (next) => {
        setPrompt(next);
        setError(null);
      },
      handleSend,
    );
  };

  return (
    <div className="border-t px-3 py-2">
      {error && (
        <div className="text-xs text-danger mb-1">{error}</div>
      )}
      <div className="relative flex items-center gap-2">
        <WorkerSlashDropdown
          commands={matches}
          activeIndex={slash.active}
          visible={slash.visible}
          onSelect={(command) => {
            setPrompt(slash.applySelection(command, prompt));
            slash.setVisible(false);
            setError(null);
          }}
        />
        <AcpPermissionModeSelector podKey={podKey} />
        <textarea
          value={prompt}
          onChange={(e) => {
            setPrompt(e.target.value);
            slash.syncMenu(e.target.value, e.target.selectionStart ?? e.target.value.length);
            setError(null);
          }}
          onKeyDown={handleKeyDown}
          placeholder={tPrompt("placeholder")}
          disabled={sending}
          className="flex-1 resize-none rounded-md border bg-background px-3 py-1.5 text-sm min-h-[36px] max-h-[120px] leading-[20px]"
          rows={1}
        />
        {isProcessing ? (
          <button
            onClick={handleCancel}
            className="shrink-0 rounded-md bg-destructive h-[36px] w-[36px] flex items-center justify-center text-destructive-foreground hover:bg-destructive/90"
            title={tPrompt("cancel")}
          >
            <StopCircle className="h-4 w-4" />
          </button>
        ) : (
          <button
            onClick={handleSend}
            disabled={sending || !prompt.trim()}
            className="shrink-0 rounded-md bg-primary h-[36px] w-[36px] flex items-center justify-center text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            <Send className="h-4 w-4" />
          </button>
        )}
      </div>
    </div>
  );
}
