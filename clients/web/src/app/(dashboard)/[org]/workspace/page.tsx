"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { toast } from "sonner";
import { useCurrentOrg } from "@/stores/auth";
import { useWorkspaceStore } from "@/stores/workspace";
import { usePodStore } from "@/stores/pod";
import { buildKbIngestPrompt } from "@/components/knowledgebase/kb-ingest-prompt";
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
  const currentOrg = useCurrentOrg();
  const panes = useWorkspaceStore((s) => s.panes);
  const addPane = useWorkspaceStore((s) => s.addPane);
  const openDeepLinkedPane = useWorkspaceStore((s) => s.openDeepLinkedPane);
  const _hasHydrated = useWorkspaceStore((s) => s._hasHydrated);
  const fetchPod = usePodStore((s) => s.fetchPod);
  // KB detail page "Ingest" entry (?ingest_kb=slug): the KB page seeds the
  // pod-creation store with the rw mount before navigating; here we only
  // open the create modal pre-filled with the llm-wiki maintenance prompt.
  const [ingestKb] = useState(() => searchParams.get("ingest_kb"));
  const [showCreateModal, setShowCreateModal] = useState(Boolean(ingestKb));
  const [recipe, setRecipe] = useState<WorkspaceRecipeSelection | null>(
    ingestKb ? { agentSlug: "", prompt: buildKbIngestPrompt(ingestKb) } : null,
  );
  const processedPodRef = useRef<string | null>(null);

  const handleCreatePod = useCallback((selection?: WorkspaceRecipeSelection) => {
    const orgSlug = currentOrg?.slug ?? (searchParams.get("org") ?? "");
    if (ingestKb) {
      setRecipe(selection ?? { agentSlug: "", prompt: buildKbIngestPrompt(ingestKb) });
      setShowCreateModal(true);
      return;
    }
    if (orgSlug) {
      router.push(`/${orgSlug}/workers/new?mode=template`);
      return;
    }
    setRecipe(selection ?? null);
    setShowCreateModal(true);
  }, [currentOrg?.slug, ingestKb, router, searchParams]);

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
      openDeepLinkedPane(podKey);
      void fetchPod(podKey).catch(() => {
        toast.error(t("common.error"));
      });
      toast.info(t("workspace.podOpened"), {
        description: `Pod: ${getShortPodKey(podKey)}`,
      });
    }

    if (searchParams.get("ingest_kb")) {
      router.replace(window.location.pathname);
    }
  }, [_hasHydrated, fetchPod, searchParams, router, t, openDeepLinkedPane]);

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
