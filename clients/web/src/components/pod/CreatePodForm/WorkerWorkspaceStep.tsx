"use client";

import { useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { POD_MODE_PTY } from "@/lib/pod-modes";
import type { WorkerCreateController } from "../hooks/workerCreateController";
import { AutomationLevelSelect } from "./AutomationLevelSelect";
import { InteractionModeToggle } from "./InteractionModeToggle";
import { PodLifecycleSection } from "./PodLifecycleSection";
import type { DestroyPolicy } from "./podLifecycleOptions";
import { BranchInput } from "./RepositorySelect";
import { WorkerRepositoryField } from "./WorkerRepositoryField";
import { WorkerWorkspaceCapabilities } from "./WorkerWorkspaceCapabilities";

interface WorkerWorkspaceStepProps {
  controller: WorkerCreateController;
  promptPlaceholder?: string;
  t: (key: string) => string;
}

export function WorkerWorkspaceStep({
  controller,
  promptPlaceholder,
  t,
}: WorkerWorkspaceStepProps) {
  const { draft } = controller.state;
  const selectedWorkerType = controller.options.status === "ready"
    ? controller.options.data.worker_types.find(
      (option) => option.slug === draft.worker_type_slug,
    )
    : undefined;
  const supportedModes = selectedWorkerType?.supported_interaction_modes?.length
    ? selectedWorkerType.supported_interaction_modes
    : [POD_MODE_PTY];
  const interactionMode = supportedModes.includes(draft.interaction_mode)
    ? draft.interaction_mode
    : supportedModes[0];

  useEffect(() => {
    if (draft.interaction_mode !== interactionMode) {
      controller.patchDraft({ interaction_mode: interactionMode });
    }
  }, [controller, draft.interaction_mode, interactionMode]);

  return (
    <div className="space-y-6">
      <section className="space-y-4">
        <h3 className="text-sm font-semibold">{t("workerCreate.workspace.repositoryTitle")}</h3>
        <WorkerRepositoryField
          value={draft.repository_id ?? null}
          onChange={(repositoryId) => {
            const changed = repositoryId !== (draft.repository_id ?? null);
            controller.patchDraft({
              repository_id: repositoryId ?? undefined,
              branch: changed ? "" : draft.branch,
              skill_ids: changed ? [] : draft.skill_ids,
            });
          }}
        />
        {draft.repository_id && (
          <BranchInput
            value={draft.branch}
            onChange={(branch) => controller.patchDraft({ branch })}
            error={!draft.branch.trim() ? t("workerCreate.workspace.branchRequired") : undefined}
            t={t}
          />
        )}
      </section>

      <section className="space-y-4">
        <h3 className="text-sm font-semibold">{t("workerCreate.workspace.capabilitiesTitle")}</h3>
        <WorkerWorkspaceCapabilities controller={controller} t={t} />
      </section>

      <section className="grid gap-5 md:grid-cols-2">
        <AutomationLevelSelect
          value={draft.automation_level}
          onChange={(automation_level) => controller.patchDraft({ automation_level })}
          supportedModes={supportedModes}
          t={t}
        />
        <InteractionModeToggle
          supportedModes={supportedModes}
          interactionMode={interactionMode}
          onModeChange={(interaction_mode) => controller.patchDraft({ interaction_mode })}
        />
      </section>

      <section className="space-y-4">
        <LabeledTextArea
          id="worker-instructions"
          label={t("workerCreate.workspace.instructions")}
          value={draft.instructions}
          onChange={(instructions) => controller.patchDraft({ instructions })}
        />
        <LabeledTextArea
          id="worker-initial-task"
          label={t("ide.createPod.prompt")}
          value={draft.initial_task}
          placeholder={promptPlaceholder ?? t("ide.createPod.promptPlaceholder")}
          onChange={(initial_task) => controller.patchDraft({ initial_task })}
        />
        <div>
          <label htmlFor="worker-alias" className="mb-2 block text-sm font-medium">
            {t("ide.createPod.alias")}
          </label>
          <Input
            id="worker-alias"
            value={draft.alias}
            maxLength={100}
            placeholder={t("ide.createPod.aliasPlaceholder")}
            onChange={(event) => controller.patchDraft({ alias: event.target.value })}
          />
        </div>
      </section>

      <PodLifecycleSection
        destroyPolicy={draft.termination_policy as DestroyPolicy}
        destroyAfterMinutes={draft.idle_timeout_minutes}
        onPolicyChange={(policy) =>
          controller.setLifecycle(policy, draft.idle_timeout_minutes || 30)
        }
        onAfterChange={(minutes) => controller.setLifecycle("idle", minutes)}
      />

      <details className="rounded-md border border-border p-3">
        <summary className="cursor-pointer text-sm font-medium">
          {t("workerCreate.workspace.previewTitle")}
        </summary>
        <pre className="mt-3 max-h-72 overflow-auto whitespace-pre-wrap text-xs text-muted-foreground">
          {JSON.stringify(draft, null, 2)}
        </pre>
      </details>
    </div>
  );
}

function LabeledTextArea(props: {
  id: string;
  label: string;
  value: string;
  placeholder?: string;
  onChange: (value: string) => void;
}) {
  return (
    <div>
      <label htmlFor={props.id} className="mb-2 block text-sm font-medium">{props.label}</label>
      <Textarea
        id={props.id}
        rows={4}
        value={props.value}
        placeholder={props.placeholder}
        onChange={(event) => props.onChange(event.target.value)}
      />
    </div>
  );
}
