import type { CatalogSkill } from "@/lib/api";

export type SkillCatalogGroup =
  | { kind: "tag"; tag: string; skills: CatalogSkill[] }
  | { kind: "untagged"; skills: CatalogSkill[] };

export function groupCatalogSkills(skills: CatalogSkill[]): SkillCatalogGroup[] {
  const tagged = new Map<string, CatalogSkill[]>();
  const untagged: CatalogSkill[] = [];

  for (const skill of skills) {
    if (skill.tags.length === 0) {
      untagged.push(skill);
      continue;
    }
    for (const tag of skill.tags) {
      tagged.set(tag, [...(tagged.get(tag) ?? []), skill]);
    }
  }

  const groups: SkillCatalogGroup[] = [...tagged.entries()]
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([tag, groupSkills]) => ({ kind: "tag", tag, skills: groupSkills }));
  if (untagged.length > 0) groups.push({ kind: "untagged", skills: untagged });
  return groups;
}
