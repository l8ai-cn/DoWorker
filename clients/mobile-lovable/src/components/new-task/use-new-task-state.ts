import { useState } from "react";
import type { AgentPickerOption } from "@/lib/agent-display";
import type { CodexGoalMode } from "@/components/new-task-goal-panel";
import type { LiveExpert } from "@/lib/experts-api";
import { detectSlashToken, type SlashCommand } from "@/components/slash-menu";
import type { SessionInteractionMode } from "@/lib/sessions-api";

export interface NewTaskSearch {
  expert?: string;
  project?: string;
  prompt?: string;
}

interface NewTaskStateInput {
  search: NewTaskSearch;
  engines: AgentPickerOption[];
  projectNames: string[];
  initialProjectName: string;
}

export function useNewTaskState({
  search,
  engines,
  projectNames,
  initialProjectName,
}: NewTaskStateInput) {
  const [selectedEngineID, setSelectedEngineID] = useState("codex-cli");
  const [enginePickerOpen, setEnginePickerOpen] = useState(false);
  const [expertSlug, setExpertSlug] = useState<string | undefined>(search.expert);
  const [expertPickerOpen, setExpertPickerOpen] = useState(false);
  const [prompt, setPrompt] = useState(search.prompt ?? "");
  const [projectName, setProjectName] = useState(initialProjectName);
  const [projectPickerOpen, setProjectPickerOpen] = useState(false);
  const [asGoal, setGoalEnabled] = useState(false);
  const [interactionMode, setSelectedInteractionMode] = useState<SessionInteractionMode>("acp");
  const [tokenBudget, setTokenBudget] = useState("");
  const [goalMode, setGoalMode] = useState<CodexGoalMode>("active");
  const [successCriteria, setSuccessCriteria] = useState("");
  const [slashToken, setSlashToken] = useState<{ start: number; token: string } | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const engineID = engines.some((engine) => engine.id === selectedEngineID)
    ? selectedEngineID
    : (engines[0]?.id ?? selectedEngineID);
  const normalizedInteractionMode = resolveTaskInteractionMode(engines, engineID, interactionMode);
  const normalizedProjectName = projectNames.includes(projectName)
    ? projectName
    : (projectNames[0] ?? projectName);

  function setEngineID(id: string) {
    setSelectedEngineID(id);
  }

  function selectWorker(id: string) {
    setSelectedEngineID(id);
    setExpertSlug(undefined);
  }

  function setAsGoal(value: boolean) {
    setGoalEnabled(value);
    if (value) {
      setSelectedEngineID("codex-cli");
      setExpertSlug(undefined);
      setSelectedInteractionMode("acp");
    }
  }

  function setInteractionMode(mode: SessionInteractionMode) {
    setSelectedInteractionMode(mode);
    if (mode === "pty") setExpertSlug(undefined);
  }

  function onPromptChange(value: string, caret: number) {
    setPrompt(value);
    setSlashToken(detectSlashToken(value, caret));
  }

  function applySlash(command: SlashCommand, textarea: HTMLTextAreaElement) {
    if (!slashToken) return;
    const before = prompt.slice(0, slashToken.start);
    const after = prompt.slice(slashToken.start + slashToken.token.length);
    const insert = `${command.cmd} `;
    setPrompt(before + insert + after);
    setSlashToken(null);
    if (command.cmd === "/goal") setAsGoal(true);
    requestAnimationFrame(() => {
      textarea.focus();
      textarea.setSelectionRange((before + insert).length, (before + insert).length);
    });
  }

  return {
    engineID,
    setEngineID,
    selectWorker,
    enginePickerOpen,
    setEnginePickerOpen,
    expertSlug,
    setExpertSlug,
    expertPickerOpen,
    setExpertPickerOpen,
    prompt,
    projectName: normalizedProjectName,
    setProjectName,
    projectPickerOpen,
    setProjectPickerOpen,
    asGoal,
    setAsGoal,
    interactionMode: normalizedInteractionMode,
    setInteractionMode,
    tokenBudget,
    setTokenBudget,
    goalMode,
    setGoalMode,
    successCriteria,
    setSuccessCriteria,
    slashToken,
    setSlashToken,
    submitError,
    setSubmitError,
    submitting,
    setSubmitting,
    onPromptChange,
    applySlash,
  };
}

export function findSelectedExpert(experts: LiveExpert[], slug: string | undefined) {
  return experts.find((expert) => expert.slug === slug);
}

export function availableAcpExperts(experts: LiveExpert[], engines: AgentPickerOption[]) {
  return experts.filter((expert) => {
    const worker = engines.find((engine) => engine.id === expert.agent_slug);
    return expert.interaction_mode === "acp" && worker?.supportedModes.includes("acp");
  });
}

export function resolveTaskInteractionMode(
  engines: AgentPickerOption[],
  engineID: string,
  requested: SessionInteractionMode,
): SessionInteractionMode {
  const selected = engines.find((engine) => engine.id === engineID);
  return selected?.supportedModes.includes(requested)
    ? requested
    : (selected?.supportedModes[0] ?? requested);
}
