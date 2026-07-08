"use client";

import { cn } from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import type { RepositoryData } from "@/lib/api";

interface RepositorySelectProps {
  repositories: RepositoryData[];
  selectedRepositoryId: number | null;
  onSelect: (repositoryId: number | null) => void;
  t: (key: string) => string;
}

export function RepositorySelect({
  repositories,
  selectedRepositoryId,
  onSelect,
  t,
}: RepositorySelectProps) {
  const selectedRepo = repositories.find((repo) => repo.id === selectedRepositoryId);

  return (
    <div>
      <label
        htmlFor="repository-select"
        className="block text-sm font-medium mb-2"
      >
        {t("ide.createPod.selectRepository")}
      </label>
      <Select
        value={selectedRepositoryId ? String(selectedRepositoryId) : ""}
        onValueChange={(value) => onSelect(value ? Number(value) : null)}
      >
        <SelectTrigger id="repository-select">
          <span className={cn(!selectedRepositoryId && "text-muted-foreground")}>
            {selectedRepo?.slug ?? t("ide.createPod.selectRepositoryPlaceholder")}
          </span>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="">{t("ide.createPod.selectRepositoryPlaceholder")}</SelectItem>
          {repositories.map((repo) => (
            <SelectItem key={repo.id} value={String(repo.id)}>
              {repo.slug}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}

interface BranchInputProps {
  value: string;
  onChange: (value: string) => void;
  error?: string;
  t: (key: string) => string;
}

export function BranchInput({
  value,
  onChange,
  error,
  t,
}: BranchInputProps) {
  return (
    <div>
      <label
        htmlFor="branch-input"
        className="block text-sm font-medium mb-2"
      >
        {t("ide.createPod.branch")}
      </label>
      <input
        id="branch-input"
        type="text"
        className={`w-full px-3 py-2 border rounded-md bg-background ${
          error ? "border-destructive" : "border-border"
        }`}
        placeholder={t("ide.createPod.branchPlaceholder")}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        aria-invalid={!!error}
        aria-describedby={error ? "branch-error" : undefined}
      />
      {error && (
        <p id="branch-error" className="text-xs text-destructive mt-1">
          {error}
        </p>
      )}
    </div>
  );
}
