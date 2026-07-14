"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import {
  applyLoopCompile,
  readLoopSnapshot,
  requestLoopCompile,
  runLoopProgram,
  setLoopActiveEditor,
  setLoopSource,
} from "@/lib/api/facade/loopProgramConnect";
import { listGoalLoopWorkerSnapshots } from "@/lib/api/facade/goalLoopConnect";
import {
  createDefaultLoopSource,
  type LoopEditor,
  type LoopWorkbenchSnapshot,
} from "@/lib/viewModels/loop-program";

const EMPTY: LoopWorkbenchSnapshot = {
  source: "",
  canonicalSource: "",
  diagnostics: [],
  parseStatus: "empty",
  activeEditor: "blocks",
  revision: 0,
  semanticRevision: 0,
};

export function useLoopWorkbench(orgSlug: string) {
  const [snapshot, setSnapshot] = useState<LoopWorkbenchSnapshot>(EMPTY);
  const [loading, setLoading] = useState(true);
  const [running, setRunning] = useState(false);
  const [error, setError] = useState<string>();
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
        setError(cause instanceof Error ? cause.message : "Loop 编译失败");
      }
    }, delay);
  }, [orgSlug]);

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
        const workers = await listGoalLoopWorkerSnapshots(orgSlug);
        if (cancelled) return;
        if (workers.length === 0) {
          throw new Error("没有可执行的 Worker 快照，请先创建或更新 Worker");
        }
        const source = createDefaultLoopSource(workers[0].id);
        await compile(source, "blocks");
      } catch (cause) {
        if (!cancelled) {
          setError(cause instanceof Error ? cause.message : "Loop 初始化失败");
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
  }, [compile, orgSlug]);

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

  const run = useCallback(async () => {
    setRunning(true);
    setError(undefined);
    try {
      setSnapshot(await runLoopProgram(orgSlug, snapshot.source));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Loop 运行失败");
    } finally {
      setRunning(false);
    }
  }, [orgSlug, snapshot.source]);

  return {
    snapshot,
    loading,
    running,
    error,
    setEditor,
    updateBlocks,
    updateCode,
    run,
  };
}
