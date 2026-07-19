"use client";

import * as Blockly from "blockly";
import { useEffect, useRef, useState } from "react";
import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";
import { setLoopBlockAccess } from "./loop-block-access";
import { insertPointFromDoubleClick, type LoopBlockInsertPoint } from "./loop-block-insert-point";
import { insertLoopBlock } from "./loop-block-insertion";
import { loopBlockProgrammingHostAdapter } from "./loop-block-programming-host-adapter";
import { loopBlocklyTheme } from "./loop-blockly-theme";
import type { LoopCustomBlockDefinition } from "./loop-custom-block-types";
import { LoopQuickInsert } from "./loop-quick-insert";
import type { LoopBlockCatalogMessages, LoopQuickInsertMessages } from "./loop-workbench-messages";

interface LoopBlocklyCanvasProps {
  program?: LoopProgram;
  semanticRevision: number;
  readOnly: boolean;
  customDefinitions: readonly LoopCustomBlockDefinition[];
  messages: { blockly: LoopBlockCatalogMessages; quickInsert: LoopQuickInsertMessages };
  onCreateCustom: () => void;
  onSourceChange: (source: string) => void;
  onSelectNode: (nodeId?: string) => void;
}

export function LoopBlocklyCanvas(props: LoopBlocklyCanvasProps) {
  const {
    program, semanticRevision, readOnly, customDefinitions, onSourceChange, onCreateCustom, onSelectNode, messages,
  } = props;
  const hostRef = useRef<HTMLDivElement>(null);
  const workspaceRef = useRef<Blockly.WorkspaceSvg | undefined>(undefined);
  const programRef = useRef(program);
  const readOnlyRef = useRef(readOnly);
  const customDefinitionsRef = useRef(customDefinitions);
  const customInsertPointRef = useRef<LoopBlockInsertPoint | undefined>(undefined);
  const customCountRef = useRef(customDefinitions.length);
  const callbackRef = useRef({ onSourceChange, onSelectNode });
  const [insertPoint, setInsertPoint] = useState<LoopBlockInsertPoint>();

  useEffect(() => {
    callbackRef.current = { onSourceChange, onSelectNode };
    programRef.current = program;
    readOnlyRef.current = readOnly;
    customDefinitionsRef.current = customDefinitions;
  }, [customDefinitions, onSelectNode, onSourceChange, program, readOnly]);

  useEffect(() => {
    if (!hostRef.current) return;
    loopBlockProgrammingHostAdapter.registerBlocks(messages.blockly, customDefinitionsRef.current);
    const { toolbox } = loopBlockProgrammingHostAdapter.createCatalog(
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
          selected instanceof Blockly.Block
            ? loopBlockProgrammingHostAdapter.nodeIdForBlock(selected)
            : undefined,
        );
      }
      if (!event.isUiEvent) {
        callbackRef.current.onSourceChange(
          loopBlockProgrammingHostAdapter
            .workspaceToSource(workspace, customDefinitionsRef.current)
            .source,
        );
      }
    };
    workspace.addChangeListener(listener);
    const currentProgram = programRef.current;
    if (currentProgram) {
      loopBlockProgrammingHostAdapter.projectProgram(
        workspace,
        currentProgram,
        customDefinitionsRef.current,
      );
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
      loopBlockProgrammingHostAdapter.projectProgram(
        workspace,
        currentProgram,
        customDefinitionsRef.current,
      );
    }
  }, [semanticRevision]);

  useEffect(() => {
    const workspace = workspaceRef.current;
    if (!workspace) return;
    loopBlockProgrammingHostAdapter.registerBlocks(messages.blockly, customDefinitions);
    workspace.updateToolbox(
      loopBlockProgrammingHostAdapter.createCatalog(messages.blockly, customDefinitions).toolbox,
    );
    const createdDefinition =
      customDefinitions.length > customCountRef.current ? customDefinitions.at(-1) : undefined;
    customCountRef.current = customDefinitions.length;
    const customInsertPoint = customInsertPointRef.current;
    if (createdDefinition && customInsertPoint) {
      insertLoopBlock({
        workspace,
        type: loopBlockProgrammingHostAdapter.customBlockType(createdDefinition),
        insertPoint: customInsertPoint,
        customDefinitions,
      });
      customInsertPointRef.current = undefined;
    }
    callbackRef.current.onSourceChange(
      loopBlockProgrammingHostAdapter.workspaceToSource(workspace, customDefinitions).source,
    );
  }, [customDefinitions, messages.blockly]);

  useEffect(() => {
    const workspace = workspaceRef.current;
    if (!workspace) return;
    setLoopBlockAccess(workspace, readOnly);
  }, [readOnly, semanticRevision]);

  function insert(type: string) {
    const workspace = workspaceRef.current;
    if (!workspace || !insertPoint || readOnly) return;
    insertLoopBlock({
      workspace,
      type,
      insertPoint,
      customDefinitions: customDefinitionsRef.current,
    });
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
            customInsertPointRef.current = insertPoint;
            setInsertPoint(undefined);
            onCreateCustom();
          }}
          onInsert={insert}
        />
      )}
    </div>
  );
}
