"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Plus, Play, Trash2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/empty-state";
import { useCoordinatorStore } from "@/stores/coordinator";
import { CreateProjectDialog } from "./CreateProjectDialog";
import { ExecutionBoard } from "./ExecutionBoard";

export function AutomationView() {
  const t = useTranslations("automation");
  const { projects, executions, loading, error, loadProjects, loadExecutions, runNow, deleteProject, updateProject } =
    useCoordinatorStore();
  const [createOpen, setCreateOpen] = useState(false);
  const [selected, setSelected] = useState<number | null>(null);

  useEffect(() => {
    loadProjects();
  }, [loadProjects]);

  // Derive the active project so the first one shows without a setState-in-effect
  // auto-select (which triggers cascading renders).
  const current = projects.find((p) => p.id === selected) ?? projects[0] ?? null;
  const currentId = current?.id ?? null;

  useEffect(() => {
    if (currentId != null) loadExecutions(currentId);
  }, [currentId, loadExecutions]);

  return (
    <div className="flex h-full flex-col gap-4 overflow-auto p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">{t("title")}</h1>
          <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-1.5 h-4 w-4" />
          {t("newProject")}
        </Button>
      </div>

      {error && <div className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</div>}

      {loading && projects.length === 0 ? (
        <div className="flex flex-1 items-center justify-center">
          <Spinner />
        </div>
      ) : projects.length === 0 ? (
        <EmptyState title={t("empty.title")} description={t("empty.description")} />
      ) : (
        <div className="flex flex-1 flex-col gap-4 lg:flex-row">
          <div className="flex w-full flex-col gap-2 lg:w-72">
            {projects.map((p) => (
              <Card
                key={p.id}
                variant={p.id === selected ? "inset" : "default"}
                className="cursor-pointer p-3"
                onClick={() => setSelected(p.id)}
              >
                <div className="flex items-center justify-between gap-2">
                  <span className="truncate text-sm font-medium">{p.name}</span>
                  <Badge variant={p.enabled ? "default" : "outline"}>
                    {p.enabled ? t("status.enabled") : t("status.paused")}
                  </Badge>
                </div>
                <p className="mt-1 truncate text-xs text-muted-foreground">
                  {p.platform_type} · {p.source_type} · {p.scan_interval_seconds}s
                </p>
              </Card>
            ))}
          </div>

          <div className="flex-1">
            {current && (
              <>
                <div className="mb-3 flex items-center gap-2">
                  <Button size="sm" variant="outline" onClick={() => runNow(current.id)}>
                    <Play className="mr-1.5 h-3.5 w-3.5" />
                    {t("actions.runNow")}
                  </Button>
                  <Button size="sm" variant="outline" onClick={() => loadExecutions(current.id)}>
                    <RefreshCw className="mr-1.5 h-3.5 w-3.5" />
                    {t("actions.refresh")}
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => updateProject(current.id, { enabled: !current.enabled })}
                  >
                    {current.enabled ? t("actions.pause") : t("actions.resume")}
                  </Button>
                  <Button size="sm" variant="destructive" onClick={() => deleteProject(current.id)}>
                    <Trash2 className="mr-1.5 h-3.5 w-3.5" />
                    {t("actions.delete")}
                  </Button>
                </div>
                <ExecutionBoard executions={executions[current.id] ?? []} />
              </>
            )}
          </div>
        </div>
      )}

      <CreateProjectDialog open={createOpen} onOpenChange={setCreateOpen} />
    </div>
  );
}
