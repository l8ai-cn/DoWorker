"use client";

import { useState, useEffect } from "react";

interface PodTerminal {
  name: string;
  agent: string;
  workspace: string;
  lines: string[];
  color: string;
}

const PODS: PodTerminal[] = [
  {
    name: "pod-alpha",
    agent: "Claude Code",
    workspace: "/projects/api",
    color: "text-primary",
    lines: [
      "$ claude --resume",
      "> Resuming session...",
      "> Writing src/auth/handler.ts",
      "> Writing src/auth/oauth.ts",
      "$ go test ./internal/auth/...",
      "ok  internal/auth  0.34s",
      "> Creating merge request...",
      "> ✓ MR !41 created",
    ],
  },
  {
    name: "pod-beta",
    agent: "Codex CLI",
    workspace: "/projects/web",
    color: "text-info",
    lines: [
      "$ codex start",
      "> Analyzing codebase...",
      "> Writing src/components/Auth.tsx",
      "> Writing src/hooks/useAuth.ts",
      "$ pnpm test --run",
      "✓ 8 tests passed",
      "> Pushing to feature/auth-ui",
      "> ✓ Branch pushed",
    ],
  },
  {
    name: "pod-gamma",
    agent: "Aider",
    workspace: "/projects/mobile",
    color: "text-accent-foreground",
    lines: [
      "$ aider --model opus",
      "> Loading repo map...",
      "> Editing lib/auth/login.dart",
      "> Editing lib/auth/token.dart",
      "$ flutter test",
      "All 14 tests passed!",
      "> Committing changes...",
      "> ✓ 2 files changed",
    ],
  },
];

function Terminal({
  pod,
  displayedLines,
}: {
  pod: PodTerminal;
  displayedLines: number;
}) {
  return (
    <div className="surface-card rounded-lg overflow-hidden shadow-xl">
      <div className="flex items-center justify-between px-3 py-2 bg-surface-muted/50 panel-lift">
        <div className="flex items-center gap-1.5">
          <div className="w-2.5 h-2.5 rounded-full bg-danger" />
          <div className="w-2.5 h-2.5 rounded-full bg-warning" />
          <div className="w-2.5 h-2.5 rounded-full bg-success" />
        </div>
        <span className="text-[10px] font-mono text-muted-foreground">{pod.name}</span>
        <div className="flex items-center gap-1">
          <span className="w-1.5 h-1.5 rounded-full bg-success animate-pulse" />
        </div>
      </div>

      <div className="px-3 py-1 bg-background panel-lift text-[10px] font-mono text-muted-foreground flex gap-3">
        <span className={pod.color}>{pod.agent}</span>
        <span className="text-muted-foreground/50">|</span>
        <span>{pod.workspace}</span>
      </div>

      <div className="p-3 font-mono text-[11px] leading-[1.6] h-[140px] overflow-hidden">
        {pod.lines.slice(0, displayedLines).map((line, i) => (
          <div
            key={i}
            className={
              line.startsWith("$")
                ? "text-info"
                : line.startsWith(">")
                  ? line.includes("✓")
                    ? "text-success"
                    : "text-primary"
                  : line.includes("passed") || line.includes("ok ")
                    ? "text-success"
                    : "text-foreground"
            }
          >
            {line}
          </div>
        ))}
        {displayedLines < pod.lines.length && (
          <span className="text-info animate-pulse">▋</span>
        )}
      </div>
    </div>
  );
}

export function AgentPodDemo() {
  const [displayedLines, setDisplayedLines] = useState(0);
  const maxLines = Math.max(...PODS.map((p) => p.lines.length));

  useEffect(() => {
    if (displayedLines < maxLines) {
      const timer = setTimeout(() => {
        setDisplayedLines((prev) => prev + 1);
      }, 600);
      return () => clearTimeout(timer);
    } else {
      const timer = setTimeout(() => {
        setDisplayedLines(0);
      }, 3000);
      return () => clearTimeout(timer);
    }
  }, [displayedLines, maxLines]);

  return (
    <div className="space-y-2.5">
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
        <Terminal pod={PODS[0]} displayedLines={displayedLines} />
        <Terminal pod={PODS[1]} displayedLines={displayedLines} />
      </div>
      <Terminal pod={PODS[2]} displayedLines={displayedLines} />
    </div>
  );
}
