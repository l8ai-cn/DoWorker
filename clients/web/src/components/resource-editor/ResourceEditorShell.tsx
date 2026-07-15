"use client";

import dynamic from "next/dynamic";
import { FileCode2, ListChecks, SlidersHorizontal } from "lucide-react";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import {
  isResourceBindingKind,
  type ResourceEditorKind,
} from "./resource-editor-types";
import type { ResourceApplyResult } from "./resource-apply-result";
import { ResourceConfigurationPanel } from "./ResourceConfigurationPanel";
import { ResourceEditorActions } from "./ResourceEditorActions";
import { ResourceEditorFeedback } from "./ResourceEditorFeedback";
import { ResourcePlanReview } from "./ResourcePlanReview";
import { useResourceEditorController } from "./use-resource-editor-controller";

const ResourceYamlPanel = dynamic(
  () => import("./ResourceYamlPanel").then((module) => module.ResourceYamlPanel),
  { ssr: false },
);

interface ResourceEditorShellProps {
  orgSlug: string;
  kind?: ResourceEditorKind;
  onApplied?: (result: ResourceApplyResult) => void;
  onWorkerCreated?: (podKey: string) => void;
}

export function ResourceEditorShell({
  orgSlug,
  kind = "WorkerTemplate",
  onApplied,
  onWorkerCreated,
}: ResourceEditorShellProps) {
  const t = useTranslations("resourceEditor");
  const controller = useResourceEditorController(orgSlug, kind);
  const { state } = controller;
  const isApplying = state.apply.status === "loading";
  const status = state.apply.status === "ready"
    ? "applied"
    : state.plan.status === "ready" && state.plan.response.plan
      ? "planReady"
      : "draft";

  return (
    <section
      className="space-y-5"
      data-testid="resource-editor"
      aria-busy={isApplying}
    >
      <header className="flex flex-wrap items-start justify-between gap-3 border-b border-border pb-4">
        <div>
          <h2 className="text-lg font-semibold">
            {resourceHeading(t, kind).title}
          </h2>
          <p className="mt-1 max-w-3xl text-sm text-muted-foreground">
            {resourceHeading(t, kind).subtitle}
          </p>
        </div>
        <Badge variant={status === "applied"
          ? "success"
          : status === "planReady"
            ? "info"
            : "secondary"}
        >
          {t(`status.${status}`)}
        </Badge>
      </header>

      <fieldset
        disabled={isApplying}
        className="m-0 min-w-0 space-y-5 border-0 p-0"
      >
        <Tabs
          value={state.mode}
          onValueChange={(value) => {
            void controller.setMode(value as "form" | "yaml" | "plan");
          }}
        >
        <TabsList className="h-auto max-w-full overflow-x-auto">
          <TabsTrigger value="form">
            <SlidersHorizontal className="mr-2 h-4 w-4" />
            {t("tabs.form")}
          </TabsTrigger>
          <TabsTrigger value="yaml">
            <FileCode2 className="mr-2 h-4 w-4" />
            {t("tabs.yaml")}
          </TabsTrigger>
          <TabsTrigger value="plan">
            <ListChecks className="mr-2 h-4 w-4" />
            {t("tabs.plan")}
          </TabsTrigger>
        </TabsList>
        <TabsContent value="form" className="pt-4">
          <ResourceConfigurationPanel
            orgSlug={orgSlug}
            draft={state.draft}
            onChange={controller.replaceDraft}
          />
        </TabsContent>
        <TabsContent value="yaml" className="pt-4">
          <ResourceYamlPanel
            kind={kind}
            value={state.source.text}
            error={state.source.error}
            onChange={controller.setSource}
          />
        </TabsContent>
        <TabsContent value="plan" className="pt-4">
          <ResourcePlanReview planState={state.plan} />
        </TabsContent>
        </Tabs>

        <ResourceEditorFeedback state={state} />
        <ResourceEditorActions
          state={state}
          kind={kind}
          canSubmit={controller.canSubmit}
          canApply={controller.canApply}
          onValidate={() => void controller.runValidation()}
          onPlan={() => void controller.runPlan()}
          onApply={() => {
            void controller.apply().then((result) => {
              if (result) onApplied?.(result);
              if (kind === "Worker" && result && "podKey" in result) {
                onWorkerCreated?.(result.podKey);
              }
            });
          }}
        />
      </fieldset>
    </section>
  );
}

function resourceHeading(
  t: ReturnType<typeof useTranslations<"resourceEditor">>,
  kind: ResourceEditorKind,
) {
  if (kind === "WorkerTemplate") {
    return { title: t("title"), subtitle: t("subtitle") };
  }
  if (kind === "Worker") {
    return { title: t("worker.title"), subtitle: t("worker.subtitle") };
  }
  if (kind === "Prompt") {
    return { title: t("prompt.title"), subtitle: t("prompt.subtitle") };
  }
  if (kind === "Expert") {
    return { title: t("expert.title"), subtitle: t("expert.subtitle") };
  }
  if (kind === "Workflow") {
    return { title: t("workflow.title"), subtitle: t("workflow.subtitle") };
  }
  if (kind === "GoalLoop") {
    return { title: t("goalLoop.title"), subtitle: t("goalLoop.subtitle") };
  }
  if (isResourceBindingKind(kind)) {
    return {
      title: t("binding.title", { kind }),
      subtitle: t("binding.subtitle"),
    };
  }
  return { title: kind, subtitle: "" };
}
