import { MessageSquare, Terminal } from "lucide-react";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import type { SessionInteractionMode } from "@/lib/sessions-api";

export function InteractionModeSelector({
  mode,
  supportedModes,
  disabled,
  onChange,
}: {
  mode: SessionInteractionMode;
  supportedModes: SessionInteractionMode[];
  disabled?: boolean;
  onChange: (mode: SessionInteractionMode) => void;
}) {
  return (
    <div>
      <div className="mb-1.5 flex items-center justify-between px-1">
        <p className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
          交互方式
        </p>
        <span className="text-[10px] text-muted-foreground/70">创建后固定</span>
      </div>
      <div className="grid grid-cols-2 gap-2">
        <ModeButton
          active={mode === "acp"}
          disabled={disabled || !supportedModes.includes("acp")}
          icon={<MessageSquare className="h-4 w-4" />}
          label="可视化对话"
          detail="计划、工具和授权"
          onClick={() => onChange("acp")}
        />
        <ModeButton
          active={mode === "pty"}
          disabled={disabled || !supportedModes.includes("pty")}
          icon={<Terminal className="h-4 w-4" />}
          label="命令行"
          detail="原生 CLI 终端"
          onClick={() => onChange("pty")}
        />
      </div>
    </div>
  );
}

function ModeButton({
  active,
  disabled,
  icon,
  label,
  detail,
  onClick,
}: {
  active: boolean;
  disabled?: boolean;
  icon: ReactNode;
  label: string;
  detail: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={cn(
        "flex min-h-16 items-center gap-2 rounded-md border px-3 text-left transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50",
        active ? "border-primary bg-primary/10" : "border-border bg-card hover:border-primary/50",
      )}
    >
      <span className={cn("shrink-0", active ? "text-primary" : "text-muted-foreground")}>
        {icon}
      </span>
      <span className="min-w-0">
        <span className="block text-xs font-semibold">{label}</span>
        <span className="block truncate text-[10px] text-muted-foreground">{detail}</span>
      </span>
    </button>
  );
}
