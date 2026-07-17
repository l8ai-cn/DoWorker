"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { AlertMessage } from "@/components/ui/alert-message";
import { Spinner } from "@/components/ui/spinner";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type { WorkerSpecDraft } from "@/lib/api/facade/podConnect";
import type { AsyncState } from "../hooks/workerCreateDraft";
import { modelResourceLabel } from "./workerModelResources";
import { WorkerRuntimeSelectField } from "./WorkerRuntimeSelectField";

interface WorkerPrimaryModelFieldProps {
  state: AsyncState<EffectiveResource[]>;
  draft: WorkerSpecDraft;
  onPatch: (patch: Partial<WorkerSpecDraft>) => void;
  t: (key: string) => string;
}

export function WorkerPrimaryModelField(props: WorkerPrimaryModelFieldProps) {
  const params = useParams() ?? {};
  const orgSlug = String(params.org ?? "");
  if (props.state.status === "idle" || props.state.status === "loading") {
    return (
      <div data-runtime-field="model" data-testid="worker-runtime-field-model">
        <Spinner className="my-4" />
      </div>
    );
  }
  if (props.state.status === "error") {
    return (
      <div data-runtime-field="model" data-testid="worker-runtime-field-model">
        <AlertMessage type="error" message={props.state.error} />
      </div>
    );
  }
  if (props.state.data.length === 0) {
    return (
      <div data-runtime-field="model" data-testid="worker-runtime-field-model">
        <AlertMessage type="error" message={props.t("ide.createPod.noModelResourcesAvailableHint")} />
        <ModelResourceSettingsLink orgSlug={orgSlug} t={props.t} />
      </div>
    );
  }
  return (
    <div data-testid="worker-runtime-field-model">
      <WorkerRuntimeSelectField
        field="model"
        label={props.t("workerCreate.runtime.model")}
        value={props.draft.model_resource_id > 0 ? String(props.draft.model_resource_id) : ""}
        options={props.state.data.map((item) => ({
          value: String(item.resource?.id ?? 0),
          label: modelResourceLabel(item),
          selectable: item.selectable,
          blockingReason: item.blockingReason,
        }))}
        onChange={(value) => props.onPatch({ model_resource_id: Number(value) })}
      />
      <ModelResourceSettingsLink orgSlug={orgSlug} t={props.t} />
    </div>
  );
}

function ModelResourceSettingsLink({
  orgSlug,
  t,
}: {
  orgSlug: string;
  t: (key: string) => string;
}) {
  if (!orgSlug) return null;
  return (
    <Link
      href={`/${orgSlug}/settings?tab=ai-resources`}
      className="mt-1 inline-block text-xs font-medium text-primary hover:underline"
    >
      {t("ide.createPod.manageModelResources")}
    </Link>
  );
}
