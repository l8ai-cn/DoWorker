import { ChevronDownIcon, FolderIcon } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { WorkspacePicker, isNavigablePath } from "./WorkspacePicker";

export function NewChatWorkspaceChip({
  open,
  onOpenChange,
  label,
  workspace,
  hostId,
  onWorkspaceChange,
  occupancyForPath,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  label: string;
  workspace: string;
  hostId: string | null;
  onWorkspaceChange: (value: string) => void;
  occupancyForPath?: (path: string) => number;
}) {
  return (
    <Popover open={open} onOpenChange={onOpenChange}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="flex h-6 items-center gap-1 rounded-full px-2.5 text-13 font-normal text-muted-foreground transition-colors hover:text-foreground"
          data-testid="new-chat-landing-workspace-chip"
        >
          <FolderIcon className="size-4 shrink-0" />
          <span className={`hidden max-w-40 truncate sm:block ${workspace !== "" ? "text-foreground" : ""}`}>{label}</span>
          <ChevronDownIcon className="size-3.5 shrink-0 opacity-60" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-[min(420px,calc(100vw-2rem))] p-0">
        {hostId ? (
          <WorkspacePicker
            hostId={hostId}
            initialPath={isNavigablePath(workspace) ? workspace : undefined}
            onNavigate={onWorkspaceChange}
            occupancyForPath={occupancyForPath}
          />
        ) : (
          <p className="p-3 text-xs text-muted-foreground">Select a host first.</p>
        )}
      </PopoverContent>
    </Popover>
  );
}
