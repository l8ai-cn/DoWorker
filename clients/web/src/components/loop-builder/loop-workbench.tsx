"use client";

import { useLocale } from "next-intl";
import { useState } from "react";
import { BlockProgrammingWorkbench } from "@/components/block-programming/BlockProgrammingWorkbench";
import { Spinner } from "@/components/ui/spinner";
import { LoopBlocklyCanvas } from "./loop-blockly-canvas";
import { LoopAIAssistantDialog } from "./loop-ai-assistant-dialog";
import { LoopCodeEditor } from "./loop-code-editor";
import { LoopCustomBlockDialog } from "./loop-custom-block-dialog";
import { LoopRuntimeDialog } from "./loop-runtime-dialog";
import { LoopStatusPanel } from "./loop-status-panel";
import { LoopWorkbenchToolbar } from "./loop-workbench-toolbar";
import { useLoopWorkbenchMessages } from "./loop-workbench-messages";
import { useLoopAIAssistant } from "./use-loop-ai-assistant";
import { useLoopCustomBlockLibrary } from "./use-loop-custom-block-library";
import { useLoopWorkbench } from "./use-loop-workbench";

export function LoopWorkbench({ orgSlug }: { orgSlug: string }) {
  const messages = useLoopWorkbenchMessages();
  const model = useLoopWorkbench(orgSlug, messages.errors);
  const locale = useLocale();
  const ai = useLoopAIAssistant({
    orgSlug,
    locale,
    snapshot: model.snapshot,
    messages: messages.ai,
    onApplied: model.applySnapshot,
  });
  const [selectedNodeId, setSelectedNodeId] = useState<string>();
  const customBlocks = useLoopCustomBlockLibrary();
  const [customDialogOpen, setCustomDialogOpen] = useState(false);
  const [runtimeDialogOpen, setRuntimeDialogOpen] = useState(false);
  const blocksWritable =
    model.snapshot.activeEditor === "blocks" &&
    model.snapshot.parseStatus === "valid" &&
    !model.running;
  const codeWritable =
    model.snapshot.activeEditor === "code" && !model.running;

  return (
    <>
      <BlockProgrammingWorkbench
        canvas={(
          <LoopBlocklyCanvas
            customDefinitions={customBlocks.definitions}
            messages={{ blockly: messages.blockly, quickInsert: messages.quickInsert }}
            onCreateCustom={() => setCustomDialogOpen(true)}
            onSelectNode={setSelectedNodeId}
            onSourceChange={model.updateBlocks}
            program={model.snapshot.program}
            readOnly={!blocksWritable}
            semanticRevision={model.snapshot.semanticRevision}
          />
        )}
        editor={(
          <LoopCodeEditor
            onChange={model.updateCode}
            readOnly={!codeWritable}
            value={model.snapshot.source}
          />
        )}
        loading={model.loading && !model.snapshot.source}
        loadingFallback={<Spinner />}
        messages={{
          canvasTitle: messages.shell.canvasTitle,
          canvasHint: messages.shell.canvasHint,
          editorTitle: messages.shell.editorTitle,
          editorMetadata: messages.shell.editorMetadata(
            model.snapshot.revision,
            model.snapshot.semanticRevision,
          ),
        }}
        status={(
          <LoopStatusPanel
            error={model.error ?? customBlocks.error}
            messages={messages.status}
            onRepairDiagnostic={(diagnostic) => ai.openRepair({
              diagnosticCode: diagnostic.code,
              diagnosticLabel: messages.status.diagnosticLabel(diagnostic.code),
              nodeId: diagnostic.nodeId,
              fieldPath: diagnostic.fieldPath,
            })}
            repairingTarget={ai.busy ? ai.repairTarget : undefined}
            selectedNodeId={selectedNodeId}
            snapshot={model.snapshot}
          />
        )}
        toolbar={(
          <LoopWorkbenchToolbar
            aiLabel={messages.ai.toolbar}
            editor={model.snapshot.activeEditor}
            messages={messages.toolbar}
            onAI={() => ai.setOpen(true)}
            onEditorChange={model.setEditor}
            onRun={() => setRuntimeDialogOpen(true)}
            orgSlug={orgSlug}
            parseStatus={model.snapshot.parseStatus}
            running={model.running}
          />
        )}
      />
      <LoopRuntimeDialog
        error={model.runtimeError}
        loading={model.runtimeLoading}
        messages={messages.runtime}
        open={runtimeDialogOpen}
        running={model.running}
        templates={model.runtimeTemplates}
        onOpenChange={setRuntimeDialogOpen}
        onRetry={model.retryRuntimeLoad}
        onRun={(templateName) => {
          setRuntimeDialogOpen(false);
          void model.run(templateName);
        }}
      />
      <LoopCustomBlockDialog
        definitions={customBlocks.definitions}
        messages={messages.customBlock}
        open={customDialogOpen}
        onCreate={customBlocks.create}
        onOpenChange={setCustomDialogOpen}
      />
      <LoopAIAssistantDialog
        busy={ai.busy}
        messages={messages.ai}
        mode={ai.mode}
        open={ai.open}
        parseStatus={model.snapshot.parseStatus}
        prompt={ai.prompt}
        program={model.snapshot.program}
        proposal={ai.proposal}
        repairTarget={ai.repairTarget}
        requestError={ai.requestError}
        resourceError={ai.resourceError}
        resources={ai.resources}
        resourcesLoading={ai.resourcesLoading}
        selectedResourceId={ai.selectedResourceId}
        onBack={ai.back}
        onConfirm={() => void ai.confirm()}
        onModeChange={ai.setMode}
        onOpenChange={ai.setOpen}
        onPromptChange={ai.setPrompt}
        onResourceChange={ai.setSelectedResourceId}
        onRetryResources={ai.retryResources}
        onSubmit={() => void ai.submit()}
      />
    </>
  );
}
