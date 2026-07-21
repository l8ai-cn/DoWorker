"use client";

import { useState } from "react";
import { CheckCircle2, Loader2 } from "lucide-react";
import { AlertMessage } from "@/components/ui/alert-message";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type { WorkerCreateController } from "../hooks/workerCreateController";
import { workerPreflightHasBlockingIssues } from "../hooks/workerCreateController";

interface WorkerQuickCreateProps {
  controller: WorkerCreateController;
  t: (key: string) => string;
}

export function WorkerQuickCreate({ controller, t }: WorkerQuickCreateProps) {
  const [localError, setLocalError] = useState<string | null>(null);
  const { state } = controller;
  const loading = state.preflight.status === "loading" || state.create.status === "loading";
  const task = state.draft.initial_task;

  async function createFromDefaults() {
    if (!task.trim()) {
      setLocalError(t("workers.create.quick.taskRequired"));
      return;
    }
    setLocalError(null);
    const checked = await controller.runPreflight();
    if (
      !checked ||
      workerPreflightHasBlockingIssues(checked) ||
      !checked.resolved_spec_json?.trim()
    ) {
      setLocalError(t("workers.create.quick.preflightFailed"));
      return;
    }
    await controller.createWorker(checked);
  }

  return (
    <section className="rounded-lg border border-border bg-surface-raised p-4 md:p-5">
      <div className="mb-4">
        <h2 className="text-base font-semibold">{t("workers.create.quick.title")}</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          {t("workers.create.quick.subtitle")}
        </p>
      </div>

      <label htmlFor="worker-quick-task" className="mb-2 block text-sm font-medium">
        {t("workers.create.quick.taskLabel")}
      </label>
      <Textarea
        id="worker-quick-task"
        value={task}
        rows={4}
        maxLength={10000}
        placeholder={t("workers.create.quick.taskPlaceholder")}
        onChange={(event) => {
          setLocalError(null);
          controller.patchDraft({ initial_task: event.target.value });
        }}
      />

      <DefaultSummary controller={controller} t={t} />

      <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-xs text-muted-foreground">
          {t("workers.create.quick.defaultsHint")}
        </p>
        <Button
          type="button"
          className="h-11 sm:h-9"
          disabled={loading || !controller.validity.workspace}
          onClick={() => void createFromDefaults()}
        >
          {loading ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <CheckCircle2 className="mr-2 h-4 w-4" />
          )}
          {loading ? t("workers.create.quick.creating") : t("workers.create.quick.create")}
        </Button>
      </div>

      <div className="mt-4 space-y-3">
        {!controller.validity.workspace && (
          <AlertMessage type="warning" message={t("workers.create.quick.defaultsNotReady")} />
        )}
        {localError && <AlertMessage type="error" message={localError} />}
        {state.create.status === "error" && (
          <AlertMessage type="error" message={state.create.error} />
        )}
      </div>
    </section>
  );
}

function DefaultSummary({
  controller,
  t,
}: {
  controller: WorkerCreateController;
  t: (key: string) => string;
}) {
  const { state, options, modelResources } = controller;
  if (options.status !== "ready") {
    return (
      <p className="mt-3 text-xs text-muted-foreground">
        {t("workers.create.quick.loadingDefaults")}
      </p>
    );
  }
  const draft = state.draft;
  const selected = [
    optionName(options.data.worker_types, draft.worker_type_slug, (item) => item.slug),
    modelResources.status === "ready"
      ? modelResourceName(modelResources.data, draft.model_resource_id)
      : null,
    optionName(options.data.runtime_images, draft.runtime_image_id, (item) => item.id),
    optionName(options.data.compute_targets, draft.compute_target_id, (item) => item.id),
    optionName(options.data.resource_profiles, draft.resource_profile_id, (item) => item.id),
  ].filter(Boolean);

  return (
    <div className="mt-3 flex flex-wrap gap-2">
      {selected.map((label) => (
        <span
          key={label}
          className="rounded-md border border-border bg-surface-muted px-2 py-1 text-xs text-muted-foreground"
        >
          {label}
        </span>
      ))}
    </div>
  );
}

function modelResourceName(
  resources: EffectiveResource[],
  id: number,
): string | null {
  const found = resources.find((item) => item.resource?.id === id);
  return found?.resource?.displayName ?? null;
}

function optionName<T, V>(
  options: T[],
  value: V,
  pick: (item: T) => V,
  name: (item: T) => string = (item) => String((item as { name?: string }).name ?? value),
): string | null {
  const found = options.find((item) => pick(item) === value);
  return found ? name(found) : null;
}
