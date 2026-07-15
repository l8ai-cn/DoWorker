import Link from "next/link";
import { ArrowLeft, Blocks, Braces, Play, RefreshCw } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { LoopEditor } from "@/lib/viewModels/loop-program";
import { loopParseStatusLabel } from "./loop-display-text";

interface LoopWorkbenchToolbarProps {
  orgSlug: string;
  editor: LoopEditor;
  parseStatus: string;
  running: boolean;
  onEditorChange: (editor: LoopEditor) => void;
  onRun: () => void;
}

export function LoopWorkbenchToolbar({
  orgSlug,
  editor,
  parseStatus,
  running,
  onEditorChange,
  onRun,
}: LoopWorkbenchToolbarProps) {
  const valid = parseStatus === "valid";
  return (
    <header className="grid shrink-0 grid-cols-[auto_minmax(0,1fr)] items-center gap-2 border-b border-border bg-surface-raised px-3 py-2 sm:flex sm:min-h-16 sm:gap-3 sm:px-4 sm:py-3">
      <Button asChild size="icon" variant="ghost">
        <Link aria-label="返回循环列表" href={`/${orgSlug}/loops`}>
          <ArrowLeft className="h-4 w-4" />
        </Link>
      </Button>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <Blocks className="h-4 w-4 text-primary" />
          <h1 className="truncate text-sm font-semibold">循环工作台</h1>
          <Badge variant={valid ? "success" : "warning"}>
            {loopParseStatusLabel(parseStatus)}
          </Badge>
        </div>
        <p className="mt-0.5 hidden text-xs text-muted-foreground sm:block">
          积木与循环脚本共享同一份后端确认语义
        </p>
      </div>
      <div className="col-span-2 flex min-w-0 items-center gap-2 sm:ml-auto">
        <div className="flex min-w-0 flex-1 rounded-md border border-border bg-surface p-0.5 sm:flex-none" role="tablist">
          <button
            aria-selected={editor === "blocks"}
            className={`flex h-8 flex-1 items-center justify-center gap-1.5 rounded px-3 text-xs font-medium sm:flex-none ${editor === "blocks" ? "bg-surface-raised text-foreground shadow-sm" : "text-muted-foreground"}`}
            onClick={() => onEditorChange("blocks")}
            role="tab"
            type="button"
          >
            <Blocks className="h-3.5 w-3.5" />积木
          </button>
          <button
            aria-selected={editor === "code"}
            className={`flex h-8 flex-1 items-center justify-center gap-1.5 rounded px-3 text-xs font-medium sm:flex-none ${editor === "code" ? "bg-surface-raised text-foreground shadow-sm" : "text-muted-foreground"}`}
            onClick={() => onEditorChange("code")}
            role="tab"
            type="button"
          >
            <Braces className="h-3.5 w-3.5" />代码
          </button>
        </div>
        <Button className="shrink-0" disabled={!valid || running} loading={running} onClick={onRun}>
          {running ? <RefreshCw className="mr-1.5 h-4 w-4 animate-spin" /> : <Play className="mr-1.5 h-4 w-4" />}
          运行循环
        </Button>
      </div>
    </header>
  );
}
