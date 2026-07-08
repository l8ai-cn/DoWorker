"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { ArrowLeft, FileInput } from "lucide-react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { CreatePodForm } from "@/components/pod/CreatePodForm";
import { NlWorkerCreate } from "@/components/workers/NlWorkerCreate";
import { ImportCodexDialog } from "@/components/workers/ImportCodexDialog";
import { Button } from "@/components/ui/button";
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
  const [wizardPrompt, setWizardPrompt] = useState(
    () => searchParams.get("prompt") ?? undefined,
  );
  const initialPrompt = wizardPrompt;
  const initialExpertSlug = searchParams.get("expert") ?? undefined;
  const [importOpen, setImportOpen] = useState(false);

  const formConfig = useMemo(
    () => ({
      scenario: "workspace" as const,
      initialAgentSlug,
      initialPrompt,
      initialExpertSlug,
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
    [initialAgentSlug, initialPrompt, initialExpertSlug, orgSlug, router, t],
  );

  return (
    <div className="min-h-full bg-background">
      <div className="mx-auto w-full max-w-5xl px-4 py-8 md:px-6">
        <Link
          href={`/${orgSlug}/workspace`}
          className="mb-6 inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          {t("workers.create.backToWorkspace")}
        </Link>

        <header className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-2">
            <h1 className="text-2xl font-semibold tracking-tight">
              {t("workers.create.title")}
            </h1>
            <p className="text-sm leading-relaxed text-muted-foreground">
              {t("workers.create.subtitle")}
            </p>
          </div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="shrink-0"
            data-testid="open-import-codex"
            onClick={() => setImportOpen(true)}
          >
            <FileInput className="mr-2 h-4 w-4" />
            {t("workers.create.import.button")}
          </Button>
        </header>

        <ImportCodexDialog
          open={importOpen}
          onOpenChange={setImportOpen}
          onImported={(podKey) => {
            router.push(`/${orgSlug}/workspace?pod=${encodeURIComponent(podKey)}`);
          }}
        />

        <NlWorkerCreate orgSlug={orgSlug} onNeedsWizard={setWizardPrompt} />

        <CreatePodForm key={wizardPrompt ?? ""} config={formConfig} className="pb-8" />
      </div>
    </div>
  );
}
