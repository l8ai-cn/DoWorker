import * as Blockly from "blockly";
import { useEffect, useRef } from "react";

import { registerLoomBlocks } from "../blockly/block-catalog";
import { loomTheme } from "../blockly/block-theme";
import { createLoomToolbox } from "../blockly/toolbox";
import {
  registerCustomBlock,
  type CustomBlockDefinition,
} from "../custom-blocks/custom-block-definition";

interface CanvasPoint {
  menuX: number;
  menuY: number;
  workspaceX: number;
  workspaceY: number;
}

interface BlocklyCanvasProps {
  customDefinitions: CustomBlockDefinition[];
  initialState?: Record<string, unknown>;
  onCanvasDoubleClick: (point: CanvasPoint) => void;
  onChange: (workspace: Blockly.WorkspaceSvg) => void;
  onLoadError: (message: string) => void;
  onReady: (
    workspace: Blockly.WorkspaceSvg,
    loadedSuccessfully: boolean,
  ) => void;
  onSelectionChange: (block: Blockly.BlockSvg | null) => void;
}

export function BlocklyCanvas({
  customDefinitions,
  initialState,
  onCanvasDoubleClick,
  onChange,
  onLoadError,
  onReady,
  onSelectionChange,
}: BlocklyCanvasProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const workspaceRef = useRef<Blockly.WorkspaceSvg | null>(null);
  const callbacksRef = useRef({
    onChange,
    onLoadError,
    onReady,
    onSelectionChange,
  });
  callbacksRef.current = { onChange, onLoadError, onReady, onSelectionChange };

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;
    registerLoomBlocks();
    try {
      customDefinitions.forEach(registerCustomBlock);
    } catch (error) {
      callbacksRef.current.onLoadError(
        error instanceof Error ? error.message : "自定义积木注册失败。",
      );
      return;
    }
    const workspace = Blockly.inject(container, {
      toolbox: createLoomToolbox(customDefinitions),
      theme: loomTheme,
      renderer: "zelos",
      trashcan: true,
      grid: { spacing: 20, length: 1, colour: "#d7dce2", snap: true },
      zoom: {
        controls: true,
        wheel: true,
        startScale: 0.82,
        maxScale: 1.35,
        minScale: 0.45,
        scaleSpeed: 1.1,
      },
      move: { scrollbars: true, drag: true, wheel: true },
    });
    workspaceRef.current = workspace;
    let loadedSuccessfully = true;
    if (initialState) {
      try {
        Blockly.serialization.workspaces.load(initialState, workspace);
      } catch (error) {
        loadedSuccessfully = false;
        callbacksRef.current.onLoadError(
          error instanceof Error ? error.message : "工作区加载失败。",
        );
      }
    }
    const listener = (event: Blockly.Events.Abstract) => {
      if (event.type === Blockly.Events.SELECTED) {
        const selected = Blockly.getSelected();
        callbacksRef.current.onSelectionChange(
          selected instanceof Blockly.BlockSvg ? selected : null,
        );
      }
      if (!event.isUiEvent) callbacksRef.current.onChange(workspace);
    };
    workspace.addChangeListener(listener);
    const observer = new ResizeObserver(() => Blockly.svgResize(workspace));
    observer.observe(container);
    callbacksRef.current.onReady(workspace, loadedSuccessfully);
    return () => {
      observer.disconnect();
      workspace.removeChangeListener(listener);
      workspace.dispose();
      workspaceRef.current = null;
    };
  }, []);

  useEffect(() => {
    const workspace = workspaceRef.current;
    if (!workspace) return;
    try {
      customDefinitions.forEach(registerCustomBlock);
    } catch (error) {
      callbacksRef.current.onLoadError(
        error instanceof Error ? error.message : "自定义积木注册失败。",
      );
      return;
    }
    workspace.updateToolbox(createLoomToolbox(customDefinitions));
  }, [customDefinitions]);

  return (
    <div
      className="blockly-canvas"
      ref={containerRef}
      onDoubleClick={(event) => {
        const target = event.target;
        const workspace = workspaceRef.current;
        if (
          !workspace ||
          !(target instanceof Element) ||
          !target.classList.contains("blocklyMainBackground")
        ) {
          return;
        }
        const bounds = event.currentTarget.getBoundingClientRect();
        const workspacePoint = Blockly.utils.svgMath.screenToWsCoordinates(
          workspace,
          new Blockly.utils.Coordinate(event.clientX, event.clientY),
        );
        onCanvasDoubleClick({
          menuX: event.clientX - bounds.left,
          menuY: event.clientY - bounds.top,
          workspaceX: workspacePoint.x,
          workspaceY: workspacePoint.y,
        });
      }}
    />
  );
}
