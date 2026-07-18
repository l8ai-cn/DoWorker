"use client";

import * as Blockly from "blockly";
import { useEffect, useRef, useState } from "react";
import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";
import { setLoopBlockAccess } from "./loop-block-access";
import {
  insertPointFromDoubleClick,
  type LoopBlockInsertPoint,
} from "./loop-block-insert-point";
import { createLoopBlockCatalog, registerLoopBlocks } from "./loop-block-catalog";
import { loopBlocklyTheme } from "./loop-blockly-theme";
import type { LoopCustomBlockDefinition } from "./loop-custom-block-types";
import { nodeIdForBlock, projectProgramToWorkspace, workspaceToLoopSource } from "./loop-block-projection";
import { LoopQuickInsert } from "./loop-quick-insert";
import type { LoopBlockCatalogMessages, LoopQuickInsertMessages } from "./loop-workbench-messages";

interface LoopBlocklyCanvasProps {
  program?: LoopProgram;
  semanticRevision: number;
  readOnly: boolean;
  customDefinitions: readonly LoopCustomBlockDefinition[];
  messages: {
    blockly: LoopBlockCatalogMessages;
    quickInsert: LoopQuickInsertMessages;
  };
  onCreateCustom: () => void;
  onSourceChange: (source: string) => void;
  onSelectNode: (nodeId?: string) => void;
}

export function LoopBlocklyCanvas({
  program,
  semanticRevision,
  readOnly,
  customDefinitions,
  onSourceChange,
  onCreateCustom,
  onSelectNode,
  messages,
}: LoopBlocklyCanvasProps) {
  const hostRef = useRef<HTMLDivElement>(null);
  const workspaceRef = useRef<Blockly.WorkspaceSvg | undefined>(undefined);
  const programRef = useRef(program);
  const readOnlyRef = useRef(readOnly);
  const customDefinitionsRef = useRef(customDefinitions);
  const callbackRef = useRef({ onSourceChange, onSelectNode });
  const [insertPoint, setInsertPoint] = useState<LoopBlockInsertPoint>();

  useEffect(() => {
    callbackRef.current = { onSourceChange, onSelectNode };
  }, [onSourceChange, onSelectNode]);

  useEffect(() => {
    programRef.current = program;
  }, [program]);

  useEffect(() => {
    readOnlyRef.current = readOnly;
  }, [readOnly]);

  useEffect(() => {
    customDefinitionsRef.current = customDefinitions;
  }, [customDefinitions]);

  useEffect(() => {
    if (!hostRef.current) return;
    registerLoopBlocks(messages.blockly, customDefinitionsRef.current);
    const { toolbox } = createLoopBlockCatalog(
      messages.blockly,
      customDefinitionsRef.current,
    );
    const compact = hostRef.current.clientWidth < 640;
    const workspace = Blockly.inject(hostRef.current, {
      media: "/blockly-media/",
      toolbox,
      theme: loopBlocklyTheme,
      renderer: "zelos",
      trashcan: true,
      grid: { spacing: 20, length: 1, colour: "#cbd5e1", snap: true },
      zoom: {
        controls: true,
        wheel: true,
        startScale: compact ? 0.55 : 0.78,
        minScale: 0.4,
        maxScale: 1.25,
      },
      move: { scrollbars: true, drag: true, wheel: true },
    });
    workspaceRef.current = workspace;
    const listener = (event: Blockly.Events.Abstract) => {
      if (event.type === Blockly.Events.SELECTED) {
        const selected = Blockly.getSelected();
        callbackRef.current.onSelectNode(
          selected instanceof Blockly.Block ? nodeIdForBlock(selected) : undefined,
        );
      }
      if (!event.isUiEvent) {
        callbackRef.current.onSourceChange(
          workspaceToLoopSource(workspace, customDefinitionsRef.current).source,
        );
      }
    };
    workspace.addChangeListener(listener);
    const currentProgram = programRef.current;
    if (currentProgram) {
      projectProgramToWorkspace(workspace, currentProgram, customDefinitionsRef.current);
      setLoopBlockAccess(workspace, readOnlyRef.current);
    }
    const observer = new ResizeObserver(() => Blockly.svgResize(workspace));
    observer.observe(hostRef.current);
    return () => {
      observer.disconnect();
      workspace.removeChangeListener(listener);
      workspace.dispose();
      workspaceRef.current = undefined;
    };
  }, [messages.blockly]);

  useEffect(() => {
    const workspace = workspaceRef.current;
    const currentProgram = programRef.current;
    if (workspace && currentProgram) {
      projectProgramToWorkspace(workspace, currentProgram, customDefinitionsRef.current);
    }
  }, [semanticRevision]);

  useEffect(() => {
    const workspace = workspaceRef.current;
    if (!workspace) return;
    registerLoopBlocks(messages.blockly, customDefinitions);
    workspace.updateToolbox(createLoopBlockCatalog(messages.blockly, customDefinitions).toolbox);
    callbackRef.current.onSourceChange(workspaceToLoopSource(workspace, customDefinitions).source);
  }, [customDefinitions, messages.blockly]);

  useEffect(() => {
    const workspace = workspaceRef.current;
    if (!workspace) return;
    setLoopBlockAccess(workspace, readOnly);
  }, [readOnly, semanticRevision]);

  function insert(type: string) {
    const workspace = workspaceRef.current;
    if (!workspace || !insertPoint || readOnly) return;
    const created = workspace.newBlock(type) as Blockly.BlockSvg;
    created.initSvg();
    created.render();
    created.moveBy(insertPoint.workspaceX, insertPoint.workspaceY);
    created.select();
    setInsertPoint(undefined);
  }

  return (
    <div className="relative h-full min-h-[420px] overflow-hidden">
      <div
        className="absolute inset-0"
        onDoubleClick={(event) => {
          setInsertPoint(insertPointFromDoubleClick(event, workspaceRef.current, readOnly));
        }}
        ref={hostRef}
      />
      {readOnly && <div className="absolute inset-0 z-[5] cursor-not-allowed bg-background/5" />}
      {insertPoint && (
        <LoopQuickInsert
          customDefinitions={customDefinitions}
          messages={messages.quickInsert}
          x={insertPoint.menuX}
          y={insertPoint.menuY}
          onClose={() => setInsertPoint(undefined)}
          onCreateCustom={() => {
            setInsertPoint(undefined);
            onCreateCustom();
          }}
          onInsert={insert}
        />
      )}
    </div>
  );
}
