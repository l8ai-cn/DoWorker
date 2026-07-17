import { Link } from "@tanstack/react-router";
import { ArrowLeft, MessageSquare, MonitorPlay, Terminal } from "lucide-react";

type WorkerConsoleMode = "acp" | "pty";

export function MobileWorkerConsoleHeader({
  mode,
  podKey,
  previewAvailable,
}: {
  mode: WorkerConsoleMode;
  podKey: string;
  previewAvailable: boolean;
}) {
  const isAcp = mode === "acp";
  const Icon = isAcp ? MessageSquare : Terminal;
  const title = isAcp ? "对话" : "命令行";

  return (
    <header className="safe-top flex shrink-0 items-center gap-2 border-b border-border/60 bg-background/90 px-3 pb-2 pt-2 backdrop-blur-xl">
      <Link
        to="/"
        className="flex h-9 w-9 items-center justify-center rounded-md hover:bg-surface"
        aria-label="返回会话列表"
      >
        <ArrowLeft className="h-4 w-4" />
      </Link>
      <div className="min-w-0 flex-1">
        <p className="flex items-center gap-1.5 text-sm font-semibold">
          <Icon className="h-4 w-4 text-primary" /> {title}
        </p>
        <p className="truncate font-mono text-[10px] text-muted-foreground">{podKey}</p>
      </div>
      {previewAvailable && (
        <Link
          to="/workers/$podKey/preview"
          params={{ podKey }}
          aria-label="打开预览"
          title="打开预览"
          className="flex h-11 w-11 items-center justify-center rounded-md text-muted-foreground hover:bg-surface hover:text-foreground"
        >
          <MonitorPlay className="h-4 w-4" />
        </Link>
      )}
    </header>
  );
}
