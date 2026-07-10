import { useMemo, useState } from "react";
import type { AgentData } from "@/lib/api";
import { POD_MODE_ACP, POD_MODE_PTY, type PodMode } from "@/lib/pod-modes";

export function useCreatePodInteractionMode(
  selectedAgent: string | null,
  availableAgents: AgentData[],
  automationLevel: string,
) {
  const [requestedInteractionMode, setInteractionMode] = useState<PodMode>(POD_MODE_PTY);
  const supportedModes = useMemo(() => {
    if (!selectedAgent) return [POD_MODE_PTY];
    const agent = availableAgents.find((a) => a.slug === selectedAgent);
    const raw = agent?.supported_modes;
    const modes = Array.isArray(raw)
      ? raw.map((m: string) => m.trim()).filter(Boolean)
      : (typeof raw === "string" ? raw.split(",").map((m: string) => m.trim()).filter(Boolean) : []);
    return modes.length > 0 ? modes : [POD_MODE_PTY];
  }, [selectedAgent, availableAgents]);

  const interactionMode = useMemo<PodMode>(() => {
    if (!selectedAgent) return POD_MODE_PTY;
    if (automationLevel === "autonomous" && supportedModes.includes(POD_MODE_ACP)) {
      return POD_MODE_ACP;
    }
    if (supportedModes.includes(requestedInteractionMode)) return requestedInteractionMode;
    return (supportedModes[0] ?? POD_MODE_PTY) as PodMode;
  }, [automationLevel, requestedInteractionMode, selectedAgent, supportedModes]);

  return { interactionMode, setInteractionMode, supportedModes };
}
