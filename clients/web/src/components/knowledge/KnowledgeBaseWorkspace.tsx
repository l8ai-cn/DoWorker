"use client";

import type React from "react";
import { useEffect, useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";

import { BlocksDocHeader } from "@/components/blocks/BlocksDocHeader";
import { DocumentView } from "@/components/blocks/DocumentView";
import { SearchPanel } from "@/components/blocks/search/SearchPanel";
import { CenteredSpinner } from "@/components/ui/spinner";
import { blockstoreApi } from "@/lib/api/facade/blockstoreApi";
import { pageDisplayMeta } from "@/lib/blockstore/pageDisplayMeta";
import { useJumpToBlock } from "@/lib/blockstore/useJumpToBlock";
import { useSelectPage } from "@/lib/blockstore/useSelectPage";
import { getErrorMessage } from "@/lib/utils";
import type { Workspace } from "@/lib/viewModels/blockstore";
import { useCurrentOrg } from "@/stores/auth";
import { useBlocks, useBlockstoreStore } from "@/stores/blockstore";
import "@/stores/blockstoreSubscribe";

export function KnowledgeBaseWorkspace() {
  const t = useTranslations();
  const searchParams = useSearchParams();
  const wsParam = searchParams.get("ws");
  const pageParam = searchParams.get("page");
  const currentOrg = useCurrentOrg();
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [searchOpen, setSearchOpen] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const blocks = useBlocks();
  const selectPage = useSelectPage();
  const jumpToBlock = useJumpToBlock();

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setWorkspace(null);
      setError(null);
      try {
        const ws = await resolveWorkspace(wsParam);
        if (cancelled) return;
        setWorkspace(ws);
        useBlockstoreStore.getState().actions.setActiveWorkspaceId(ws.id);
        void useBlockstoreStore.getState().actions.loadTypeDefs(ws.id);
      } catch (e) {
        if (!cancelled) setError(getErrorMessage(e, t("blockstore.loadFailed")));
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [t, wsParam, currentOrg]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setSearchOpen(true);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  const rootId = workspace?.root_block_id ?? null;
  const selectedPageID = pageParam ?? rootId;
  const rootMeta = useMemo(
    () => pageDisplayMeta(rootId ? blocks[rootId] : undefined),
    [rootId, blocks],
  );
  const currentMeta = useMemo(
    () => pageDisplayMeta(selectedPageID ? blocks[selectedPageID] : undefined),
    [selectedPageID, blocks],
  );

  if (error) {
    return (
      <KnowledgeBaseShell>
        <div className="p-6 text-sm text-destructive">{error}</div>
      </KnowledgeBaseShell>
    );
  }
  if (!workspace || !rootId || !selectedPageID) {
    return (
      <KnowledgeBaseShell>
        <CenteredSpinner />
      </KnowledgeBaseShell>
    );
  }

  return (
    <KnowledgeBaseShell>
      <div className="flex min-h-0 flex-1">
        <main className="flex min-w-0 flex-1 flex-col">
        <BlocksDocHeader
          rootTitle={rootMeta.title}
          rootIcon={rootMeta.icon}
          currentTitle={currentMeta.title}
          currentIcon={currentMeta.icon}
          isRoot={selectedPageID === rootId}
          onAddBlock={() => setMenuOpen(true)}
          onNavigateRoot={() => selectPage(rootId)}
        />
        <div className="min-h-0 flex-1 overflow-y-auto">
          <DocumentView
            workspaceID={workspace.id}
            rootBlockID={selectedPageID}
            menuOpen={menuOpen}
            onMenuOpenChange={setMenuOpen}
          />
        </div>
        </main>
        <SearchPanel
          workspaceID={workspace.id}
          open={searchOpen}
          onClose={() => setSearchOpen(false)}
          onJumpToBlock={jumpToBlock}
        />
      </div>
    </KnowledgeBaseShell>
  );
}

function KnowledgeBaseShell({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-full min-h-0 w-full flex-col bg-background">
      <header className="flex h-14 shrink-0 items-center border-b border-border px-5">
        <div>
          <h1 className="text-sm font-semibold text-foreground">知识库</h1>
          <p className="text-xs text-muted-foreground">维护可挂载到 Pod 的团队资料</p>
        </div>
      </header>
      {children}
    </div>
  );
}

async function resolveWorkspace(wsParam: string | null): Promise<Workspace> {
  if (wsParam) {
    const list = await blockstoreApi.listWorkspaces();
    const found = list.workspaces.find((w) => w.id === wsParam);
    if (found) return found;
  }
  return useBlockstoreStore.getState().actions.ensureDefaultWorkspace();
}
