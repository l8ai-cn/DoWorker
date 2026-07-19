import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

export function RestartWithModelDialog({
  sessionId: _sessionId,
  currentModel: _currentModel,
  open,
  onOpenChange,
}: {
  sessionId: string;
  currentModel?: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent data-testid="restart-model-unavailable-dialog" className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>模型变更不可用</DialogTitle>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">
          当前协议不允许在同一 Agent 的会话副本中单独更换模型资源。请创建新的 Worker 会话。
        </p>
      </DialogContent>
    </Dialog>
  );
}
