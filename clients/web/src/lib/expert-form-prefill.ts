import type { Expert } from "@/lib/api/expertApi";
import { parseExpertKnowledgeMounts } from "@/lib/api/expertApi";
import type { CreatePodFormState } from "@/components/pod/hooks/useCreatePodFormTypes";
import type { PodMode } from "@/lib/pod-modes";

export function applyExpertToForm(
  expert: Expert,
  form: CreatePodFormState,
  setSelectedRunnerId: (id: number | null) => void,
): void {
  form.setSelectedAgent(expert.agent_slug);
  if (expert.runner_id) setSelectedRunnerId(expert.runner_id);
  if (expert.repository_id) form.setSelectedRepository(expert.repository_id);
  if (expert.branch_name) form.setSelectedBranch(expert.branch_name);
  if (expert.prompt) form.setPrompt(expert.prompt);
  form.setInteractionMode(expert.interaction_mode as PodMode);
  form.setPerpetual(expert.perpetual);
  form.setSelectedSkillSlugs(expert.skill_slugs ?? []);
  form.setSelectedRuntimeBundleNames(expert.used_env_bundles ?? []);
  const mounts = parseExpertKnowledgeMounts(expert.knowledge_mounts).map((m) => ({
    slug: m.slug,
    mode: (m.mode === "rw" ? "rw" : "ro") as "ro" | "rw",
  }));
  form.setSelectedKnowledgeMounts(mounts);
  if (expert.agentfile_layer?.trim()) {
    form.setRawLayerMode(true);
    form.setRawLayerText(expert.agentfile_layer);
  }
}
