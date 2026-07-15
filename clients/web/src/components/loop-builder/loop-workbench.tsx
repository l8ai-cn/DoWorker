"use client";

import { useState } from "react";
import { Spinner } from "@/components/ui/spinner";
import { LoopBlocklyCanvas } from "./loop-blockly-canvas";
import { LoopCodeEditor } from "./loop-code-editor";
import { LoopRuntimeDialog } from "./loop-runtime-dialog";
import { LoopStatusPanel } from "./loop-status-panel";
import { LoopWorkbenchToolbar } from "./loop-workbench-toolbar";
import { useLoopWorkbench } from "./use-loop-workbench";

export function LoopWorkbench({ orgSlug }: { orgSlug: string }) {
  const model = useLoopWorkbench(orgSlug);
  const [selectedNodeId, setSelectedNodeId] = useState<string>();
  const [runtimeDialogOpen, setRuntimeDialogOpen] = useState(false);
  const blocksWritable =
    model.snapshot.activeEditor === "blocks" &&
    model.snapshot.parseStatus === "valid" &&
    !model.running;
  const codeWritable =
    model.snapshot.activeEditor === "code" && !model.running;

  return (
    <div className="flex h-full min-h-0 flex-col bg-background">
      <LoopWorkbenchToolbar
        editor={model.snapshot.activeEditor}
        onEditorChange={model.setEditor}
        onRun={() => setRuntimeDialogOpen(true)}
        orgSlug={orgSlug}
        parseStatus={model.snapshot.parseStatus}
        running={model.running}
      />
      {model.loading && !model.snapshot.source ? (
        <div className="flex flex-1 items-center justify-center">
          <Spinner />
        </div>
      ) : (
        <main className="grid min-h-0 flex-1 grid-cols-1 overflow-auto xl:grid-cols-[minmax(0,1.35fr)_minmax(380px,0.65fr)] xl:overflow-hidden">
          <section className="min-h-[520px] border-b border-border xl:min-h-0 xl:border-b-0 xl:border-r">
            <div className="flex h-10 items-center justify-between border-b border-border bg-surface px-4">
              <h2 className="text-xs font-semibold uppercase text-muted-foreground">积木画布</h2>
              <span className="text-xs text-muted-foreground">
                双击空白处快速插入
              </span>
            </div>
            <div className="h-[calc(100%-2.5rem)]">
              <LoopBlocklyCanvas
                onSelectNode={setSelectedNodeId}
                onSourceChange={model.updateBlocks}
                program={model.snapshot.program}
                readOnly={!blocksWritable}
                semanticRevision={model.snapshot.semanticRevision}
              />
            </div>
          </section>
          <aside className="grid min-h-[620px] grid-rows-[minmax(360px,1fr)_auto] xl:min-h-0">
            <section className="min-h-0">
              <div className="flex h-10 items-center justify-between border-b border-border bg-surface px-4">
                <h2 className="text-xs font-semibold text-muted-foreground">循环脚本</h2>
                <span className="text-[11px] text-muted-foreground">
                  编辑版本 {model.snapshot.revision} · 语义版本 {model.snapshot.semanticRevision}
                </span>
              </div>
              <div className="h-[calc(100%-2.5rem)]">
                <LoopCodeEditor
                  onChange={model.updateCode}
                  readOnly={!codeWritable}
                  value={model.snapshot.source}
                />
              </div>
            </section>
            <LoopStatusPanel
              error={model.error}
              selectedNodeId={selectedNodeId}
              snapshot={model.snapshot}
            />
          </aside>
        </main>
      )}
      <LoopRuntimeDialog
        error={model.runtimeError}
        loading={model.runtimeLoading}
        open={runtimeDialogOpen}
        running={model.running}
        snapshots={model.runtimeSnapshots}
        onOpenChange={setRuntimeDialogOpen}
        onRetry={model.retryRuntimeLoad}
        onRun={(snapshotId) => {
          setRuntimeDialogOpen(false);
          void model.run(snapshotId);
        }}
      />
    </div>
  );
}
