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

interface TicketRepositoryFieldProps {
  value: number | null;
  onChange: (value: number | null) => void;
}

export function TicketRepositoryField({ value, onChange }: TicketRepositoryFieldProps) {
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
    ? "tickets.createDialog.repositoryLinkedHint"
    : "tickets.createDialog.repositoryOptionalHint";

  if (!loading && !hasRepos) {
    return (
      <div className="space-y-3 rounded-md border border-dashed border-border p-4">
        <div className="flex items-start gap-3">
          <FolderGit2 className="mt-0.5 h-5 w-5 shrink-0 text-muted-foreground" />
          <div className="space-y-1">
            <p className="text-sm font-medium">{t("tickets.createDialog.emptyRepositoriesTitle")}</p>
            <p className="text-sm text-muted-foreground">
              {t("tickets.createDialog.emptyRepositoriesDescription")}
            </p>
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button type="button" size="sm" onClick={importModal.open}>
            <Plus className="mr-1 h-4 w-4" />
            {t("tickets.createDialog.importRepository")}
          </Button>
          {orgSlug && (
            <Button type="button" size="sm" variant="outline" asChild>
              <Link href={`/${orgSlug}/infra?tab=repositories`}>
                {t("tickets.createDialog.manageRepositories")}
              </Link>
            </Button>
          )}
        </div>
        <p className="text-xs text-muted-foreground">{t("tickets.createDialog.repositoryOptionalHint")}</p>
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
      <RepositorySelect
        value={value}
        onChange={onChange}
        allowNone
        noneLabel={t("tickets.createDialog.noRepositoryOption")}
        placeholder={t("tickets.createDialog.selectRepository")}
        loadingLabel={t("tickets.createDialog.loadingRepositories")}
        retryLabel={t("tickets.detail.retry")}
      />
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-xs text-muted-foreground">{t(hintKey)}</p>
        <Button type="button" variant="link" size="sm" className="h-auto px-0" onClick={importModal.open}>
          {t("tickets.createDialog.importRepository")}
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
