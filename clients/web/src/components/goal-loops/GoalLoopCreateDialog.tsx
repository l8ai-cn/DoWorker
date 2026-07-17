"use client";

import { Dialog, DialogClose, DialogContent } from "@/components/ui/dialog";
import { ResourceEditorShell } from "@/components/resource-editor/ResourceEditorShell";

interface GoalLoopCreateDialogProps {
  open: boolean;
  orgSlug: string;
  onApplied: () => void;
  onOpenChange: (open: boolean) => void;
}

export function GoalLoopCreateDialog({
  open,
  orgSlug,
  onApplied,
  onOpenChange,
}: GoalLoopCreateDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="relative max-w-5xl"
        title="创建目标 Loop"
        description="应用后创建草稿；启动仍由 Loop 列表中的显式操作触发。"
      >
        <DialogClose onClose={() => onOpenChange(false)} />
        {open && (
          <div className="px-6 pb-6">
            <ResourceEditorShell
              orgSlug={orgSlug}
              kind="GoalLoop"
              onApplied={() => {
                onOpenChange(false);
                onApplied();
              }}
            />
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
