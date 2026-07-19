import { useState } from "react";
import { useNavigate } from "@/lib/routing";
import { useQueryClient } from "@tanstack/react-query";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { AgentCard } from "@/components/AgentCard";
import { useAvailableAgents } from "@/hooks/useAvailableAgents";
import { importWorkerSession } from "@/lib/workerSessionMutations";
import { hasWorkerCreationSelection, workerCreationSelection } from "@/lib/workerCreationSelection";

/**
 * "Import from Codex" dialog.
 *
 * Migrates a local Codex conversation record into a brand-new Worker session
 * via ``POST /v1/sessions/import``. The user supplies a server-local path to a
 * Codex rollout transcript (``rollout-*.jsonl``) or a workflow ``output_*``
 * directory (auto-detected), picks the target agent, and on submit the server
 * converts + persists the transcript into a fresh Worker. We then navigate to
 * ``/c/{id}`` so the migrated history renders as an ordinary conversation.
 */
export function ImportCodexDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { data: agents } = useAvailableAgents({ enabled: open });

  const agentList = (agents ?? []).filter(hasWorkerCreationSelection);
  const unavailableAgentCount = (agents?.length ?? 0) - agentList.length;
  const [sourcePath, setSourcePath] = useState("");
  const [title, setTitle] = useState("");
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const selectedAgent = agentList.find((a) => a.id === selectedAgentId) ?? null;

  function handleOpenChange(next: boolean): void {
    if (!next) {
      setSourcePath("");
      setTitle("");
      setSelectedAgentId(null);
      setError(null);
      setSubmitting(false);
    }
    onOpenChange(next);
  }

  async function handleImport(): Promise<void> {
    const path = sourcePath.trim();
    if (!path) {
      setError("Enter the path to a Codex rollout .jsonl or output_* directory.");
      return;
    }
    if (selectedAgent === null) {
      setError("Pick a target agent for the migrated Worker.");
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      const result = await importWorkerSession({
        agentId: selectedAgent.id,
        sourcePath: path,
        title: title.trim() || undefined,
        ...workerCreationSelection(selectedAgent),
      });
      // Refresh the sidebar lists so the migrated Worker appears immediately.
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["conversations"] }),
        queryClient.invalidateQueries({ queryKey: ["project-sessions"] }),
      ]);
      handleOpenChange(false);
      navigate(`/c/${result.id}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Migration failed. Check the path and try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent
        data-testid="import-codex-dialog"
        className="flex max-h-[85vh] flex-col gap-4 sm:max-w-lg"
      >
        <DialogHeader>
          <DialogTitle>Import from Codex</DialogTitle>
        </DialogHeader>

        <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto">
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="import-codex-source"
              className="text-xs font-medium text-muted-foreground"
            >
              Codex source path
            </label>
            <input
              id="import-codex-source"
              data-testid="import-codex-source-input"
              type="text"
              value={sourcePath}
              onChange={(e) => setSourcePath(e.target.value)}
              placeholder="~/.codex/sessions/2026/07/08/rollout-….jsonl or output_20260708_161508"
              className="rounded-md border border-input bg-background px-3 py-2 font-mono text-xs outline-none transition-colors focus-visible:border-ring"
            />
            <p className="text-[11px] text-muted-foreground">
              A rollout <code className="font-mono">.jsonl</code> transcript or a workflow
              <code className="font-mono"> output_*</code> directory on the server. Auto-detected.
            </p>
          </div>

          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="import-codex-title"
              className="text-xs font-medium text-muted-foreground"
            >
              Title (optional)
            </label>
            <input
              id="import-codex-title"
              data-testid="import-codex-title-input"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Derived from the first prompt when left blank"
              className="rounded-md border border-input bg-background px-3 py-2 font-mono text-xs outline-none transition-colors focus-visible:border-ring"
            />
          </div>

          <div className="flex flex-col gap-2">
            <span className="text-xs font-medium text-muted-foreground">Target agent</span>
            {agentList.length === 0 ? (
              <p data-testid="import-codex-empty" className="text-xs text-muted-foreground">
                No agents available on this server. Register one with{" "}
                <code className="font-mono">omnigent server --agent</code>.
              </p>
            ) : (
              agentList.map((agent) => (
                <AgentCard
                  key={agent.id}
                  agent={agent}
                  selected={agent.id === selectedAgentId}
                  onSelect={() => {
                    setSelectedAgentId(agent.id);
                    setError(null);
                  }}
                  hover
                />
              ))
            )}
            {unavailableAgentCount > 0 && (
              <p data-testid="import-codex-unavailable" className="text-xs text-muted-foreground">
                Some agents cannot import sessions because their Worker creation metadata is unavailable.
              </p>
            )}
          </div>

          {error !== null && (
            <p data-testid="import-codex-error" className="text-xs text-destructive">
              {error}
            </p>
          )}
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => handleOpenChange(false)} disabled={submitting}>
            Cancel
          </Button>
          <Button
            data-testid="import-codex-submit"
            onClick={handleImport}
            disabled={!sourcePath.trim() || selectedAgent === null || submitting}
          >
            {submitting ? "Migrating…" : "Migrate & open"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
