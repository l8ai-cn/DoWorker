import type * as Blockly from "blockly";
import { useCallback, useEffect, useRef, useState } from "react";

import { setWorkspaceEditing } from "../blockly/workspace-editing";
import type { CompileResult } from "../domain/loop-types";
import {
  runSimulation,
  type SimulationEvidence,
} from "../simulation/run-simulation";

export function useLoopSimulation(
  workspace: Blockly.WorkspaceSvg | null,
  compileResult: CompileResult,
) {
  const [evidence, setEvidence] = useState<SimulationEvidence[]>([]);
  const [running, setRunning] = useState(false);
  const [error, setError] = useState<string>();
  const abortRef = useRef<AbortController | null>(null);
  const mountedRef = useRef(true);

  useEffect(() => () => {
    mountedRef.current = false;
    abortRef.current?.abort();
  }, []);

  const stop = useCallback(() => abortRef.current?.abort(), []);

  const start = useCallback(async () => {
    if (!workspace || !compileResult.program || running) return false;
    const controller = new AbortController();
    const program = structuredClone(compileResult.program);
    const executionBlockIds = [...compileResult.executionBlockIds];
    abortRef.current = controller;
    setEvidence([]);
    setError(undefined);
    setRunning(true);
    setWorkspaceEditing(workspace, false);
    try {
      await runSimulation(program, executionBlockIds, {
        signal: controller.signal,
        onEvidence: (event) => {
          if (mountedRef.current) {
            setEvidence((current) => [...current, event]);
          }
        },
        onHighlight: (blockId) => {
          if (mountedRef.current) workspace.highlightBlock(blockId);
        },
      });
    } catch (caught) {
      if (!(caught instanceof DOMException && caught.name === "AbortError")) {
        if (mountedRef.current) {
          setError(caught instanceof Error ? caught.message : "模拟运行失败。");
        }
      }
    } finally {
      if (mountedRef.current) {
        workspace.highlightBlock(null);
        setWorkspaceEditing(workspace, true);
        setRunning(false);
        abortRef.current = null;
      }
    }
    return true;
  }, [compileResult, running, workspace]);

  return { error, evidence, running, start, stop };
}
