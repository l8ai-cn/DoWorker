"use client";

import { useEffect, useMemo } from "react";
import { cn } from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { RepositoryData } from "@/lib/api";
import { useRepositories, useRepositoryStore } from "@/stores/repository";

export interface RepositorySelectProps {
  value: number | null;
  onChange: (value: number | null, repository?: RepositoryData) => void;
  disabled?: boolean;
  placeholder?: string;
  loadingLabel?: string;
  retryLabel?: string;
  noneLabel?: string;
  allowNone?: boolean;
  className?: string;
  activeOnly?: boolean;
  id?: string;
}

export function RepositorySelect({
  value,
  onChange,
  disabled = false,
  placeholder = "Select a repository...",
  loadingLabel = "Loading repositories...",
  retryLabel = "Retry",
  noneLabel,
  allowNone = false,
  className = "",
  activeOnly = true,
  id,
}: RepositorySelectProps) {
  const allRepos = useRepositories();
  const loading = useRepositoryStore((s) => s.isLoading);
  const error = useRepositoryStore((s) => s.error);
  const fetchRepositories = useRepositoryStore((s) => s.fetchRepositories);

  useEffect(() => { fetchRepositories(); }, [fetchRepositories]);

  const repositories = useMemo(
    () => (activeOnly ? allRepos.filter((r) => r.is_active) : allRepos),
    [allRepos, activeOnly],
  );

  const selectedRepo = repositories.find((repo) => repo.id === value);
  const emptyLabel = loading
    ? loadingLabel
    : allowNone && noneLabel
      ? noneLabel
      : placeholder;

  if (error) {
    return (
      <div className={`text-sm text-destructive ${className}`}>
        {error}
        <button
          type="button"
          onClick={() => fetchRepositories()}
          className="ml-2 underline hover:no-underline"
        >
          {retryLabel}
        </button>
      </div>
    );
  }

  return (
    <Select
      value={value ? String(value) : ""}
      onValueChange={(next) => {
        if (!next) {
          onChange(null);
          return;
        }
        const selectedId = Number(next);
        onChange(selectedId, repositories.find((r) => r.id === selectedId));
      }}
      disabled={disabled || loading}
    >
      <SelectTrigger id={id} className={className}>
        <span className={cn(!value && "text-muted-foreground")}>
          {selectedRepo?.slug ?? emptyLabel}
        </span>
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="">{emptyLabel}</SelectItem>
        {repositories.map((repo) => (
          <SelectItem key={repo.id} value={String(repo.id)}>
            {repo.slug}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

export default RepositorySelect;
