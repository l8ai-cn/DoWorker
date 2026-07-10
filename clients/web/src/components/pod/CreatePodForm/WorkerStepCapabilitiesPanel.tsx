"use client";

import { BookOpen, Braces, Sparkles, Wallet } from "lucide-react";
import { ExpertPickerSection } from "@/components/experts/ExpertPickerSection";
import { CustomEnvSection } from "@/components/settings/CustomEnvSection";
import { CapabilityConfigPanel } from "./CapabilityConfigPanel";
import { KnowledgeBaseMountSelect } from "./KnowledgeBaseMountSelect";
import { SkillMultiSelect } from "./SkillMultiSelect";
import { WorkerBudgetSection } from "./WorkerBudgetSection";
import type { StepPanelsProps } from "./WorkerStepPanels";

const EMPTY_DECLARED_KEYS: Set<string> = new Set();

export function WorkerStepCapabilitiesPanel({
  form,
  showPerpetual,
  setSelectedRunnerId,
  initialExpertSlug,
  t,
}: StepPanelsProps) {
  return (
    <div className="space-y-4 animate-in fade-in duration-200">
      <p className="text-sm text-muted-foreground">{t("ide.createPod.stepCapabilitiesPanelHint")}</p>

      {showPerpetual && (
        <ExpertPickerSection
          form={form}
          setSelectedRunnerId={setSelectedRunnerId}
          initialExpertSlug={initialExpertSlug}
        />
      )}

      <CapabilityConfigPanel
        icon={BookOpen}
        title={t("ide.createPod.knowledgeConfigTitle")}
        description={t("ide.createPod.knowledgeConfigDescription")}
        testId="worker-knowledge-select"
      >
        <KnowledgeBaseMountSelect
          embedded
          selectedMounts={form.selectedKnowledgeMounts}
          onChange={form.setSelectedKnowledgeMounts}
        />
      </CapabilityConfigPanel>
      <CapabilityConfigPanel
        icon={Sparkles}
        title={t("ide.createPod.skillConfigTitle")}
        description={t("ide.createPod.skillConfigDescription")}
        testId="worker-skills-select"
      >
        <div className="space-y-3">
          {!form.selectedRepository && (
            <p className="text-sm text-muted-foreground">
              {t("ide.createPod.skillsRequireRepositoryHint")}
            </p>
          )}
          <SkillMultiSelect
            embedded
            skills={form.repoSkills}
            selectedSlugs={form.selectedSkillSlugs}
            onChange={form.setSelectedSkillSlugs}
            loading={form.loadingSkills}
            repositorySelected={Boolean(form.selectedRepository)}
            t={t}
          />
        </div>
      </CapabilityConfigPanel>
      <CapabilityConfigPanel
        icon={Wallet}
        title={t("ide.createPod.budgetConfigTitle")}
        description={t("ide.createPod.budgetConfigDescription")}
        testId="worker-budget-select"
      >
        <WorkerBudgetSection
          tokenBudget={form.tokenBudget}
          onChange={form.setTokenBudget}
          t={t}
        />
      </CapabilityConfigPanel>
      <CapabilityConfigPanel
        icon={Braces}
        title={t("ide.createPod.customEnvTitle")}
        description={t("ide.createPod.customEnvDescription")}
        testId="worker-custom-env"
      >
        <CustomEnvSection
          entries={form.customEnv}
          declaredKeys={EMPTY_DECLARED_KEYS}
          onChange={form.setCustomEnv}
          isEditing={false}
          valueType="text"
          t={t}
        />
        <p className="mt-3 text-xs text-muted-foreground">
          {t("ide.createPod.customEnvPlaintextWarning")}
        </p>
      </CapabilityConfigPanel>
    </div>
  );
}
