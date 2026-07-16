"use client";

import { useState, useEffect, useCallback, useRef } from "react";
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
  const [loadError, setLoadError] = useState(false);
  const [showImport, setShowImport] = useState(false);
  const [syncingSlug, setSyncingSlug] = useState<string | null>(null);
  const [savingSlugs, setSavingSlugs] = useState<Set<string>>(() => new Set());
  const [saveErrorSlugs, setSaveErrorSlugs] = useState<Set<string>>(() => new Set());
  const savingSlugsRef = useRef(new Set<string>());
  const loadGenerationRef = useRef(0);

  const load = useCallback(async () => {
    const generation = ++loadGenerationRef.current;
    if (!orgSlug) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setLoadError(false);
    try {
      const result = await skillCatalogApi.listAll();
      if (generation === loadGenerationRef.current) {
        setSkills(result.skills);
      }
    } catch (error) {
      if (generation === loadGenerationRef.current) {
        setLoadError(true);
        toast.error(getLocalizedErrorMessage(error, t, t("extensions.failedToLoadSkills")));
      }
    } finally {
      if (generation === loadGenerationRef.current) {
        setLoading(false);
      }
    }
  }, [orgSlug, t]);

  useEffect(() => {
    load();
    return () => {
      loadGenerationRef.current += 1;
    };
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

  const handleUpdateTags = useCallback(async (slug: string, tags: string[]) => {
    if (savingSlugsRef.current.has(slug)) return;
    savingSlugsRef.current.add(slug);
    setSavingSlugs(new Set(savingSlugsRef.current));
    setSaveErrorSlugs((current) => {
      const next = new Set(current);
      next.delete(slug);
      return next;
    });
    try {
      const updated = await skillCatalogApi.update(slug, { tags });
      setSkills((current) => current.map((skill) => skill.slug === slug ? updated : skill));
      toast.success(t("extensions.skillCatalog.tagsSaved"));
    } catch (error) {
      setSaveErrorSlugs((current) => new Set(current).add(slug));
      toast.error(getLocalizedErrorMessage(
        error,
        t,
        t("extensions.skillCatalog.failedToSaveTags"),
      ));
      throw error;
    } finally {
      savingSlugsRef.current.delete(slug);
      setSavingSlugs(new Set(savingSlugsRef.current));
    }
  }, [t]);

  return (
    <div className="space-y-6">
      <CatalogSkillList
        t={t}
        loading={loading}
        loadError={loadError}
        skills={skills}
        syncingSlug={syncingSlug}
        savingSlugs={savingSlugs}
        saveErrorSlugs={saveErrorSlugs}
        onSync={handleSync}
        onDelete={handleDelete}
        onImport={() => setShowImport(true)}
        onRetry={load}
        onEditTags={(slug) => setSaveErrorSlugs((current) => {
          const next = new Set(current);
          next.delete(slug);
          return next;
        })}
        onUpdateTags={handleUpdateTags}
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
