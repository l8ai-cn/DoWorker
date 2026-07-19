import { useMemo } from "react";
import { useDirectorySessions } from "@/hooks/useDirectorySessions";
import { useRunnerHealthRegistration } from "@/hooks/RunnerHealthProvider";
import { normalizeWorkspacePath } from "./newChatWorkspace";

export function useNewChatWorkspaceOccupancy(selectedHostId: string | null, branchName: string) {
  const { data: directorySessions } = useDirectorySessions(true);
  const conflictCandidates = useMemo(
    () =>
      (directorySessions ?? []).filter(
        (session) => session.host_id === selectedHostId && session.workspace != null,
      ),
    [directorySessions, selectedHostId],
  );
  const runnerHealth = useRunnerHealthRegistration(conflictCandidates);
  const occupancyByDir = useMemo(() => {
    const counts = new Map<string, number>();
    for (const session of conflictCandidates) {
      if (session.workspace == null || runnerHealth.get(session.id) !== true) continue;
      const directory = normalizeWorkspacePath(session.workspace);
      if (directory) counts.set(directory, (counts.get(directory) ?? 0) + 1);
    }
    return counts;
  }, [conflictCandidates, runnerHealth]);

  return branchName.trim() === ""
    ? (path: string) => occupancyByDir.get(normalizeWorkspacePath(path) ?? "") ?? 0
    : undefined;
}
