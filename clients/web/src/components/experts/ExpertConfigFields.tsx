"use client";

import { useTranslations } from "next-intl";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { RunnerSelect } from "@/components/pod/CreatePodForm/RunnerSelect";
import { RepositorySelect, BranchInput } from "@/components/pod/CreatePodForm/RepositorySelect";
import { EnvBundleMultiSelect } from "@/components/pod/CreatePodForm/EnvBundleMultiSelect";
import { KnowledgeBaseMountSelect } from "@/components/pod/CreatePodForm/KnowledgeBaseMountSelect";
import { AutomationLevelSelect } from "@/components/pod/CreatePodForm/AutomationLevelSelect";
import { isValidConfigOverrides, type ExpertFormState } from "./expertFormModel";
import { useExpertConfigData } from "./useExpertConfigData";

interface Props {
  open: boolean;
  form: ExpertFormState;
  patch: (next: Partial<ExpertFormState>) => void;
}

export function ExpertConfigFields({ open, form, patch }: Props) {
  const t = useTranslations();
  const te = useTranslations("experts.edit");
  const { runners, repositories, envBundles, loadingBundles } = useExpertConfigData(
    open,
    form.agentSlug,
  );
  const configInvalid = !isValidConfigOverrides(form.configOverrides);

  return (
    <div className="space-y-4 border-t border-border/50 pt-4">
      <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {te("configSectionLabel")}
      </p>

      <AutomationLevelSelect
        value={form.automationLevel}
        onChange={(level) => patch({ automationLevel: level })}
        t={t}
      />

      <RunnerSelect
        runners={runners}
        selectedRunnerId={form.runnerId}
        onSelect={(id) => patch({ runnerId: id })}
        t={t}
      />

      <RepositorySelect
        repositories={repositories}
        selectedRepositoryId={form.repositoryId}
        onSelect={(id) => patch({ repositoryId: id })}
        t={t}
      />
      {form.repositoryId != null && (
        <BranchInput value={form.branchName} onChange={(v) => patch({ branchName: v })} t={t} />
      )}

      <EnvBundleMultiSelect
        bundles={envBundles.filter((b) => b.kind === "runtime")}
        selectedBundleNames={form.usedEnvBundles}
        onChange={(names) => patch({ usedEnvBundles: names })}
        loading={loadingBundles}
        t={t}
      />

      <div>
        <Label className="mb-2 block text-sm font-medium">
          {t("ide.createPod.knowledgeBases")}
        </Label>
        <KnowledgeBaseMountSelect
          selectedMounts={form.knowledgeMounts}
          onChange={(mounts) => patch({ knowledgeMounts: mounts })}
          embedded
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="expert-agentfile">{te("agentfileLayerLabel")}</Label>
        <Textarea
          id="expert-agentfile"
          value={form.agentfileLayer}
          onChange={(e) => patch({ agentfileLayer: e.target.value })}
          placeholder={te("agentfileLayerPlaceholder")}
          className="min-h-[80px] font-mono text-xs"
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="expert-config-overrides">{te("configOverridesLabel")}</Label>
        <Textarea
          id="expert-config-overrides"
          value={form.configOverrides}
          onChange={(e) => patch({ configOverrides: e.target.value })}
          placeholder={te("configOverridesPlaceholder")}
          className={`min-h-[72px] font-mono text-xs ${configInvalid ? "border-destructive" : ""}`}
          aria-invalid={configInvalid}
        />
        {configInvalid && <p className="text-xs text-destructive">{te("configOverridesInvalid")}</p>}
      </div>
    </div>
  );
}
