import { Loader2 } from "lucide-react";

interface AgentSessionLinkStateProps {
  locale: string;
}

export function AgentSessionLinkState({
  locale,
}: AgentSessionLinkStateProps) {
  return (
    <div className="flex flex-1 items-center justify-center bg-background">
      <div className="flex items-center gap-3 text-sm text-muted-foreground">
        <Loader2 className="size-5 animate-spin text-primary" />
        <span>
          {locale === "zh"
            ? "正在连接 Agent 会话..."
            : "Connecting to the Agent session..."}
        </span>
      </div>
    </div>
  );
}
