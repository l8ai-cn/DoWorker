"use client";

import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { CheckCircle2, Loader2, RefreshCw, RotateCw } from "lucide-react";

import { AlertMessage } from "@/components/ui/alert-message";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  getExpertMarketUpgrade,
  upgradeExpertFromMarket,
} from "@/lib/api/expertMarketApi";

interface ExpertMarketUpgradePanelProps {
  expertSlug: string;
  initialAvailability?: boolean;
  onUpgraded: () => void | Promise<void>;
}

export function ExpertMarketUpgradePanel({
  expertSlug,
  initialAvailability,
  onUpgraded,
}: ExpertMarketUpgradePanelProps) {
  const t = useTranslations("experts.marketUpgrade");
  const [available, setAvailable] = useState(initialAvailability ?? false);
  const [loading, setLoading] = useState(initialAvailability === undefined);
  const [upgrading, setUpgrading] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await getExpertMarketUpgrade(expertSlug);
      setAvailable(result.upgrade_available);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
    } finally {
      setLoading(false);
    }
  }, [expertSlug]);

  useEffect(() => {
    if (initialAvailability === undefined) void load();
  }, [initialAvailability, load]);

  async function upgrade() {
    setUpgrading(true);
    setError("");
    try {
      await upgradeExpertFromMarket(expertSlug);
      await onUpgraded();
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
    } finally {
      setUpgrading(false);
    }
  }

  return (
    <section className="border-b border-border px-4 py-5 sm:px-8" aria-labelledby="market-upgrade-title">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h2 id="market-upgrade-title" className="flex items-center gap-2 text-sm font-semibold">
            <RotateCw className="h-4 w-4 text-muted-foreground" />
            {t("title")}
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">{t("description")}</p>
        </div>
        {loading ? (
          <Badge variant="secondary">
            <Loader2 className="mr-1 h-3.5 w-3.5 animate-spin" />{t("checking")}
          </Badge>
        ) : available ? (
          <Button size="sm" onClick={upgrade} disabled={upgrading}>
            {upgrading ? <Loader2 className="h-4 w-4 animate-spin" /> : <RotateCw className="h-4 w-4" />}
            {upgrading ? t("upgrading") : t("upgrade")}
          </Button>
        ) : !error ? (
          <Badge variant="success">
            <CheckCircle2 className="mr-1 h-3.5 w-3.5" />{t("current")}
          </Badge>
        ) : null}
      </div>
      {error ? (
        <div className="mt-4 space-y-3">
          <AlertMessage type="error" message={error} />
          <Button size="sm" variant="outline" onClick={load} disabled={loading || upgrading}>
            <RefreshCw className="h-4 w-4" />{t("retry")}
          </Button>
        </div>
      ) : null}
    </section>
  );
}
