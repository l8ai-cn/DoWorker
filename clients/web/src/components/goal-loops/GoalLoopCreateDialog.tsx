"use client";

import { Dialog, DialogClose, DialogContent } from "@/components/ui/dialog";
import { GoalLoopCreateForm } from "./GoalLoopCreateForm";
import type { GoalLoopData, GoalLoopWorkerSnapshot } from "@/lib/viewModels/goal-loop";

interface GoalLoopCreateDialogProps {
  open: boolean;
  orgSlug: string;
  workerSnapshots: GoalLoopWorkerSnapshot[];
  onCreated: (loop: GoalLoopData) => void;
  onOpenChange: (open: boolean) => void;
}

export function GoalLoopCreateDialog({
  open,
  orgSlug,
  workerSnapshots,
  onCreated,
  onOpenChange,
}: GoalLoopCreateDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="relative max-w-3xl"
        title="创建目标 Loop"
        description="Loop 只执行一次，完成必须由验证命令退出码为 0 判定。"
      >
        <DialogClose onClose={() => onOpenChange(false)} />
        {open && (
          <GoalLoopCreateForm
            orgSlug={orgSlug}
            workerSnapshots={workerSnapshots}
            onCreated={(loop) => {
              onCreated(loop);
              onOpenChange(false);
            }}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}
