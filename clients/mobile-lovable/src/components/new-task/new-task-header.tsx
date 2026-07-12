import { Link } from "@tanstack/react-router";
import { ArrowLeft, Target } from "lucide-react";
import { cn } from "@/lib/utils";

interface NewTaskHeaderProps {
  asGoal: boolean;
  onGoalChange: (asGoal: boolean) => void;
}

export function NewTaskHeader({ asGoal, onGoalChange }: NewTaskHeaderProps) {
  return (
    <header className="safe-top sticky top-0 z-30 border-b border-border/60 bg-background/85 px-4 pb-3 pt-3 backdrop-blur-xl">
      <div className="flex items-center gap-2">
        <Link
          to="/"
          className="flex h-8 w-8 items-center justify-center rounded-full hover:bg-surface"
        >
          <ArrowLeft className="h-4 w-4" />
        </Link>
        <h1 className="flex-1 text-[14px] font-semibold">新任务</h1>
        <button
          type="button"
          onClick={() => onGoalChange(!asGoal)}
          className={cn(
            "flex items-center gap-1 rounded-full px-2.5 py-1 text-[11px] font-medium transition",
            asGoal
              ? "bg-primary text-primary-foreground glow-primary"
              : "bg-surface text-muted-foreground ring-1 ring-border/50 hover:text-foreground",
          )}
        >
          <Target className="h-3 w-3" />
          作为目标
        </button>
      </div>
    </header>
  );
}
