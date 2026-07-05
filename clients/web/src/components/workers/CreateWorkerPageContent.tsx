"use client";

import { useMemo } from "react";
import Link from "next/link";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { CreatePodForm } from "@/components/pod/CreatePodForm";
import type { PodData } from "@/lib/api";
import { getShortPodKey } from "@/lib/pod-display-name";
import { usePodStore } from "@/stores/pod";

export function CreateWorkerPageContent() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const searchParams = useSearchParams();
  const orgSlug = params.org as string;

  const initialAgentSlug = searchParams.get("image") ?? undefined;
  const initialPrompt = searchParams.get("prompt") ?? undefined;

  const formConfig = useMemo(
    () => ({
      scenario: "workspace" as const,
      initialAgentSlug,
      initialPrompt,
      onSuccess: (pod: PodData) => {
        if (!pod?.pod_key) return;
        usePodStore.getState().upsertPod(pod);
        toast.info(t("workspace.podCreated"), {
          description: getShortPodKey(pod.pod_key),
        });
        router.push(`/${orgSlug}/workspace?pod=${encodeURIComponent(pod.pod_key)}`);
      },
      onCancel: () => router.push(`/${orgSlug}/workspace`),
    }),
    [initialAgentSlug, initialPrompt, orgSlug, router, t],
  );

  return (
    <div className="min-h-full bg-background">
      <div className="mx-auto w-full max-w-2xl px-4 py-8 md:px-6">
        <Link
          href={`/${orgSlug}/workspace`}
          className="mb-6 inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          {t("workers.create.backToWorkspace")}
        </Link>

        <header className="mb-8 space-y-2">
          <h1 className="text-2xl font-semibold tracking-tight">
            {t("workers.create.title")}
          </h1>
          <p className="text-sm text-muted-foreground leading-relaxed">
            {t("workers.create.subtitle")}
          </p>
        </header>

        <section
          aria-labelledby="worker-create-steps"
          className="mb-6 rounded-lg border border-border bg-muted/30 px-4 py-3 text-sm text-muted-foreground"
        >
          <h2 id="worker-create-steps" className="sr-only">
            {t("workers.create.stepsTitle")}
          </h2>
          <ol className="list-decimal list-inside space-y-1">
            <li>{t("workers.create.stepCluster")}</li>
            <li>{t("workers.create.stepImage")}</li>
            <li>{t("workers.create.stepLaunch")}</li>
          </ol>
        </section>

        <CreatePodForm config={formConfig} className="pb-8" />
      </div>
    </div>
  );
}
