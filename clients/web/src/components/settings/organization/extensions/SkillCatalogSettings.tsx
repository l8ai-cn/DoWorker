"use client";

import { useState, useEffect, useCallback } from "react";
import type { CatalogSkill } from "@/lib/api";
import { skillCatalogApi } from "@/lib/api";
import { useCurrentOrg } from "@/stores/auth";
import { getLocalizedErrorMessage } from "@/lib/api/errors";
import { toast } from "sonner";
import type { TranslationFn } from "../GeneralSettings";
import { CatalogSkillList } from "./CatalogSkillList";
import { ImportSkillDialog } from "./ImportSkillDialog";

interface SkillCatalogSettingsProps {
  t: TranslationFn;
}

export function SkillCatalogSettings({ t }: SkillCatalogSettingsProps) {
  const currentOrg = useCurrentOrg();
  const orgSlug = currentOrg?.slug ?? "";
  const [skills, setSkills] = useState<CatalogSkill[]>([]);
  const [loading, setLoading] = useState(true);
  const [showImport, setShowImport] = useState(false);
  const [syncingSlug, setSyncingSlug] = useState<string | null>(null);

  const load = useCallback(async (signal?: AbortSignal) => {
    if (!orgSlug) return;
    try {
      const res = await skillCatalogApi.list();
      if (signal?.aborted) return;
      setSkills(res.skills);
    } catch (error) {
      if (!signal?.aborted) console.error("Failed to load skills:", error);
    } finally {
      if (!signal?.aborted) setLoading(false);
    }
  }, [orgSlug]);

  useEffect(() => {
    const controller = new AbortController();
    load(controller.signal);
    return () => controller.abort();
  }, [load]);

  const handleSync = useCallback(async (slug: string) => {
    setSyncingSlug(slug);
    try {
      await skillCatalogApi.syncUpstream(slug);
      toast.success(t("extensions.skillCatalog.synced"));
      load();
    } catch (error) {
      toast.error(getLocalizedErrorMessage(error, t, t("extensions.skillCatalog.failedToSync")));
    } finally {
      setSyncingSlug(null);
    }
  }, [t, load]);

  const handleDelete = useCallback(async (slug: string) => {
    if (!window.confirm(t("extensions.skillCatalog.confirmDelete"))) return;
    try {
      await skillCatalogApi.delete(slug);
      toast.success(t("extensions.skillCatalog.deleted"));
      load();
    } catch (error) {
      toast.error(getLocalizedErrorMessage(error, t, t("extensions.skillCatalog.failedToDelete")));
    }
  }, [t, load]);

  return (
    <div className="space-y-6">
      <CatalogSkillList
        t={t}
        loading={loading}
        skills={skills}
        syncingSlug={syncingSlug}
        onSync={handleSync}
        onDelete={handleDelete}
        onImport={() => setShowImport(true)}
      />
      <ImportSkillDialog
        t={t}
        open={showImport}
        onOpenChange={setShowImport}
        onImported={load}
      />
    </div>
  );
}
