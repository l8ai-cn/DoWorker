import type { ReactNode } from "react";

interface BlockProgrammingWorkbenchMessages {
  canvasHint: string;
  canvasTitle: string;
  editorMetadata: string;
  editorTitle: string;
}

interface BlockProgrammingWorkbenchProps {
  canvas: ReactNode;
  editor: ReactNode;
  messages: BlockProgrammingWorkbenchMessages;
  status: ReactNode;
  toolbar: ReactNode;
  loading?: boolean;
  loadingFallback?: ReactNode;
}

export function BlockProgrammingWorkbench({
  canvas,
  editor,
  messages,
  status,
  toolbar,
  loading = false,
  loadingFallback,
}: BlockProgrammingWorkbenchProps) {
  return (
    <div className="flex h-full min-h-0 flex-col bg-background">
      {toolbar}
      {loading ? (
        <div className="flex flex-1 items-center justify-center">
          {loadingFallback}
        </div>
      ) : (
        <main className="grid min-h-0 flex-1 grid-cols-1 overflow-auto xl:grid-cols-[minmax(0,1.35fr)_minmax(380px,0.65fr)] xl:overflow-hidden">
          <section className="min-h-[520px] border-b border-border xl:min-h-0 xl:border-b-0 xl:border-r">
            <div className="flex h-10 items-center justify-between border-b border-border bg-surface px-4">
              <h2 className="text-xs font-semibold uppercase text-muted-foreground">
                {messages.canvasTitle}
              </h2>
              <span className="text-xs text-muted-foreground">{messages.canvasHint}</span>
            </div>
            <div className="h-[calc(100%-2.5rem)]">{canvas}</div>
          </section>
          <aside className="grid min-h-[620px] grid-rows-[minmax(360px,1fr)_auto] xl:min-h-0">
            <section className="min-h-0">
              <div className="flex h-10 items-center justify-between border-b border-border bg-surface px-4">
                <h2 className="text-xs font-semibold text-muted-foreground">
                  {messages.editorTitle}
                </h2>
                <span className="text-[11px] text-muted-foreground">
                  {messages.editorMetadata}
                </span>
              </div>
              <div className="h-[calc(100%-2.5rem)]">{editor}</div>
            </section>
            {status}
          </aside>
        </main>
      )}
    </div>
  );
}
