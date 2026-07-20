"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import {
  applyLoopCompile,
  listLoopRuntimeTemplates,
  readLoopSnapshot,
  requestLoopCompile,
  runLoopResourceProgram,
  setLoopActiveEditor,
  setLoopSource,
} from "@/lib/api/facade/loopProgramConnect";
import {
  createDefaultLoopSource,
  type LoopEditor,
  type LoopRuntimeTemplate,
  type LoopWorkbenchSnapshot,
} from "@/lib/viewModels/loop-program";
import { createGoalLoopResourceDocument } from "./loop-resource-document";
import type { LoopErrorMessages } from "./loop-workbench-messages";

const EMPTY: LoopWorkbenchSnapshot = {
  source: "",
  canonicalSource: "",
  diagnostics: [],
  parseStatus: "empty",
  activeEditor: "blocks",
  revision: 0,
  semanticRevision: 0,
};

export function useLoopWorkbench(orgSlug: string, messages: LoopErrorMessages) {
  const [snapshot, setSnapshot] = useState<LoopWorkbenchSnapshot>(EMPTY);
  const [loading, setLoading] = useState(true);
  const [running, setRunning] = useState(false);
  const [error, setError] = useState<string>();
  const [runtimeTemplates, setRuntimeTemplates] = useState<LoopRuntimeTemplate[]>([]);
  const [runtimeLoading, setRuntimeLoading] = useState(true);
  const [runtimeError, setRuntimeError] = useState<string>();
  const [runtimeLoadAttempt, setRuntimeLoadAttempt] = useState(0);
  const activeEditorRef = useRef<LoopEditor>("blocks");
  const compileSequence = useRef(0);
  const compileTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  const compile = useCallback(async (
    source: string,
    editor: LoopEditor,
    delay = 0,
  ) => {
    const sequence = ++compileSequence.current;
    setError(undefined);
    const draft = await setLoopSource(source, editor);
    if (sequence !== compileSequence.current) return;
    setSnapshot(draft);
    if (compileTimer.current) clearTimeout(compileTimer.current);
    compileTimer.current = setTimeout(async () => {
      try {
        const response = await requestLoopCompile(orgSlug, source, draft.revision);
        if (sequence !== compileSequence.current) return;
        setSnapshot(await applyLoopCompile(response));
      } catch (cause) {
        if (sequence !== compileSequence.current) return;
        setError(cause instanceof Error ? cause.message : messages.compileFailed);
      }
    }, delay);
  }, [messages.compileFailed, orgSlug]);

  useEffect(() => {
    let cancelled = false;
    async function initialize() {
      try {
        const current = await readLoopSnapshot();
        if (cancelled) return;
        activeEditorRef.current = current.activeEditor;
        if (current.source) {
          setSnapshot(current);
          return;
        }
        const source = createDefaultLoopSource();
        await compile(source, "blocks");
      } catch (cause) {
        if (!cancelled) {
          setError(cause instanceof Error ? cause.message : messages.initFailed);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void initialize();
    return () => {
      cancelled = true;
      compileSequence.current += 1;
      if (compileTimer.current) clearTimeout(compileTimer.current);
    };
  }, [compile, messages.initFailed, orgSlug]);

  useEffect(() => {
    let cancelled = false;
    setRuntimeLoading(true);
    setRuntimeError(undefined);
    setRuntimeTemplates([]);
    listLoopRuntimeTemplates(orgSlug)
      .then((templates) => {
        if (!cancelled) setRuntimeTemplates(templates);
      })
      .catch(() => {
        if (!cancelled) setRuntimeError(messages.runtimeLoadFailed);
      })
      .finally(() => {
        if (!cancelled) setRuntimeLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [messages.runtimeLoadFailed, orgSlug, runtimeLoadAttempt]);

  const setEditor = useCallback(async (editor: LoopEditor) => {
    activeEditorRef.current = editor;
    setSnapshot(await setLoopActiveEditor(editor));
  }, []);

  const updateBlocks = useCallback((source: string) => {
    if (activeEditorRef.current === "blocks") {
      void compile(source, "blocks", 120);
    }
  }, [compile]);

  const updateCode = useCallback((source: string) => {
    if (activeEditorRef.current === "code") {
      void compile(source, "code", 300);
    }
  }, [compile]);

  const retryRuntimeLoad = useCallback(() => {
    setRuntimeLoadAttempt((attempt) => attempt + 1);
  }, []);

  const applySnapshot = useCallback((next: LoopWorkbenchSnapshot) => {
    activeEditorRef.current = next.activeEditor;
    setSnapshot(next);
  }, []);

  const run = useCallback(async (workerTemplateName: string) => {
    setRunning(true);
    setError(undefined);
    try {
      const document = createGoalLoopResourceDocument({
        namespace: orgSlug,
        program: snapshot.program,
        workerTemplateName,
      });
      setSnapshot(await runLoopResourceProgram(orgSlug, document));
    } catch (cause) {
      setError(loopRunErrorMessage(cause, messages));
    } finally {
      setRunning(false);
    }
  }, [messages, orgSlug, snapshot.program]);

  return {
    snapshot,
    loading,
    running,
    error,
    runtimeTemplates,
    runtimeLoading,
    runtimeError,
    retryRuntimeLoad,
    applySnapshot,
    setEditor,
    updateBlocks,
    updateCode,
    run,
  };
}

function loopRunErrorMessage(cause: unknown, messages: LoopErrorMessages): string {
  const detail = cause instanceof Error ? cause.message : String(cause ?? "");
  if (detail.includes("validate-plan-apply")) {
    return messages.runRequiresPlanApply;
  }
  return detail
    ? detail
    : messages.runFailed;
}
