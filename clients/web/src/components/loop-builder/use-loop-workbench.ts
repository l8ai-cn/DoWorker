"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import {
  applyLoopCompile,
  listLoopRuntimeSnapshots,
  readLoopSnapshot,
  requestLoopCompile,
  runLoopProgram,
  setLoopActiveEditor,
  setLoopSource,
} from "@/lib/api/facade/loopProgramConnect";
import {
  createDefaultLoopSource,
  type LoopEditor,
  type LoopRuntimeSnapshot,
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
  const [runtimeSnapshots, setRuntimeSnapshots] = useState<LoopRuntimeSnapshot[]>([]);
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
        setError(
          cause instanceof Error ? cause.message : "循环脚本校验失败，请检查网络或稍后重试",
        );
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
        const source = createDefaultLoopSource();
        await compile(source, "blocks");
      } catch (cause) {
        if (!cancelled) {
          setError(cause instanceof Error ? cause.message : "循环工作台初始化失败");
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

  useEffect(() => {
    let cancelled = false;
    setRuntimeLoading(true);
    setRuntimeError(undefined);
    setRuntimeSnapshots([]);
    listLoopRuntimeSnapshots(orgSlug)
      .then((snapshots) => {
        if (!cancelled) setRuntimeSnapshots(snapshots);
      })
      .catch(() => {
        if (!cancelled) setRuntimeError("运行环境加载失败，请稍后重试");
      })
      .finally(() => {
        if (!cancelled) setRuntimeLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [orgSlug, runtimeLoadAttempt]);

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

  const run = useCallback(async (workerSnapshotId: string) => {
    setRunning(true);
    setError(undefined);
    try {
      setSnapshot(await runLoopProgram(orgSlug, snapshot.source, workerSnapshotId));
    } catch {
      setError("循环启动失败，请确认运行环境仍然可用");
    } finally {
      setRunning(false);
    }
  }, [orgSlug, snapshot.source]);

  return {
    snapshot,
    loading,
    running,
    error,
    runtimeSnapshots,
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
