"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter, usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { useCurrentOrg } from "@/stores/auth";
import { useExpertStore, useExperts } from "@/stores/expert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Loader2, Plus, RefreshCw, Search, Bot } from "lucide-react";
import { useTranslations } from "next-intl";
import { formatTimeAgo } from "@/lib/utils/time";

export function ExpertsSidebarContent({ className }: { className?: string }) {
  const t = useTranslations("experts");
  const tRoot = useTranslations();
  const router = useRouter();
  const pathname = usePathname();
  const currentOrg = useCurrentOrg();
  const experts = useExperts();
  const loading = useExpertStore((s) => s.loading);
  const fetchExperts = useExpertStore((s) => s.fetchExperts);

  const [searchQuery, setSearchQuery] = useState("");
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => {
    if (currentOrg) fetchExperts();
  }, [currentOrg, fetchExperts]);

  const handleRefresh = useCallback(async () => {
    setRefreshing(true);
    try {
      await fetchExperts();
    } finally {
      setRefreshing(false);
    }
  }, [fetchExperts]);

  const filtered = searchQuery
    ? experts.filter(
        (e) =>
          e.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
          e.slug.toLowerCase().includes(searchQuery.toLowerCase()),
      )
    : experts;

  const activeSlug = pathname?.match(/\/experts\/([^/]+)/)?.[1] ?? null;

  return (
    <div className={cn("flex flex-col h-full", className)}>
      <div className="px-2 py-2">
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
          <Input
            placeholder={t("searchPlaceholder")}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-8 h-8 text-sm bg-muted/50 border-transparent focus:border-border focus:bg-background"
          />
        </div>
      </div>

      <div className="flex items-center gap-1 px-2 pb-2">
        <Button
          size="sm"
          variant="outline"
          className="flex-1 h-7 text-xs gap-1"
          onClick={() => router.push(`/${currentOrg?.slug}/experts/new`)}
        >
          <Plus className="w-3 h-3" />
          {t("createExpert")}
        </Button>
        <Button
          size="sm"
          variant="ghost"
          className="h-7 w-7 p-0"
          onClick={handleRefresh}
          disabled={refreshing}
        >
          <RefreshCw className={cn("w-3.5 h-3.5", refreshing && "animate-spin")} />
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto">
        <div className="px-3 py-1.5 text-[10px] uppercase tracking-wider text-muted-foreground font-medium">
          {t("expertCount", { count: filtered.length })}
        </div>

        {loading && experts.length === 0 ? (
          <div className="flex justify-center py-8">
            <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
          </div>
        ) : filtered.length === 0 ? (
          <p className="px-3 py-6 text-center text-sm text-muted-foreground">
            {searchQuery ? t("noMatch") : t("emptyState")}
          </p>
        ) : (
          filtered.map((expert) => (
            <button
              key={expert.slug}
              type="button"
              onClick={() => router.push(`/${currentOrg?.slug}/experts/${expert.slug}`)}
              className={cn(
                "w-full text-left px-3 py-2.5 hover:bg-surface-muted transition-colors",
                activeSlug === expert.slug && "bg-muted/40",
              )}
            >
              <div className="flex items-center gap-2">
                <Bot className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                <span className="text-sm font-medium truncate">{expert.name}</span>
              </div>
              <p className="mt-0.5 text-xs text-muted-foreground truncate pl-5.5">
                {expert.agent_slug}
                {expert.run_count > 0 && ` · ${t("runCount", { count: expert.run_count })}`}
              </p>
              {expert.last_run_at && (
                <p className="text-[10px] text-muted-foreground/70 pl-5.5 mt-0.5">
                  {t("lastRun", { time: formatTimeAgo(expert.last_run_at, tRoot) })}
                </p>
              )}
            </button>
          ))
        )}
      </div>
    </div>
  );
}
