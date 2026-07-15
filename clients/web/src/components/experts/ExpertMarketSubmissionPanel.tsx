"use client";

import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Loader2, RefreshCw, Send, Store } from "lucide-react";
import { toast } from "sonner";

import { AlertMessage } from "@/components/ui/alert-message";
import { Badge, type BadgeProps } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import {
  listExpertMarketSubmissions,
  withdrawExpertMarketRelease,
  type ExpertMarketRelease,
} from "@/lib/api/expertMarketApi";
import { ExpertMarketSubmissionDialog } from "./ExpertMarketSubmissionDialog";

const statusVariants: Record<ExpertMarketRelease["status"], BadgeProps["variant"]> = {
  draft: "secondary",
  pending_review: "warning",
  published: "success",
  rejected: "destructive",
  withdrawn: "secondary",
};

interface ExpertMarketSubmissionPanelProps {
  expertID: number;
  expertSlug: string;
  submissionReady: boolean;
}

export function ExpertMarketSubmissionPanel({
  expertID,
  expertSlug,
  submissionReady,
}: ExpertMarketSubmissionPanelProps) {
  const t = useTranslations("experts.marketSubmission");
  const [release, setRelease] = useState<ExpertMarketRelease | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [withdrawOpen, setWithdrawOpen] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await listExpertMarketSubmissions();
      const latest = result.releases
        .filter((item) => item.source_expert_id === expertID)
        .sort((left, right) => right.version - left.version)[0] ?? null;
      setRelease(latest);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
    } finally {
      setLoading(false);
    }
  }, [expertID]);

  useEffect(() => {
    void load();
  }, [load]);

  async function withdraw() {
    if (!release) return;
    try {
      await withdrawExpertMarketRelease(release.id);
      await load();
    } catch (cause) {
      toast.error(t("withdrawFailed"));
      throw cause;
    }
  }

  const canSubmit = !release || ["rejected", "withdrawn", "published"].includes(release.status);
  const marketSlug = release ? release.application_slug : expertSlug;
  return (
    <section className="border-b border-border px-4 py-5 sm:px-8" aria-labelledby="market-release-title">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2 id="market-release-title" className="flex items-center gap-2 text-sm font-semibold">
            <Store className="h-4 w-4 text-muted-foreground" />
            {t("title")}
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">{t("description")}</p>
        </div>
        {canSubmit && !loading ? (
          <Button size="sm" variant={release ? "outline" : "default"}
            onClick={() => setDialogOpen(true)} disabled={!submissionReady}>
            <Send className="h-4 w-4" />
            {release ? t("submitNew") : t("submitFirst")}
          </Button>
        ) : null}
      </div>

      {loading ? (
        <div className="mt-4 flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />{t("loading")}
        </div>
      ) : null}
      {error ? (
        <div className="mt-4 space-y-3">
          <AlertMessage type="error" message={error} />
          <Button size="sm" variant="outline" onClick={load}>
            <RefreshCw className="h-4 w-4" />{t("retry")}
          </Button>
        </div>
      ) : null}
      {!loading && !error && !release ? (
        <p className="mt-4 text-sm text-muted-foreground">
          {submissionReady ? t("empty") : t("snapshotRequired")}
        </p>
      ) : null}
      {!loading && !error && release ? (
        <div className="mt-4 flex flex-wrap items-start justify-between gap-4 rounded-md border border-border bg-muted/20 p-4">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Badge variant={statusVariants[release.status]}>{t(statusKey(release.status))}</Badge>
              <span className="text-xs text-muted-foreground">{t("version", { version: release.version })}</span>
            </div>
            <p className="text-sm font-medium">{release.summary}</p>
            {release.status === "rejected" && release.rejection_reason ? (
              <AlertMessage type="error" message={release.rejection_reason} />
            ) : null}
          </div>
          {release.status === "published" ? (
            <Button size="sm" variant="outline" onClick={() => setWithdrawOpen(true)}>
              {t("withdraw")}
            </Button>
          ) : null}
        </div>
      ) : null}

      <ExpertMarketSubmissionDialog expertSlug={expertSlug} marketSlug={marketSlug}
        marketSlugLocked={Boolean(release)} open={dialogOpen}
        onOpenChange={setDialogOpen} onSubmitted={load} />
      <ConfirmDialog open={withdrawOpen} onOpenChange={setWithdrawOpen}
        title={t("withdrawTitle")} description={t("withdrawDescription")}
        confirmText={t("withdrawConfirm")} variant="warning" onConfirm={withdraw} />
    </section>
  );
}

function statusKey(status: ExpertMarketRelease["status"]) {
  return status === "pending_review" ? "pendingReview"
    : status === "published" ? "published"
      : status === "rejected" ? "rejected"
        : status === "withdrawn" ? "withdrawn" : "draft";
}
