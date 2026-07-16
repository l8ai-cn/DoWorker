"use client";

import * as Blockly from "blockly";
import { useEffect, useRef, useState } from "react";
import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";
import { createLoopBlockCatalog, registerLoopBlocks } from "./loop-block-catalog";
import {
  nodeIdForBlock,
  projectProgramToWorkspace,
  workspaceToLoopSource,
} from "./loop-block-projection";
import { LoopQuickInsert } from "./loop-quick-insert";
import type {
  LoopBlockCatalogMessages,
  LoopQuickInsertMessages,
} from "./loop-workbench-messages";

interface InsertPoint {
  menuX: number;
  menuY: number;
  workspaceX: number;
  workspaceY: number;
}

interface LoopBlocklyCanvasProps {
  program?: LoopProgram;
  semanticRevision: number;
  readOnly: boolean;
  messages: {
    blockly: LoopBlockCatalogMessages;
    quickInsert: LoopQuickInsertMessages;
  };
  onSourceChange: (source: string) => void;
  onSelectNode: (nodeId?: string) => void;
}

const loopTheme = Blockly.Theme.defineTheme("loop", {
  name: "loop",
  base: Blockly.Themes.Classic,
  componentStyles: {
    workspaceBackgroundColour: "#f8fafc",
    toolboxBackgroundColour: "#ffffff",
    toolboxForegroundColour: "#334155",
    flyoutBackgroundColour: "#f1f5f9",
    flyoutForegroundColour: "#334155",
    scrollbarColour: "#94a3b8",
    insertionMarkerColour: "#0f766e",
    insertionMarkerOpacity: 0.35,
  },
});

export function LoopBlocklyCanvas({
  program,
  semanticRevision,
  readOnly,
  onSourceChange,
  onSelectNode,
  messages,
}: LoopBlocklyCanvasProps) {
  const hostRef = useRef<HTMLDivElement>(null);
  const workspaceRef = useRef<Blockly.WorkspaceSvg | undefined>(undefined);
  const programRef = useRef(program);
  const readOnlyRef = useRef(readOnly);
  const callbackRef = useRef({ onSourceChange, onSelectNode });
  const [insertPoint, setInsertPoint] = useState<InsertPoint>();

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
    if (!hostRef.current) return;
    registerLoopBlocks(messages.blockly);
    const { toolbox } = createLoopBlockCatalog(messages.blockly);
    const compact = hostRef.current.clientWidth < 640;
    const workspace = Blockly.inject(hostRef.current, {
      toolbox,
      theme: loopTheme,
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
        callbackRef.current.onSourceChange(workspaceToLoopSource(workspace).source);
      }
    };
    workspace.addChangeListener(listener);
    const currentProgram = programRef.current;
    if (currentProgram) {
      projectProgramToWorkspace(workspace, currentProgram);
      setBlockAccess(workspace, readOnlyRef.current);
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
      projectProgramToWorkspace(workspace, currentProgram);
    }
  }, [semanticRevision]);

  useEffect(() => {
    const workspace = workspaceRef.current;
    if (!workspace) return;
    setBlockAccess(workspace, readOnly);
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
          const workspace = workspaceRef.current;
          const target = event.target;
          if (readOnly || !workspace || !(target instanceof Element) ||
              !target.classList.contains("blocklyMainBackground")) return;
          const bounds = event.currentTarget.getBoundingClientRect();
          const point = Blockly.utils.svgMath.screenToWsCoordinates(
            workspace,
            new Blockly.utils.Coordinate(event.clientX, event.clientY),
          );
          setInsertPoint({
            menuX: event.clientX - bounds.left,
            menuY: event.clientY - bounds.top,
            workspaceX: point.x,
            workspaceY: point.y,
          });
        }}
        ref={hostRef}
      />
      {readOnly && <div className="absolute inset-0 z-[5] cursor-not-allowed bg-background/5" />}
      {insertPoint && (
        <LoopQuickInsert
          messages={messages.quickInsert}
          x={insertPoint.menuX}
          y={insertPoint.menuY}
          onClose={() => setInsertPoint(undefined)}
          onInsert={insert}
        />
      )}
    </div>
  );
}

function setBlockAccess(workspace: Blockly.WorkspaceSvg, readOnly: boolean) {
  for (const block of workspace.getAllBlocks(false)) {
    block.setEditable(!readOnly);
    block.setMovable(!readOnly);
    block.setDeletable(!readOnly);
  }
}
