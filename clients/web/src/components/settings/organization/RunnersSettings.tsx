"use client";

import { useState, useEffect } from "react";
import { useRunnerStore, useRunners, Runner } from "@/stores/runner";
import { useCurrentOrg } from "@/stores/auth";
import { RunnersPanel, TokenDialog, EditRunnerDialog } from "./runners";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { getLocalizedErrorMessage } from "@/lib/api/errors";
import type { TranslationFn } from "./GeneralSettings";
import { listExecutionClusters } from "@/lib/api/facade/executionClusterApi";
import type { ExecutionCluster } from "@/lib/api/facade/executionCluster";

interface RunnersSettingsProps {
  t: TranslationFn;
}

export function RunnersSettings({ t }: RunnersSettingsProps) {
  const i18n = useTranslations();
  const runners = useRunners();
  const loading = useRunnerStore((s) => s.loading);
  const error = useRunnerStore((s) => s.error);
  const fetchRunners = useRunnerStore((s) => s.fetchRunners);
  const updateRunner = useRunnerStore((s) => s.updateRunner);
  const deleteRunner = useRunnerStore((s) => s.deleteRunner);
  const createToken = useRunnerStore((s) => s.createToken);
  const clearError = useRunnerStore((s) => s.clearError);
  const currentOrg = useCurrentOrg();

  const [editingRunner, setEditingRunner] = useState<Runner | null>(null);
  const [generatedToken, setGeneratedToken] = useState<string | null>(null);
  const [clusters, setClusters] = useState<ExecutionCluster[]>([]);
  const [selectedClusterId, setSelectedClusterId] = useState("");
  const [loadedClusterOrgSlug, setLoadedClusterOrgSlug] = useState<
    string | null
  >(null);
  const currentOrgSlug = currentOrg?.slug ?? "";
  const clustersLoading =
    currentOrgSlug !== "" && loadedClusterOrgSlug !== currentOrgSlug;
  const visibleClusters =
    loadedClusterOrgSlug === currentOrgSlug ? clusters : [];
  const visibleClusterId =
    loadedClusterOrgSlug === currentOrgSlug ? selectedClusterId : "";

  useEffect(() => {
    fetchRunners();
  }, [fetchRunners]);
  useEffect(() => {
    if (!currentOrgSlug) return;
    let active = true;
    void listExecutionClusters(currentOrgSlug)
      .then((items) => {
        if (!active) return;
        setClusters(items);
        setLoadedClusterOrgSlug(currentOrgSlug);
      })
      .catch((err) => {
        if (!active) return;
        setClusters([]);
        setLoadedClusterOrgSlug(currentOrgSlug);
        toast.error(getLocalizedErrorMessage(err, i18n, i18n("common.error")));
      });
    return () => {
      active = false;
    };
  }, [currentOrgSlug, i18n]);

  const handleGenerateToken = async () => {
    if (!visibleClusterId) return;
    try {
      const token = await createToken({
        cluster_id: Number(visibleClusterId),
      });
      setGeneratedToken(token);
    } catch (err) {
      console.error("Failed to generate token:", err);
      toast.error(getLocalizedErrorMessage(err, i18n, i18n("common.error")));
    }
  };

  return (
    <div className="space-y-6">
      {error && (
        <div className="bg-destructive/10 border border-destructive text-destructive px-4 py-3 rounded-lg flex items-center justify-between">
          <span>{error}</span>
          <button onClick={clearError} className="text-sm underline">
            {t("settings.members.dismiss")}
          </button>
        </div>
      )}

      <RunnersPanel
        runners={runners}
        loading={loading}
        onEdit={setEditingRunner}
        onDelete={deleteRunner}
        onGenerateToken={handleGenerateToken}
        clusters={visibleClusters}
        selectedClusterId={visibleClusterId}
        onSelectCluster={setSelectedClusterId}
        clustersLoading={clustersLoading}
        t={t}
      />

      {editingRunner && (
        <EditRunnerDialog
          runner={editingRunner}
          onClose={() => setEditingRunner(null)}
          onSave={async (id, data) => {
            await updateRunner(id, data);
            setEditingRunner(null);
          }}
          t={t}
        />
      )}

      {generatedToken && (
        <TokenDialog
          token={generatedToken}
          onClose={() => setGeneratedToken(null)}
          onCopy={() => navigator.clipboard.writeText(generatedToken)}
          t={t}
        />
      )}
    </div>
  );
}
