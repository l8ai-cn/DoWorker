import { useEffect, useState } from "react";
import { listMarketSkills } from "@/lib/api/facade/marketExtension";
import { listRepoSkills } from "@/lib/api/facade/repoSkillExtension";
import { readCurrentOrg } from "@/stores/auth";
import type { WorkerSkillOption } from "../CreatePodForm/workerSkillOption";

interface WorkerSkillsState {
  skills: WorkerSkillOption[];
  loading: boolean;
  error: string | null;
}

export function useWorkerSkills(repositoryId: number | null): WorkerSkillsState {
  const orgSlug = readCurrentOrg()?.slug ?? "";
  const requestKey = orgSlug
    ? `${orgSlug}:${repositoryId ?? "catalog"}`
    : "";
  const [loaded, setLoaded] = useState<{ key: string; state: WorkerSkillsState }>({
    key: "",
    state: { skills: [], loading: false, error: null },
  });

  useEffect(() => {
    if (!requestKey) {
      return;
    }

    let cancelled = false;
    const load = repositoryId
      ? loadRepositorySkills(orgSlug, repositoryId)
      : loadCatalogSkills(orgSlug);
    void load
      .then((skills) => {
        if (cancelled) return;
        setLoaded({
          key: requestKey,
          state: { skills, loading: false, error: null },
        });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        setLoaded({
          key: requestKey,
          state: {
            skills: [],
            loading: false,
            error: error instanceof Error ? error.message : "Failed to load skills",
          },
        });
      });

    return () => {
      cancelled = true;
    };
  }, [orgSlug, repositoryId, requestKey]);

  return loaded.key === requestKey
    ? loaded.state
    : { skills: [], loading: Boolean(requestKey), error: null };
}

async function loadCatalogSkills(orgSlug: string): Promise<WorkerSkillOption[]> {
  const response = await listMarketSkills(orgSlug);
  return response.items
    .filter((skill) => skill.is_active)
    .map((skill) => ({
      id: skill.id,
      slug: skill.slug,
      scope: "org" as const,
    }));
}

async function loadRepositorySkills(
  orgSlug: string,
  repositoryId: number,
): Promise<WorkerSkillOption[]> {
  const [orgResponse, userResponse] = await Promise.all([
    listRepoSkills(orgSlug, repositoryId, { scope: "org" }),
    listRepoSkills(orgSlug, repositoryId, { scope: "user" }),
  ]);
  const options = [...orgResponse.items, ...userResponse.items]
    .filter((skill) => skill.is_enabled)
    .map((skill) => ({
      id: skill.market_item_id ?? skill.id,
      slug: skill.slug,
      scope: skill.scope,
    }));
  const seen = new Set<string>();
  return options.filter((skill) => {
    if (seen.has(skill.slug)) return false;
    seen.add(skill.slug);
    return skill.id > 0;
  });
}
