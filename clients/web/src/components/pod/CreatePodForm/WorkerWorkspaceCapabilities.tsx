"use client";

import { AlertMessage } from "@/components/ui/alert-message";
import type { InstalledSkill } from "@/lib/api";
import type { WorkerSpecDraft } from "@/lib/api/facade/podConnect";
import type { KnowledgeMountSelection } from "@/lib/api/facade/knowledgeBaseApi";
import type { WorkerCreateController } from "../hooks/workerCreateController";
import { EnvBundleMultiSelect } from "./EnvBundleMultiSelect";
import { KnowledgeBaseMountSelect } from "./KnowledgeBaseMountSelect";
import { SkillMultiSelect } from "./SkillMultiSelect";

interface WorkerWorkspaceCapabilitiesProps {
  controller: WorkerCreateController;
  t: (key: string) => string;
}

export function WorkerWorkspaceCapabilities({
  controller,
  t,
}: WorkerWorkspaceCapabilitiesProps) {
  const { draft } = controller.state;
  const skills = controller.skills.status === "ready" ? controller.skills.data : [];
  const bundles =
    controller.runtimeBundles.status === "ready"
      ? controller.runtimeBundles.data
      : [];
  const selectedSkillSlugs = draft.skill_ids.flatMap((id) => {
    const slug = skills.find((item) => item.id === id)?.slug;
    return slug ? [slug] : [];
  });
  const selectedBundleNames = draft.env_bundle_ids.flatMap((id) => {
    const name = bundles.find((item) => item.id === id)?.name;
    return name ? [name] : [];
  });
  const missingReferences =
    draft.skill_ids.length - selectedSkillSlugs.length +
    draft.env_bundle_ids.length - selectedBundleNames.length;

  return (
    <div className="space-y-5">
      {missingReferences > 0 && (
        <AlertMessage
          type="error"
          message={t("workerCreate.workspace.missingReferences")}
        />
      )}
      <KnowledgeBaseMountSelect
        selectedMounts={draft.knowledge_mounts.map((mount) => ({
          id: mount.knowledge_base_id,
          slug: "",
          mode: mount.mode === "rw" ? "rw" : "ro",
        }))}
        onChange={(mounts) =>
          controller.patchDraft({
            knowledge_mounts: workerKnowledgeMounts(mounts),
          })
        }
      />
      {controller.skills.status === "error" ? (
        <AlertMessage type="error" message={controller.skills.error} />
      ) : (
        <SkillMultiSelect
          skills={skills}
          selectedSlugs={selectedSkillSlugs}
          loading={
            controller.skills.status === "idle" ||
            controller.skills.status === "loading"
          }
          repositorySelected={Boolean(draft.repository_id)}
          onChange={(slugs) =>
            controller.patchDraft({
              skill_ids: slugs.map((slug) => requiredSkillID(slug, skills)),
            })
          }
          t={t}
        />
      )}
      <EnvBundleMultiSelect
        bundles={bundles}
        selectedBundleNames={selectedBundleNames}
        loading={
          controller.runtimeBundles.status === "idle" ||
          controller.runtimeBundles.status === "loading"
        }
        error={
          controller.runtimeBundles.status === "error"
            ? controller.runtimeBundles.error
            : null
        }
        onChange={(names) =>
          controller.patchDraft({
            env_bundle_ids: names.map((name) => requiredBundleID(name, bundles)),
          })
        }
        t={t}
      />
    </div>
  );
}

function workerKnowledgeMounts(
  mounts: KnowledgeMountSelection[],
): WorkerSpecDraft["knowledge_mounts"] {
  return mounts.map((mount) => {
    if (!mount.id) throw new Error("Knowledge base selection is missing its ID");
    return { knowledge_base_id: mount.id, mode: mount.mode };
  });
}

function requiredSkillID(
  slug: string,
  skills: InstalledSkill[],
): number {
  const skill = skills.find((item) => item.slug === slug);
  if (!skill) throw new Error(`Selected skill "${slug}" is unavailable`);
  return skill.id;
}

function requiredBundleID(
  name: string,
  bundles: Array<{ id: number; name: string }>,
): number {
  const bundle = bundles.find((item) => item.name === name);
  if (!bundle) throw new Error(`Selected environment bundle "${name}" is unavailable`);
  return bundle.id;
}
