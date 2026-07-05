import { useEffect, useMemo } from "react";
import type { RunnerData } from "@/lib/api";
import {
  hasRunnerForAgent,
  runnerSupportsAgent,
  runnersSupportingAgent,
} from "@/lib/runner-agent-capabilities";

interface Args {
  runners: RunnerData[];
  selectedAgent: string | null;
  selectedRunner: RunnerData | null;
  setSelectedRunnerId: (id: number | null) => void;
}

export function useCreatePodRunnerCompatibility({
  runners,
  selectedAgent,
  selectedRunner,
  setSelectedRunnerId,
}: Args) {
  const compatibleRunners = useMemo(
    () => runnersSupportingAgent(runners, selectedAgent),
    [runners, selectedAgent],
  );
  const selectedRunnerCompatible =
    !selectedRunner || runnerSupportsAgent(selectedRunner, selectedAgent);
  const canCreate =
    Boolean(selectedAgent) &&
    hasRunnerForAgent(runners, selectedAgent) &&
    selectedRunnerCompatible;

  useEffect(() => {
    if (selectedAgent && selectedRunner && !selectedRunnerCompatible) {
      setSelectedRunnerId(null);
    }
  }, [selectedAgent, selectedRunner, selectedRunnerCompatible, setSelectedRunnerId]);

  return { compatibleRunners, canCreate };
}
