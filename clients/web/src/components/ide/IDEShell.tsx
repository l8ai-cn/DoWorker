"use client";

import React, { useState, useCallback } from "react";
import { usePathname, useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { hideIdeChrome, hideIdeSidebar } from "@/lib/ide-chrome";
import { activityHasSidebar } from "@/lib/ide-sidebar";
import { resolveActivityFromPathname } from "@/lib/ide-route";
import { useCtaModal } from "@/hooks/useCtaModal";
import { CenteredSpinner } from "@/components/ui/spinner";
import { ActivityBar } from "./ActivityBar";
import { SideBar } from "./SideBar";
import { BottomPanel } from "./BottomPanel";
import { CommandPalette } from "./CommandPalette";
import { CreatePodModal } from "./CreatePodModal";
import { getSidebarContent, type SidebarCallbacks } from "./ideSidebarContent";
import { useIDEStore } from "@/stores/ide";
import { useWorkspaceStore } from "@/stores/workspace";
import { usePodStore } from "@/stores/pod";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { getPodDisplayName } from "@/lib/pod-display-name";
import { AddRunnerModal } from "./modals/AddRunnerModal";
import { ImportRepositoryModal } from "./modals/ImportRepositoryModal";
import { useCurrentOrg } from "@/stores/auth";
import { useIDEStoreHydration } from "@/hooks/useIDEStoreHydration";

interface IDEShellProps {
  children: React.ReactNode;
  sidebarContent?: React.ReactNode;
  className?: string;
}

export function IDEShell({
  children,
  sidebarContent,
  className,
}: IDEShellProps) {
  const pathname = usePathname();
  const noSidebar = hideIdeSidebar(pathname);
  const noChrome = hideIdeChrome(pathname);
  const bottomPanelOpen = useIDEStore((state) => state.bottomPanelOpen);
  const activeActivity = useIDEStore((state) => state.activeActivity);
  const routeActivity = resolveActivityFromPathname(pathname);
  const sidebarActivity = routeActivity ?? activeActivity;
  const _hasHydrated = useIDEStoreHydration();
  const addPane = useWorkspaceStore((state) => state.addPane);
  const fetchPods = usePodStore((state) => state.fetchPods);
  const t = useTranslations();
  const router = useRouter();
  const currentOrg = useCurrentOrg();
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const [createPodModalOpen, setCreatePodModalOpen] = useState(false);
  const addRunnerModal = useCtaModal();
  const importRepoModal = useCtaModal();

  const handleCreatePod = useCallback(() => {
    const orgSlug = currentOrg?.slug;
    if (orgSlug) {
      router.push(`/${orgSlug}/workers/new`);
      return;
    }
    setCreatePodModalOpen(true);
  }, [currentOrg?.slug, router, setCreatePodModalOpen]);

  const handlePodCreated = useCallback((pod?: { pod_key: string; title?: string }) => {
    setCreatePodModalOpen(false);
    if (pod?.pod_key) {
      const displayName = getPodDisplayName(pod);
      toast.info(t("workspace.podCreated"), {
        description: displayName,
      });
      addPane(pod.pod_key);
      fetchPods();
    }
  }, [addPane, fetchPods, t, setCreatePodModalOpen]);

  const sidebarCallbacks: SidebarCallbacks = {
    onCreatePod: handleCreatePod,
    onAddRunner: addRunnerModal.open,
    onImportRepo: importRepoModal.open,
  };
  const effectiveSidebarContent =
    sidebarContent ?? getSidebarContent(sidebarActivity, sidebarCallbacks);
  const showSidebar =
    !noSidebar &&
    effectiveSidebarContent != null &&
    activityHasSidebar(sidebarActivity);

  if (!_hasHydrated) {
    return (
      <div className="h-screen bg-background">
        <CenteredSpinner />
      </div>
    );
  }

  if (noChrome) {
    return (
      <div className={cn("app-shell flex h-screen flex-col bg-background overflow-hidden", className)}>
        <main className="flex-1 min-h-0 overflow-hidden">{children}</main>
      </div>
    );
  }

  return (
    <div className={cn("app-shell flex h-screen bg-background overflow-hidden", className)}>
      <ActivityBar className="flex-shrink-0" />

      {!showSidebar ? null : (
        <SideBar className="flex-shrink-0">{effectiveSidebarContent}</SideBar>
      )}

      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <main
          className={cn(
            "flex-1 overflow-auto",
            activeActivity === "workspace" && bottomPanelOpen ? "" : "pb-8"
          )}
        >
          {children}
        </main>

        {activeActivity === "workspace" && <BottomPanel />}
      </div>

      <CommandPalette
        open={commandPaletteOpen}
        onOpenChange={setCommandPaletteOpen}
      />

      <CreatePodModal
        open={createPodModalOpen}
        onClose={() => setCreatePodModalOpen(false)}
        onCreated={handlePodCreated}
      />

      <AddRunnerModal
        open={addRunnerModal.isOpen}
        onClose={addRunnerModal.close}
        onCreated={addRunnerModal.commit}
      />

      <ImportRepositoryModal
        open={importRepoModal.isOpen}
        onClose={importRepoModal.close}
        onImported={importRepoModal.commit}
      />
    </div>
  );
}

export default IDEShell;
