import { getApiBaseUrl } from "./env";

export interface PublicMarketSkill {
  id: number;
  slug: string;
  display_name: string;
  description: string;
  license: string;
  category: string;
  version: number;
  package_size: number;
  updated_at: string;
}

export interface PublicMarketSkillsResponse {
  items: PublicMarketSkill[];
  total: number;
}

export async function fetchPublicMarketSkills(): Promise<PublicMarketSkillsResponse> {
  const resp = await fetch(`${resolvePublicApiBase()}/api/v1/public/market/skills`, {
    cache: "no-store",
  });
  if (!resp.ok) {
    throw new Error(`Failed to load public skill market: HTTP ${resp.status}`);
  }
  return resp.json() as Promise<PublicMarketSkillsResponse>;
}

function resolvePublicApiBase(): string {
  const configured = getApiBaseUrl();
  if (configured) return configured;
  if (typeof window !== "undefined") return "";
  return process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000";
}
