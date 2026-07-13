import * as Blockly from "blockly";
import { useCallback, useEffect, useRef, useState } from "react";

import { workspaceToDraft } from "../blockly/workspace-to-draft";
import { compileLoop } from "../domain/compile-loop";
import type { CompileResult } from "../domain/loop-types";
import type { CustomBlockDefinition } from "../custom-blocks/custom-block-definition";
import {
  clearLoomProject,
  loadLoomProject,
  saveLoomProject,
  type OutputTab,
} from "../persistence/workspace-storage";

const EMPTY_RESULT: CompileResult = {
  diagnostics: [],
  executionBlockIds: [],
};

export function useLoomWorkspace() {
  const [loaded] = useState(loadLoomProject);
  const savedFingerprintRef = useRef<string | null>(null);
  const [workspace, setWorkspace] = useState<Blockly.WorkspaceSvg | null>(null);
  const [customDefinitions, setCustomDefinitions] = useState<
    CustomBlockDefinition[]
  >(loaded.project?.customDefinitions ?? []);
  const [compileResult, setCompileResult] = useState(EMPTY_RESULT);
  const [selectedBlock, setSelectedBlock] = useState<Blockly.BlockSvg | null>(null);
  const [outputTab, setOutputTabState] = useState<OutputTab>(
    loaded.project?.outputTab ?? "diagnostics",
  );
  const [dirty, setDirty] = useState(!loaded.project);
  const [loadError, setLoadError] = useState(loaded.error);

  const fingerprint = useCallback((
    target: Blockly.WorkspaceSvg,
    definitions: CustomBlockDefinition[],
    tab: OutputTab,
  ) => JSON.stringify({
    workspace: Blockly.serialization.workspaces.save(target),
    definitions,
    tab,
  }), []);

  const refresh = useCallback((
    target: Blockly.WorkspaceSvg,
    definitions = customDefinitions,
  ) => {
    setCompileResult(compileLoop(workspaceToDraft(target, definitions)));
  }, [customDefinitions]);

  const handleReady = useCallback((
    target: Blockly.WorkspaceSvg,
    loadedSuccessfully: boolean,
  ) => {
    setWorkspace(target);
    refresh(target);
    savedFingerprintRef.current = loaded.project && loadedSuccessfully
      ? fingerprint(target, customDefinitions, outputTab)
      : null;
    setDirty(!loaded.project || !loadedSuccessfully);
  }, [
    customDefinitions,
    fingerprint,
    loaded.project,
    outputTab,
    refresh,
  ]);

  const handleChange = useCallback((target: Blockly.WorkspaceSvg) => {
    refresh(target);
    const current = fingerprint(target, customDefinitions, outputTab);
    setDirty(savedFingerprintRef.current !== current);
  }, [customDefinitions, fingerprint, outputTab, refresh]);

  useEffect(() => {
    if (!workspace) return;
    refresh(workspace, customDefinitions);
  }, [customDefinitions, refresh, workspace]);

  const addCustomDefinition = useCallback((
    definition: CustomBlockDefinition,
  ) => {
    setCustomDefinitions((current) => {
      if (current.some(({ id }) => id === definition.id)) {
        setLoadError(`自定义积木 ID 重复：${definition.id}。`);
        return current;
      }
      const next = [...current, definition];
      if (workspace) {
        const currentFingerprint = fingerprint(workspace, next, outputTab);
        setDirty(savedFingerprintRef.current !== currentFingerprint);
      }
      return next;
    });
  }, [fingerprint, outputTab, workspace]);

  const save = useCallback(() => {
    if (!workspace) return;
    const result = saveLoomProject(workspace, customDefinitions, outputTab);
    if (!result.ok) {
      setLoadError(result.error);
      return;
    }
    savedFingerprintRef.current = fingerprint(
      workspace,
      customDefinitions,
      outputTab,
    );
    setDirty(false);
  }, [customDefinitions, fingerprint, outputTab, workspace]);

  const setOutputTab = useCallback((tab: OutputTab) => {
    setOutputTabState(tab);
    if (!workspace) return;
    const current = fingerprint(workspace, customDefinitions, tab);
    setDirty(savedFingerprintRef.current !== current);
  }, [customDefinitions, fingerprint, workspace]);

  const focusBlock = useCallback((blockId: string) => {
    if (!workspace) return;
    const block = workspace.getBlockById(blockId);
    if (!block) return;
    block.select();
    workspace.centerOnBlock(blockId);
  }, [workspace]);

  const resetStoredProject = useCallback(() => {
    const result = clearLoomProject();
    if (!result.ok) {
      setLoadError(result.error);
      return;
    }
    window.location.reload();
  }, []);

  return {
    addCustomDefinition,
    compileResult,
    customDefinitions,
    dirty,
    focusBlock,
    handleChange,
    handleReady,
    initialWorkspaceState: loaded.project?.workspaceState,
    loadError,
    outputTab,
    resetStoredProject,
    save,
    selectedBlock,
    setLoadError,
    setOutputTab,
    setSelectedBlock,
    workspace,
  };
}
