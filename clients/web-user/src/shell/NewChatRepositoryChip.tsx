import { CircleHelpIcon, ChevronDownIcon, GitBranchIcon } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

export function NewChatRepositoryChip({
  label,
  hasRepo,
  repoUrl,
  repoBranch,
  gitTooltip,
  onRepoUrlChange,
  onRepoBranchChange,
}: {
  label: string;
  hasRepo: boolean;
  repoUrl: string;
  repoBranch: string;
  gitTooltip?: string;
  onRepoUrlChange: (value: string) => void;
  onRepoBranchChange: (value: string) => void;
}) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="flex h-6 items-center gap-1 rounded-full px-2.5 text-13 font-normal text-muted-foreground transition-colors hover:text-foreground"
          data-testid="new-chat-landing-repo-chip"
        >
          <GitBranchIcon className="size-4 shrink-0" />
          <span className={`hidden max-w-40 truncate sm:block ${hasRepo ? "text-foreground" : "text-muted-foreground"}`}>{label}</span>
          <ChevronDownIcon className="size-3.5 shrink-0 opacity-60" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-96 p-3">
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-1.5">
            <label htmlFor="landing-repo-url" className="text-xs font-medium text-foreground">
              Repository (optional)
            </label>
            {gitTooltip && (
              <Tooltip>
                <TooltipTrigger asChild>
                  <button type="button" className="inline-flex size-4 items-center justify-center rounded-sm text-muted-foreground transition-colors hover:text-foreground" aria-label="How to set up Databricks git credentials">
                    <CircleHelpIcon className="size-3.5" />
                  </button>
                </TooltipTrigger>
                <TooltipContent className="max-w-64">{gitTooltip}</TooltipContent>
              </Tooltip>
            )}
          </div>
          <Input
            id="landing-repo-url"
            value={repoUrl}
            onChange={(event) => onRepoUrlChange(event.target.value)}
            placeholder="https://github.com/org/repo"
            className="text-xs"
            data-testid="new-chat-landing-repo-input"
          />
          <Input
            value={repoBranch}
            onChange={(event) => onRepoBranchChange(event.target.value)}
            placeholder="Branch (defaults to the repo's default)"
            aria-label="Repository branch"
            className="text-xs"
            data-testid="new-chat-landing-repo-branch-input"
          />
          <p className="text-xs text-muted-foreground">
            Cloned into the sandbox as the session's working directory. Leave blank to start in an empty workspace.
          </p>
        </div>
      </PopoverContent>
    </Popover>
  );
}
