"use client";

import { useCallback, useMemo, useState } from "react";
import type { KeyboardEvent } from "react";
import {
  buildWorkerSlashCommands,
  filterWorkerSlashCommands,
  getSlashQuery,
  parseWorkerSlashInput,
  type WorkerSlashCommand,
  type WorkerSlashCommandDef,
} from "@/lib/workerSlashCommands";

type Translate = (key: string) => string;

export function useWorkerSlashComposer(
  t: Translate,
  extraCommands: WorkerSlashCommandDef[] = [],
) {
  const commands = useMemo(
    () => buildWorkerSlashCommands(t, extraCommands),
    [t, extraCommands],
  );

  const [visible, setVisible] = useState(false);
  const [active, setActive] = useState(0);

  const syncMenu = useCallback((text: string, cursor: number) => {
    setVisible(!!getSlashQuery(text, cursor));
    setActive(0);
  }, []);

  const matchesFor = useCallback(
    (text: string, cursor: number) => {
      if (!getSlashQuery(text, cursor)) return [];
      const query = getSlashQuery(text, cursor)?.query ?? "";
      return filterWorkerSlashCommands(commands, query);
    },
    [commands],
  );

  const applySelection = useCallback(
    (command: WorkerSlashCommand, currentText: string): string => {
      if (command.hasArg) return `${command.label} `;
      return command.label;
    },
    [],
  );

  const resolveSubmit = useCallback(
    (
      text: string,
    ): { kind: "slash"; prompt: string } | { kind: "plain"; prompt: string } | null => {
      const trimmed = text.trim();
      if (!trimmed) return null;
      const parsed = parseWorkerSlashInput(trimmed, commands);
      if (parsed) {
        if (parsed.command.hasArg && !parsed.arg) {
          return null;
        }
        return { kind: "slash", prompt: trimmed };
      }
      return { kind: "plain", prompt: text };
    },
    [commands],
  );

  const handleKeyDown = useCallback(
    (
      e: KeyboardEvent<HTMLTextAreaElement>,
      text: string,
      matches: WorkerSlashCommand[],
      onSelect: (nextText: string) => void,
      onSubmit: () => void,
    ) => {
      if (e.nativeEvent.isComposing) return false;

      const safeActive = Math.min(active, Math.max(matches.length - 1, 0));

      if (visible && matches.length > 0) {
        if (e.key === "ArrowDown") {
          e.preventDefault();
          setActive((prev) => (prev < matches.length - 1 ? prev + 1 : 0));
          return true;
        }
        if (e.key === "ArrowUp") {
          e.preventDefault();
          setActive((prev) => (prev > 0 ? prev - 1 : matches.length - 1));
          return true;
        }
        if (e.key === "Enter" || e.key === "Tab") {
          e.preventDefault();
          onSelect(applySelection(matches[safeActive]!, text));
          setVisible(false);
          return true;
        }
        if (e.key === "Escape") {
          e.preventDefault();
          setVisible(false);
          return true;
        }
      }

      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        onSubmit();
        return true;
      }

      return false;
    },
    [active, applySelection, visible],
  );

  return {
    commands,
    visible,
    setVisible,
    active,
    syncMenu,
    matchesFor,
    applySelection,
    resolveSubmit,
    handleKeyDown,
  };
}
