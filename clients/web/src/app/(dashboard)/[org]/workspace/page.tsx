"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { toast } from "sonner";
import { useWorkspaceStore } from "@/stores/workspace";
import { usePodStore } from "@/stores/pod";
import { WorkspaceManager } from "@/components/workspace";
import { WorkspaceEmptyState } from "@/components/workspace/WorkspaceEmptyState";
import { CenteredSpinner } from "@/components/ui/spinner";
import { useTranslations } from "next-intl";
import { CreatePodModal } from "@/components/ide/CreatePodModal";
import { getShortPodKey } from "@/lib/pod-display-name";
import type { PodData } from "@/lib/api";
import type { WorkspaceRecipeSelection } from "@/components/workspace/workspace-recipes";

export default function WorkspacePage() {
  const t = useTranslations();
  const searchParams = useSearchParams();
  const router = useRouter();
  const panes = useWorkspaceStore((s) => s.panes);
  const addPane = useWorkspaceStore((s) => s.addPane);
  const _hasHydrated = useWorkspaceStore((s) => s._hasHydrated);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [recipe, setRecipe] = useState<WorkspaceRecipeSelection | null>(null);
  const processedPodRef = useRef<string | null>(null);

  const handleCreatePod = useCallback((selection?: WorkspaceRecipeSelection) => {
    setRecipe(selection ?? null);
    setShowCreateModal(true);
  }, []);

  const handleCloseCreate = useCallback(() => {
    setShowCreateModal(false);
    setRecipe(null);
  }, []);

  const handleOpenPod = useCallback((podKey: string) => {
    addPane(podKey);
  }, [addPane]);

  const handlePodCreated = useCallback((pod?: PodData) => {
    setShowCreateModal(false);
    if (!pod?.pod_key) return;

    toast.info(t("workspace.podCreated"), {
      description: `Pod: ${getShortPodKey(pod.pod_key)}`,
    });
    handleOpenPod(pod.pod_key);

    usePodStore.getState().upsertPod(pod);
  }, [t, handleOpenPod]);

  useEffect(() => {
    if (!_hasHydrated) return;

    const podKey = searchParams.get("pod");
    if (podKey && podKey !== processedPodRef.current) {
      processedPodRef.current = podKey;
      const isAlreadyOpen = panes.some((p) => p.podKey === podKey);
      if (!isAlreadyOpen) {
        handleOpenPod(podKey);
        toast.info(t("workspace.podOpened"), {
          description: `Pod: ${getShortPodKey(podKey)}`,
        });
      }
      router.replace(window.location.pathname);
    }
  }, [_hasHydrated, searchParams, panes, router, t, handleOpenPod]);

  if (!_hasHydrated) {
    return <CenteredSpinner />;
  }

  if (panes.length === 0) {
    return (
      <>
        <WorkspaceEmptyState onCreatePod={handleCreatePod} />
        <CreatePodModal
          open={showCreateModal}
          onClose={handleCloseCreate}
          onCreated={handlePodCreated}
          initialAgentSlug={recipe?.agentSlug}
          initialPrompt={recipe?.prompt}
        />
      </>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <WorkspaceManager className="flex-1" />

      <CreatePodModal
        open={showCreateModal}
        onClose={handleCloseCreate}
        onCreated={handlePodCreated}
        initialAgentSlug={recipe?.agentSlug}
        initialPrompt={recipe?.prompt}
      />
    </div>
  );
}
