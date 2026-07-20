import { useEffect } from "react";
import {
  readLoopSnapshot,
  restoreLoopResourceProgram,
} from "@/lib/api/facade/loopProgramConnect";
import { createDefaultLoopSource, type LoopEditor, type LoopWorkbenchSnapshot } from "@/lib/viewModels/loop-program";
import type { LoopErrorMessages } from "./loop-workbench-messages";

interface LoopWorkbenchInitialization {
  orgSlug: string;
  resourceName?: string;
  activeEditorRef: { current: LoopEditor };
  compile: (source: string, editor: LoopEditor, delay?: number) => Promise<void>;
  setError: (error?: string) => void;
  setLoading: (loading: boolean) => void;
  setSnapshot: (snapshot: LoopWorkbenchSnapshot) => void;
}

export function useLoopWorkbenchInitialization({
  orgSlug,
  resourceName,
  activeEditorRef,
  compile,
  setError,
  setLoading,
  setSnapshot,
}: LoopWorkbenchInitialization) {
  useEffect(() => {
    let cancelled = false;
    async function initialize() {
      try {
        if (resourceName) {
          const restored = await restoreLoopResourceProgram(orgSlug, resourceName);
          if (!cancelled) {
            activeEditorRef.current = "blocks";
            setSnapshot(restored);
          }
          return;
        }
        const current = await readLoopSnapshot();
        if (cancelled) return;
        activeEditorRef.current = current.activeEditor;
        if (current.source) {
          setSnapshot(current);
          return;
        }
        await compile(createDefaultLoopSource(), "blocks");
      } catch (cause) {
        if (!cancelled) {
          setError(cause instanceof Error ? cause.message : "Unable to initialize Loop Builder.");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void initialize();
    return () => {
      cancelled = true;
    };
  }, [
    activeEditorRef,
    compile,
    orgSlug,
    resourceName,
    setError,
    setLoading,
    setSnapshot,
  ]);
}
