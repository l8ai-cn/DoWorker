"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { FolderGit2, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { RepositorySelect } from "@/components/common/RepositorySelect";
import { ImportRepositoryModal } from "@/components/ide/modals/ImportRepositoryModal";
import { useCtaModal } from "@/hooks/useCtaModal";
import { useCurrentOrg } from "@/stores/auth";
import { useRepositories, useRepositoryStore } from "@/stores/repository";

interface WorkerRepositoryFieldProps {
  value: number | null;
  onChange: (value: number | null) => void;
}

export function WorkerRepositoryField({ value, onChange }: WorkerRepositoryFieldProps) {
  const t = useTranslations();
  const currentOrg = useCurrentOrg();
  const orgSlug = currentOrg?.slug;
  const allRepos = useRepositories();
  const loading = useRepositoryStore((s) => s.isLoading);
  const fetchRepositories = useRepositoryStore((s) => s.fetchRepositories);
  const importModal = useCtaModal(fetchRepositories);

  const activeRepos = allRepos.filter((r) => r.is_active);
  const hasRepos = activeRepos.length > 0;
  const hintKey = value
    ? "ide.createPod.repositoryLinkedHint"
    : "ide.createPod.repositoryOptionalHint";

  if (!loading && !hasRepos) {
    return (
      <div className="space-y-3 rounded-md border border-dashed border-border p-4">
        <div className="flex items-start gap-3">
          <FolderGit2 className="mt-0.5 h-5 w-5 shrink-0 text-muted-foreground" />
          <div className="space-y-1">
            <p className="text-sm font-medium">{t("ide.createPod.emptyRepositoriesTitle")}</p>
            <p className="text-sm text-muted-foreground">
              {t("ide.createPod.emptyRepositoriesDescription")}
            </p>
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button type="button" size="sm" onClick={importModal.open}>
            <Plus className="mr-1 h-4 w-4" />
            {t("ide.createPod.importRepository")}
          </Button>
          {orgSlug && (
            <Link
              href={`/${orgSlug}/infra?tab=repositories`}
              className="inline-flex h-8 items-center justify-center rounded-md bg-surface-raised px-3 text-xs font-medium text-foreground shadow-xs ring-1 ring-border/30 hover:bg-accent hover:text-accent-foreground hover:shadow-[var(--shadow-soft)] hover:ring-border/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40 focus-visible:ring-offset-2 focus-visible:ring-offset-background"
            >
              {t("ide.createPod.manageRepositories")}
            </Link>
          )}
        </div>
        <p className="text-xs text-muted-foreground">{t("ide.createPod.repositoryOptionalHint")}</p>
        <ImportRepositoryModal
          open={importModal.isOpen}
          onClose={importModal.close}
          onImported={importModal.commit}
          existingRepositories={allRepos}
        />
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <label htmlFor="worker-repository-select" className="block text-sm font-medium">
        {t("ide.createPod.selectRepository")}
      </label>
      <RepositorySelect
        id="worker-repository-select"
        value={value}
        onChange={onChange}
        allowNone
        noneLabel={t("ide.createPod.noRepositoryOption")}
        placeholder={t("ide.createPod.selectRepositoryPlaceholder")}
        loadingLabel={t("ide.createPod.loadingRepositories")}
        retryLabel={t("tickets.detail.retry")}
      />
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-xs text-muted-foreground">{t(hintKey)}</p>
        <Button type="button" variant="link" size="sm" className="h-auto px-0" onClick={importModal.open}>
          {t("ide.createPod.importRepository")}
        </Button>
      </div>
      <ImportRepositoryModal
        open={importModal.isOpen}
        onClose={importModal.close}
        onImported={importModal.commit}
        existingRepositories={allRepos}
      />
    </div>
  );
}
