"use client";

import { CheckCircle2, ClipboardList, Upload } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import type { ResourceEditorKind } from "./resource-editor-types";
import type { ResourceDraftState } from "./resource-draft-reducer";

interface ResourceEditorActionsProps {
  state: ResourceDraftState;
  kind: ResourceEditorKind;
  canSubmit: boolean;
  canPlan: boolean;
  canApply: boolean;
  onValidate: () => void;
  onPlan: () => void;
  onApply: () => void;
}

export function ResourceEditorActions({
  state,
  kind,
  canSubmit,
  canPlan,
  canApply,
  onValidate,
  onPlan,
  onApply,
}: ResourceEditorActionsProps) {
  const t = useTranslations("resourceEditor");
  return (
    <div className="sticky bottom-0 z-10 -mx-1 grid grid-cols-2 gap-2 border-t border-border bg-background/95 px-1 py-4 backdrop-blur sm:flex sm:flex-wrap sm:justify-end">
      <Button
        type="button"
        variant="outline"
        className="h-11 sm:h-9"
        disabled={!canSubmit}
        loading={state.validation.status === "loading"}
        onClick={onValidate}
      >
        <CheckCircle2 className="mr-2 h-4 w-4" />
        {state.validation.status === "loading"
          ? t("actions.validating")
          : t("actions.validate")}
      </Button>
      <Button
        type="button"
        variant="outline"
        className="h-11 sm:h-9"
        disabled={!canPlan}
        loading={state.plan.status === "loading"}
        onClick={onPlan}
      >
        <ClipboardList className="mr-2 h-4 w-4" />
        {state.plan.status === "loading"
          ? t("actions.planning")
          : t("actions.plan")}
      </Button>
      <Button
        type="button"
        className="col-span-2 h-11 sm:h-9"
        disabled={!canApply}
        loading={state.apply.status === "loading"}
        onClick={onApply}
      >
        <Upload className="mr-2 h-4 w-4" />
        {state.apply.status === "loading"
          ? t("actions.applying")
          : kind === "Worker"
            ? t("actions.createWorker")
            : kind === "WorkerTemplate"
              ? t("actions.apply")
              : t("actions.applyResource")}
      </Button>
    </div>
  );
}
