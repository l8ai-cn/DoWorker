import { Plus } from "lucide-react";
import { useState } from "react";

import { LOOP_BLOCK_TYPES } from "../blockly/block-catalog";
import {
  createGoalRoot,
  insertLoopBlock,
  loadExampleProgram,
} from "../blockly/workspace-seeds";
import { useLoopWorkspace } from "../hooks/use-loop-workspace";
import { useLoopSimulation } from "../hooks/use-loop-simulation";
import { BlockInspector } from "./block-inspector";
import { BlocklyCanvas } from "./blockly-canvas";
import { CustomBlockDialog } from "./custom-block-dialog";
import { ConfirmationDialog } from "./confirmation-dialog";
import { OutputPanel } from "./output-panel";
import { QuickInsertMenu } from "./quick-insert-menu";
import { WorkbenchToolbar } from "./workbench-toolbar";

export function LoopWorkbench() {
  const model = useLoopWorkspace();
  const [quickPoint, setQuickPoint] = useState<{
    menuX: number;
    menuY: number;
    workspaceX: number;
    workspaceY: number;
  }>();
  const [customDialogOpen, setCustomDialogOpen] = useState(false);
  const [exampleConfirmOpen, setExampleConfirmOpen] = useState(false);
  const simulation = useLoopSimulation(model.workspace, model.compileResult);
  const hasBlocks = (model.workspace?.getAllBlocks(false).length ?? 0) > 0;
  const hasRoot = (model.workspace?.getBlocksByType(
    LOOP_BLOCK_TYPES.root,
    false,
  ).length ?? 0) > 0;
  const valid = Boolean(model.compileResult.program);

  const showDiagnostics = () => {
    model.setOutputTab("diagnostics");
    const first = model.compileResult.diagnostics.find(({ blockId }) => blockId);
    if (first?.blockId) model.focusBlock(first.blockId);
  };

  const startSimulation = () => {
    if (!model.workspace || !valid) {
      showDiagnostics();
      return;
    }
    setQuickPoint(undefined);
    model.setOutputTab("evidence");
    void simulation.start();
  };

  const insertAtQuickPoint = (type: string) => {
    if (!model.workspace || !quickPoint) return;
    insertLoopBlock(
      model.workspace,
      type,
      quickPoint.workspaceX,
      quickPoint.workspaceY,
    );
    setQuickPoint(undefined);
  };

  return (
    <main className="loop-app">
      <WorkbenchToolbar
        dirty={model.dirty}
        onGenerate={() => model.setOutputTab("json")}
        onLoadExample={() => {
          if (model.dirty && hasBlocks) {
            setExampleConfirmOpen(true);
          } else if (model.workspace) {
            loadExampleProgram(model.workspace);
          }
        }}
        onOpenCustom={() => setCustomDialogOpen(true)}
        onRun={startSimulation}
        onSave={model.save}
        onStop={simulation.stop}
        onValidate={showDiagnostics}
        problemCount={model.compileResult.diagnostics.length}
        running={simulation.running}
        valid={valid}
        workspaceReady={Boolean(model.workspace)}
      />

      {model.loadError && (
        <div className="load-error">
          <span>{model.loadError}</span>
          <button onClick={model.resetStoredProject} type="button">清除本地项目</button>
        </div>
      )}
      {simulation.error && (
        <div className="load-error"><span>{simulation.error}</span></div>
      )}

      <div className="workbench-grid">
        <section className="canvas-shell">
          <BlocklyCanvas
            customDefinitions={model.customDefinitions}
            initialState={model.initialWorkspaceState}
            onCanvasDoubleClick={setQuickPoint}
            onChange={model.handleChange}
            onLoadError={model.setLoadError}
            onReady={model.handleReady}
            onSelectionChange={model.setSelectedBlock}
          />
          {!hasBlocks && model.workspace && (
            <button
              className="empty-canvas-action"
              onClick={() => createGoalRoot(model.workspace!)}
              type="button"
            >
              <Plus size={18} /> 创建 Goal Loop
            </button>
          )}
          {simulation.running && <div className="running-shield" />}
          {quickPoint && (
            <QuickInsertMenu
              customDefinitions={model.customDefinitions}
              hasRoot={hasRoot}
              onClose={() => setQuickPoint(undefined)}
              onCreateCustom={() => {
                setQuickPoint(undefined);
                setCustomDialogOpen(true);
              }}
              onInsert={insertAtQuickPoint}
              point={{ x: quickPoint.menuX, y: quickPoint.menuY }}
            />
          )}
        </section>
        <BlockInspector
          block={model.selectedBlock}
          customDefinitions={model.customDefinitions}
          diagnostics={model.compileResult.diagnostics}
          disabled={simulation.running}
        />
        <OutputPanel
          compileResult={model.compileResult}
          evidence={simulation.evidence}
          onFocusBlock={model.focusBlock}
          onTabChange={model.setOutputTab}
          tab={model.outputTab}
        />
      </div>
      <CustomBlockDialog
        onClose={() => setCustomDialogOpen(false)}
        onCreate={model.addCustomDefinition}
        open={customDialogOpen}
      />
      <ConfirmationDialog
        body="当前未保存的工作区会被示例替换。"
        confirmLabel="载入示例"
        onClose={() => setExampleConfirmOpen(false)}
        onConfirm={() => model.workspace &&
          loadExampleProgram(model.workspace)}
        open={exampleConfirmOpen}
        title="替换当前工作区？"
      />
    </main>
  );
}
