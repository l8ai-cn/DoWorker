import { Loader2, Target, Zap } from "lucide-react";
import { useRef } from "react";
import {
  detectSlashToken,
  SlashMenu,
  SLASH_COMMANDS,
  type SlashCommand,
} from "@/components/slash-menu";
import type { LiveExpert } from "@/lib/experts-api";
import { cn } from "@/lib/utils";

interface TaskPromptProps {
  asGoal: boolean;
  authenticated: boolean;
  currentExpert: LiveExpert | undefined;
  currentWorkerAvailable: boolean;
  prompt: string;
  slashToken: { start: number; token: string } | null;
  submitting: boolean;
  error: string | null;
  onPromptChange: (value: string, caret: number) => void;
  onSlashTokenChange: (token: { start: number; token: string } | null) => void;
  onSlashPick: (command: SlashCommand, textarea: HTMLTextAreaElement) => void;
  onSubmit: () => void;
}

export function TaskPrompt({
  asGoal,
  authenticated,
  currentExpert,
  currentWorkerAvailable,
  prompt,
  slashToken,
  submitting,
  error,
  onPromptChange,
  onSlashTokenChange,
  onSlashPick,
  onSubmit,
}: TaskPromptProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const slashCommands = SLASH_COMMANDS.filter((command) => command.cmd !== "/attach");
  const placeholder = asGoal
    ? "描述你希望持续达成的目标，例如：保持主分支 CI 常绿..."
    : currentExpert
      ? `告诉 ${currentExpert.name} 要做什么... （输入 / 唤起命令）`
      : "想让 agent 做什么？（输入 / 唤起命令）";

  return (
    <div className="relative rounded-2xl bg-card p-3 ring-1 ring-border/50 focus-within:ring-primary/50">
      <textarea
        ref={textareaRef}
        rows={asGoal ? 4 : 3}
        value={prompt}
        onChange={(event) =>
          onPromptChange(
            event.target.value,
            event.target.selectionStart ?? event.target.value.length,
          )
        }
        onKeyUp={(event) => {
          const element = event.currentTarget;
          onSlashTokenChange(
            detectSlashToken(element.value, element.selectionStart ?? element.value.length),
          );
        }}
        onBlur={() => setTimeout(() => onSlashTokenChange(null), 150)}
        placeholder={placeholder}
        autoFocus
        className="w-full resize-none bg-transparent text-[14px] leading-relaxed outline-none placeholder:text-muted-foreground"
      />
      {slashToken && (
        <SlashMenu
          token={slashToken.token}
          commands={slashCommands}
          onPick={(command) => {
            if (textareaRef.current) onSlashPick(command, textareaRef.current);
          }}
          className="absolute left-3 right-3 top-full mt-1 w-auto"
        />
      )}
      <div className="mt-2 flex items-center gap-2 border-t border-border/40 pt-2">
        <button
          type="button"
          onClick={() => {
            const element = textareaRef.current;
            if (!element) return;
            const caret = element.selectionStart ?? prompt.length;
            onPromptChange(`${prompt.slice(0, caret)}/${prompt.slice(caret)}`, caret + 1);
            onSlashTokenChange({ start: caret, token: "/" });
            requestAnimationFrame(() => {
              element.focus();
              element.setSelectionRange(caret + 1, caret + 1);
            });
          }}
          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-surface font-mono text-[13px] text-muted-foreground ring-1 ring-border/40 hover:text-foreground hover:ring-primary/40"
          aria-label="斜杠命令"
        >
          /
        </button>
        <button
          type="button"
          onClick={onSubmit}
          disabled={submitting || !prompt.trim() || (authenticated && !currentWorkerAvailable)}
          className={cn(
            "flex h-9 min-w-0 flex-1 items-center justify-center gap-1.5 rounded-full bg-primary px-3 text-[13px] font-semibold text-primary-foreground transition active:scale-[0.98] disabled:opacity-40",
            !submitting && "glow-primary",
          )}
        >
          {submitting ? (
            <>
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
              创建中…
            </>
          ) : asGoal ? (
            <>
              <Target className="h-3.5 w-3.5" />
              保存目标
            </>
          ) : (
            <>
              <Zap className="h-3.5 w-3.5" />
              {authenticated ? "派发任务" : "登录后派发"}
            </>
          )}
        </button>
      </div>
      {error && <p className="mt-2 text-center text-xs text-destructive">{error}</p>}
    </div>
  );
}
