import { type Dispatch, type RefObject, type SetStateAction, useEffect, useMemo, useState } from "react";
import type { AvailableAgent } from "@/hooks/useAvailableAgents";
import { SKILL_PILL_AGENTS } from "./newChatConstants";

export function useNewChatSlashState({
  selectedAgent,
  isNativeTerminalAgent,
  message,
  setMessage,
  textareaRef,
}: {
  selectedAgent: AvailableAgent | undefined;
  isNativeTerminalAgent: boolean;
  message: string;
  setMessage: Dispatch<SetStateAction<string>>;
  textareaRef: RefObject<HTMLTextAreaElement | null>;
}) {
  const [slashMenuIndex, setSlashMenuIndex] = useState(-1);
  const skillCommands = useMemo(() => {
    if (isNativeTerminalAgent) return {};
    const commands: Record<string, string> = {};
    for (const skill of selectedAgent?.skills ?? []) commands[`/${skill.name}`] = skill.description;
    return commands;
  }, [selectedAgent, isNativeTerminalAgent]);
  const trimmedMessage = message.trimStart();
  const slashMenuOpen =
    trimmedMessage.startsWith("/") &&
    !trimmedMessage.slice(1).includes("/") &&
    !trimmedMessage.includes(" ");
  const slashMenuQuery = slashMenuOpen ? trimmedMessage.slice(1) : "";
  const slashMenuMatches = useMemo(
    () =>
      slashMenuOpen
        ? Object.keys(skillCommands).filter((name) =>
            name.slice(1).startsWith(slashMenuQuery.toLowerCase()),
          )
        : [],
    [skillCommands, slashMenuOpen, slashMenuQuery],
  );
  useEffect(() => {
    setSlashMenuIndex(slashMenuMatches.length > 0 ? 0 : -1);
  }, [slashMenuMatches]);

  const applySlashSelection = (command: string) => {
    setSlashMenuIndex(-1);
    setMessage(command + " ");
    textareaRef.current?.focus();
  };
  const pillSkills =
    selectedAgent && SKILL_PILL_AGENTS.has(selectedAgent.name) ? selectedAgent.skills : [];
  const applySkillPill = (name: string) => {
    setMessage(`/${name} `);
    textareaRef.current?.focus();
  };

  return {
    slashMenuIndex,
    setSlashMenuIndex,
    slashMenuOpen,
    slashMenuQuery,
    slashMenuMatches,
    skillCommands,
    applySlashSelection,
    pillSkills,
    applySkillPill,
  };
}
