import { cn } from "@/lib/utils";

export interface SlashCommand {
  cmd: string;         // e.g. "/plan"
  label: string;       // 中文简介
  hint?: string;       // usage hint
  emoji: string;
}

export const SLASH_COMMANDS: SlashCommand[] = [
  { cmd: "/plan", label: "生成执行计划", hint: "让 agent 先出 plan 再动手", emoji: "📋" },
  { cmd: "/think", label: "深度思考", hint: "开启扩展思考预算", emoji: "🧠" },
  { cmd: "/web", label: "联网搜索", hint: "/web <query>", emoji: "🌐" },
  { cmd: "/diff", label: "查看当前 diff", hint: "汇总本次改动", emoji: "🔀" },
  { cmd: "/approve", label: "批准工具调用", hint: "批准最近的待审批", emoji: "✅" },
  { cmd: "/reject", label: "拒绝工具调用", emoji: "🛑" },
  { cmd: "/model", label: "切换模型", hint: "/model gpt-5 | claude-4 ...", emoji: "🎛️" },
  { cmd: "/expert", label: "切换专家", hint: "/expert email-butler", emoji: "✨" },
  { cmd: "/goal", label: "转成常驻目标", hint: "把当前描述保存为 Goal", emoji: "🎯" },
  { cmd: "/attach", label: "添加附件/图片", emoji: "📎" },
  { cmd: "/clear", label: "清空当前上下文", emoji: "🧹" },
  { cmd: "/compact", label: "压缩对话历史", hint: "节省 token", emoji: "🗜️" },
  { cmd: "/stop", label: "停止执行", emoji: "⏹️" },
  { cmd: "/help", label: "查看全部命令", emoji: "❔" },
];

/**
 * Given input value + caret pos, return the current slash token (starting at "/")
 * or null if not in one. Only triggers when "/" is at start or after whitespace.
 */
export function detectSlashToken(value: string, caret: number): { start: number; token: string } | null {
  if (caret <= 0) return null;
  // walk back from caret to find "/" or whitespace
  let i = caret - 1;
  while (i >= 0 && !/\s/.test(value[i])) {
    if (value[i] === "/") {
      // must be at start or preceded by whitespace
      if (i === 0 || /\s/.test(value[i - 1])) {
        return { start: i, token: value.slice(i, caret) };
      }
      return null;
    }
    i--;
  }
  return null;
}

interface SlashMenuProps {
  token: string;
  onPick: (cmd: SlashCommand) => void;
  className?: string;
}

export function SlashMenu({ token, onPick, className }: SlashMenuProps) {
  const q = token.slice(1).toLowerCase();
  const items = SLASH_COMMANDS.filter(
    (c) => !q || c.cmd.slice(1).startsWith(q) || c.label.toLowerCase().includes(q),
  ).slice(0, 7);
  if (items.length === 0) return null;
  return (
    <div className={cn(
      "z-30 max-h-[240px] overflow-y-auto rounded-xl border border-border/60 bg-card p-1 shadow-lg ring-1 ring-black/5 stream-in",
      className,
    )}>
      <p className="px-2 pb-1 pt-1 text-[9.5px] font-semibold uppercase tracking-wider text-muted-foreground">
        斜杠命令 · {items.length}
      </p>
      {items.map((c) => (
        <button
          key={c.cmd}
          type="button"
          onMouseDown={(e) => { e.preventDefault(); onPick(c); }}
          className="flex w-full items-start gap-2 rounded-lg px-2 py-1.5 text-left hover:bg-surface"
        >
          <span className="text-base leading-none">{c.emoji}</span>
          <div className="min-w-0 flex-1">
            <p className="flex items-center gap-1.5 text-[12px]">
              <span className="font-mono font-semibold text-primary">{c.cmd}</span>
              <span className="truncate text-foreground/85">{c.label}</span>
            </p>
            {c.hint && <p className="truncate text-[10.5px] text-muted-foreground">{c.hint}</p>}
          </div>
        </button>
      ))}
    </div>
  );
}
