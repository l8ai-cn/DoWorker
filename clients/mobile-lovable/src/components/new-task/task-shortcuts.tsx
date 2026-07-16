import { SLASH_COMMANDS } from "@/components/slash-menu";
import type { LiveExpert } from "@/lib/experts-api";

const taskTemplates = [
  "修一下 CI 里最新失败的那个 test",
  "review 我最近的 PR 并给出改进建议",
  "把 README 翻译成日语",
];

const goalTemplates = ["保持主分支 CI 常绿", "持续处理依赖安全告警", "维护 README 与 API 文档同步"];

const taskCommands = SLASH_COMMANDS.filter((command) => command.cmd !== "/attach");

interface TaskShortcutsProps {
  asGoal: boolean;
  expert: LiveExpert | undefined;
  prompt: string;
  onPromptChange: (value: string, caret: number) => void;
}

export function TaskShortcuts({ asGoal, expert, prompt, onPromptChange }: TaskShortcutsProps) {
  const templates = asGoal
    ? goalTemplates
    : expert?.prompt
      ? [expert.prompt.slice(0, 60)]
      : taskTemplates;
  return (
    <>
      <div className="flex flex-wrap gap-1.5">
        {templates.map((template) => (
          <button
            key={template}
            type="button"
            onClick={() => onPromptChange(template, template.length)}
            className="rounded-full bg-surface px-2.5 py-1 text-[11px] text-foreground/80 ring-1 ring-border/40 hover:ring-primary/40"
          >
            {template}
          </button>
        ))}
      </div>
      {!asGoal && (
        <div className="flex flex-wrap gap-1">
          {taskCommands.slice(0, 6).map((command) => (
            <button
              key={command.cmd}
              type="button"
              onClick={() => {
                const next = `${prompt}${prompt ? " " : ""}${command.cmd} `;
                onPromptChange(next, next.length);
              }}
              className="flex items-center gap-1 rounded-full bg-surface/60 px-2 py-0.5 font-mono text-[10.5px] text-muted-foreground ring-1 ring-border/40 hover:text-primary hover:ring-primary/40"
            >
              {command.emoji} {command.cmd}
            </button>
          ))}
        </div>
      )}
    </>
  );
}
