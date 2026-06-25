"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Loader2, Search } from "lucide-react";
import type { SkillMarketItem } from "@/lib/api";
import { listMarketSkills } from "@/lib/api/facade/marketExtension";
import { useCurrentOrg } from "@/stores/auth";
import type { TranslationFn } from "../GeneralSettings";
import { SkillMarketCard } from "./SkillMarketCard";
import { SkillMarketDetailDrawer } from "./SkillMarketDetailDrawer";
import { SkillMarketInstallDialog } from "./SkillMarketInstallDialog";

interface SkillMarketPanelProps {
  t: TranslationFn;
}

export function SkillMarketPanel({ t }: SkillMarketPanelProps) {
  const currentOrg = useCurrentOrg();
  const orgSlug = currentOrg?.slug ?? "";
  const [skills, setSkills] = useState<SkillMarketItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState<string | null>(null);
  const [detailSkill, setDetailSkill] = useState<SkillMarketItem | null>(null);
  const [installItem, setInstallItem] = useState<SkillMarketItem | null>(null);

  const loadSkills = useCallback(
    async (query?: string, mounted?: { current: boolean }) => {
      if (!orgSlug) return;
      setLoading(true);
      try {
        const res = await listMarketSkills(orgSlug, { query });
        if (mounted && !mounted.current) return;
        setSkills(res.items);
      } catch (error) {
        if (mounted && !mounted.current) return;
        console.error("Failed to load skill market:", error);
      } finally {
        if (!mounted || mounted.current) {
          setLoading(false);
        }
      }
    },
    [orgSlug],
  );

  useEffect(() => {
    const mounted = { current: true };
    const timer = setTimeout(() => {
      loadSkills(search || undefined, mounted);
    }, 300);
    return () => {
      mounted.current = false;
      clearTimeout(timer);
    };
  }, [search, loadSkills]);

  const categories = useMemo(() => {
    const values = new Set<string>();
    for (const skill of skills) {
      if (skill.category) values.add(skill.category);
    }
    return Array.from(values).sort();
  }, [skills]);

  const visibleSkills = useMemo(
    () => (category ? skills.filter((s) => s.category === category) : skills),
    [skills, category],
  );

  const openInstall = (skill: SkillMarketItem) => {
    setDetailSkill(null);
    setInstallItem(skill);
  };

  return (
    <>
      <div className="surface-card p-6">
        <div className="mb-4">
          <h2 className="text-lg font-semibold">{t("extensions.skillMarket.title")}</h2>
          <p className="text-sm text-muted-foreground">
            {t("extensions.skillMarket.description")}
          </p>
        </div>

        <div className="relative mb-4">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            className="pl-9"
            placeholder={t("extensions.searchSkills")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>

        {categories.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-4">
            <Button
              size="sm"
              variant={category == null ? "default" : "outline"}
              onClick={() => setCategory(null)}
            >
              {t("extensions.skillMarket.allCategories")}
            </Button>
            {categories.map((cat) => (
              <Button
                key={cat}
                size="sm"
                variant={category === cat ? "default" : "outline"}
                onClick={() => setCategory(cat)}
              >
                {cat}
              </Button>
            ))}
          </div>
        )}

        {loading ? (
          <div className="flex items-center justify-center py-8 text-muted-foreground gap-2">
            <Loader2 className="h-4 w-4 animate-spin" />
            {t("extensions.loading")}
          </div>
        ) : visibleSkills.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            {t("extensions.skillMarket.noSkills")}
          </div>
        ) : (
          <>
            <p className="text-xs text-muted-foreground mb-3">
              {visibleSkills.length} {t("extensions.skillMarket.skillsFound")}
            </p>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {visibleSkills.map((skill) => (
                <SkillMarketCard
                  key={skill.id}
                  skill={skill}
                  t={t}
                  onView={() => setDetailSkill(skill)}
                  onInstall={() => openInstall(skill)}
                />
              ))}
            </div>
          </>
        )}
      </div>

      <SkillMarketDetailDrawer
        skill={detailSkill}
        open={detailSkill != null}
        onOpenChange={(open) => {
          if (!open) setDetailSkill(null);
        }}
        onInstall={openInstall}
        t={t}
      />

      <SkillMarketInstallDialog
        item={installItem}
        open={installItem != null}
        onOpenChange={(open) => {
          if (!open) setInstallItem(null);
        }}
        t={t}
      />
    </>
  );
}
