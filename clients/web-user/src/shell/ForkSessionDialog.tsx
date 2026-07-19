import { InfoIcon } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { ForkSessionForm } from "./ForkSessionForm";

export { ForkSessionForm } from "./ForkSessionForm";

export function ForkSessionDialog({
  sourceSessionId,
  sourceTitle,
  sourceWorkspace,
  sourceHostId,
  sourceGitBranch,
  upToResponseId,
  open,
  onOpenChange,
}: {
  sourceSessionId: string;
  sourceTitle?: string | null;
  sourceWorkspace?: string | null;
  sourceHostId?: string | null;
  sourceGitBranch?: string | null;
  upToResponseId?: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const truncated = upToResponseId != null;
  const description = truncated
    ? "Copies history through the selected response. The source remains unchanged."
    : "Copies the session history. The source remains unchanged.";
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent data-testid="fork-session-dialog" className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-1.5">
            {truncated ? "Fork from this response" : "Clone session"}
            <Tooltip>
              <TooltipTrigger asChild>
                <button type="button" aria-label="What does cloning do?" className="text-muted-foreground">
                  <InfoIcon className="size-4" />
                </button>
              </TooltipTrigger>
              <TooltipContent>{description}</TooltipContent>
            </Tooltip>
          </DialogTitle>
          <DialogDescription className="sr-only">{description}</DialogDescription>
        </DialogHeader>
        <ForkSessionForm
          sourceSessionId={sourceSessionId}
          sourceTitle={sourceTitle}
          sourceWorkspace={sourceWorkspace}
          sourceHostId={sourceHostId}
          sourceGitBranch={sourceGitBranch}
          upToResponseId={upToResponseId}
          onClose={() => onOpenChange(false)}
        />
      </DialogContent>
    </Dialog>
  );
}
