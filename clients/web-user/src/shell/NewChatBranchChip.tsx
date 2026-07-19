import { ChevronDownIcon, GitBranchIcon } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

export function NewChatBranchChip({
  branchName,
  baseBranch,
  label,
  onBranchNameChange,
  onBaseBranchChange,
}: {
  branchName: string;
  baseBranch: string;
  label: string;
  onBranchNameChange: (value: string) => void;
  onBaseBranchChange: (value: string) => void;
}) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="flex h-6 items-center gap-1 rounded-full px-2.5 text-13 font-normal text-muted-foreground transition-colors hover:text-foreground"
          data-testid="new-chat-landing-branch-chip"
        >
          <GitBranchIcon className="size-4 shrink-0" />
          <span className={`hidden max-w-32 truncate sm:block ${branchName.trim() ? "text-foreground" : ""}`}>{label}</span>
          <ChevronDownIcon className="size-3.5 shrink-0 opacity-60" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-[min(20rem,calc(100vw-2rem))] p-3">
        <div className="flex flex-col gap-2">
          <label htmlFor="landing-branch-name" className="text-xs font-medium text-foreground">
            Git worktree branch (optional)
          </label>
          <Input
            id="landing-branch-name"
            value={branchName}
            onChange={(event) => onBranchNameChange(event.target.value)}
            placeholder="feature/my-branch"
            className="text-xs"
            data-testid="new-chat-landing-branch-input"
          />
          {branchName.trim() !== "" && (
            <Input
              value={baseBranch}
              onChange={(event) => onBaseBranchChange(event.target.value)}
              placeholder="Base branch (defaults to current branch)"
              aria-label="Base branch"
              className="text-xs"
              data-testid="new-chat-landing-base-branch-input"
            />
          )}
          <p className="text-xs text-muted-foreground">
            Creates an isolated git worktree for a new branch. Leave blank to start directly in the working directory.
          </p>
        </div>
      </PopoverContent>
    </Popover>
  );
}
