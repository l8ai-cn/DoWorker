import { useState } from "react";
import { useNavigate } from "@/lib/routing";
import { useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAvailableAgents, type AvailableAgent } from "@/hooks/useAvailableAgents";
import { useSessionAgent } from "@/hooks/useAgents";
import { agentRootName, forkTargetCarriesHistory } from "@/lib/forkHarness";
import { hasWorkerCreationSelection, workerCreationSelection } from "@/lib/workerCreationSelection";
import { forkSnapshotSession, forkWorkerSession } from "@/lib/workerSessionMutations";

const SAME_AS_SOURCE = "__same__";

export function ForkSessionForm({
  sourceSessionId,
  sourceTitle,
  upToResponseId,
  onClose,
}: {
  sourceSessionId: string;
  sourceTitle?: string | null;
  sourceWorkspace?: string | null;
  sourceHostId?: string | null;
  sourceGitBranch?: string | null;
  upToResponseId?: string | null;
  onClose: () => void;
}) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [title, setTitle] = useState("");
  const [agentChoice, setAgentChoice] = useState(SAME_AS_SOURCE);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { data: sourceAgent } = useSessionAgent(sourceSessionId);
  const { data: agents } = useAvailableAgents({ enabled: true });
  const sourceRoot = sourceAgent ? agentRootName(sourceAgent.name) : null;
  const choices = (agents ?? []).filter(
    (agent) =>
      hasWorkerCreationSelection(agent) &&
      forkTargetCarriesHistory(agent.harness) &&
      agent.id !== sourceAgent?.id &&
      agent.name !== sourceAgent?.name &&
      agent.name !== sourceRoot,
  );
  const target = choices.find((agent) => agent.id === agentChoice) ?? null;
  const canSubmit = agentChoice === SAME_AS_SOURCE || (target !== null && sourceAgent !== undefined);
  const placeholder = sourceTitle?.trim() ? `Fork of ${sourceTitle.trim()}` : "Name the cloned session";

  async function submit(): Promise<void> {
    if (!canSubmit) return;
    setSubmitting(true);
    setError(null);
    try {
      const titleValue = title.trim() || undefined;
      const fork =
        agentChoice === SAME_AS_SOURCE
          ? await forkSnapshotSession({
              sourceId: sourceSessionId,
              title: titleValue,
              upToResponseId: upToResponseId ?? undefined,
            })
          : await forkWithAgent(sourceSessionId, sourceAgent?.id, target, titleValue, upToResponseId);
      void queryClient.invalidateQueries({ queryKey: ["conversations"] });
      onClose();
      navigate(`/c/${fork.id}`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Couldn't clone the session. Try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form
      className="flex flex-col gap-4"
      onSubmit={(event) => {
        event.preventDefault();
        void submit();
      }}
    >
      <label className="flex flex-col gap-1.5 text-xs font-medium text-muted-foreground">
        Name
        <input
          data-testid="fork-session-title-input"
          value={title}
          onChange={(event) => setTitle(event.target.value)}
          placeholder={placeholder}
          className="h-9 rounded-md border border-input bg-background px-3 text-sm text-foreground"
        />
      </label>
      <label className="flex flex-col gap-1.5 text-xs font-medium text-muted-foreground">
        Agent
        <Select value={agentChoice} onValueChange={setAgentChoice}>
          <SelectTrigger data-testid="fork-session-agent-select">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={SAME_AS_SOURCE} data-testid="fork-session-agent-option-same">
              {sourceAgent?.name ?? "Same as source"} (same as original session)
            </SelectItem>
            {choices.map((agent) => (
              <SelectItem
                key={agent.id}
                value={agent.id}
                data-testid={`fork-session-agent-option-${agent.id}`}
              >
                {agent.display_name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </label>
      {agentChoice !== SAME_AS_SOURCE && target === null && (
        <p data-testid="fork-session-error" className="text-xs text-destructive">
          The selected Agent cannot start because authoritative Worker metadata is unavailable.
        </p>
      )}
      {error !== null && (
        <p data-testid="fork-session-error" className="text-xs text-destructive">
          {error}
        </p>
      )}
      <div className="flex justify-end gap-2">
        <Button type="button" variant="ghost" onClick={onClose} disabled={submitting}>
          Cancel
        </Button>
        <Button type="submit" data-testid="fork-session-submit" disabled={!canSubmit || submitting}>
          {submitting ? "Cloning…" : "Clone session"}
        </Button>
      </div>
    </form>
  );
}

async function forkWithAgent(
  sourceId: string,
  sourceAgentId: string | undefined,
  target: AvailableAgent | null,
  title: string | undefined,
  upToResponseId: string | null | undefined,
) {
  if (sourceAgentId === undefined || target === null || !hasWorkerCreationSelection(target)) {
    throw new Error("The selected Agent cannot start because authoritative Worker metadata is unavailable");
  }
  return forkWorkerSession({
    sourceId,
    sourceAgentId,
    agentId: target.id,
    title,
    upToResponseId: upToResponseId ?? undefined,
    ...workerCreationSelection(target),
  });
}
