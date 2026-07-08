"use client";

import { useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { Bot, Plus } from "lucide-react";
import { useExpertStore, useExperts } from "@/stores/expert";
import { useCurrentOrg } from "@/stores/auth";
import { CenteredSpinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/empty-state";
import { Button } from "@/components/ui/button";

export default function ExpertsIndexPage() {
  const t = useTranslations("experts");
  const router = useRouter();
  const params = useParams();
  const orgSlug = params.org as string;
  const currentOrg = useCurrentOrg();
  const experts = useExperts();
  const loading = useExpertStore((s) => s.loading);
  const fetchExperts = useExpertStore((s) => s.fetchExperts);

  useEffect(() => {
    if (currentOrg) fetchExperts();
  }, [currentOrg, fetchExperts]);

  useEffect(() => {
    if (loading || experts.length === 0) return;
    router.replace(`/${orgSlug}/experts/${experts[0].slug}`);
  }, [experts, loading, orgSlug, router]);

  if (loading && experts.length === 0) return <CenteredSpinner className="h-full" />;

  if (experts.length === 0) {
    return (
      <EmptyState
        size="full"
        icon={<Bot className="h-12 w-12" />}
        title={t("emptyTitle")}
        description={t("emptyDescription")}
        actions={
          <Button onClick={() => router.push(`/${orgSlug}/experts/new`)} className="gap-1.5">
            <Plus className="h-4 w-4" />
            {t("createExpert")}
          </Button>
        }
      />
    );
  }

  return <CenteredSpinner className="h-full" />;
}
