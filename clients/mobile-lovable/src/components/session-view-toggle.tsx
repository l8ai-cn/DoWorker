import { Link } from "@tanstack/react-router";
import { MessageSquare, Terminal as TerminalIcon } from "lucide-react";
import { cn } from "@/lib/utils";

export function SessionViewToggle({
  sessionId,
  mode,
  interactionMode,
}: {
  sessionId: string;
  mode: "chat" | "terminal";
  interactionMode?: "acp" | "pty" | null;
}) {
  const chatUnavailable = interactionMode === "pty";
  const terminalUnavailable = interactionMode === "acp";
  return (
    <div className="flex shrink-0 rounded-full bg-surface p-0.5 ring-1 ring-border/60">
      {chatUnavailable ? (
        <span className="flex items-center gap-1 rounded-full px-2 py-1 text-[10px] text-muted-foreground/50">
          <MessageSquare className="h-3 w-3" /> 聊天
        </span>
      ) : (
        <Link
          to="/sessions/$sessionId"
          params={{ sessionId }}
          className={cn(
            "flex items-center gap-1 rounded-full px-2 py-1 text-[10px] font-medium transition",
            mode === "chat" ? "bg-primary text-primary-foreground shadow-sm" : "text-muted-foreground",
          )}
        >
          <MessageSquare className="h-3 w-3" /> 聊天
        </Link>
      )}
      {terminalUnavailable ? (
        <span className="flex items-center gap-1 rounded-full px-2 py-1 text-[10px] text-muted-foreground/50">
          <TerminalIcon className="h-3 w-3" /> 终端
        </span>
      ) : (
        <Link
          to="/sessions/$sessionId/terminal"
          params={{ sessionId }}
          className={cn(
            "flex items-center gap-1 rounded-full px-2 py-1 text-[10px] font-medium transition",
            mode === "terminal" ? "bg-primary text-primary-foreground shadow-sm" : "text-muted-foreground",
          )}
        >
          <TerminalIcon className="h-3 w-3" /> 终端
        </Link>
      )}
    </div>
  );
}
