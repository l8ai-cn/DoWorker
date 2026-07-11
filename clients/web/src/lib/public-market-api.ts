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

export interface PublicMarketApplication {
  slug: string;
  name: string;
  summary: string;
  description: string;
  category: string;
  icon: "rocket" | "network" | "git-compare";
  agent_slug: string;
  skill_slugs: string[];
  tags: string[];
  outcomes: string[];
  version: number;
  featured: boolean;
}

export interface PublicMarketApplicationsResponse {
  items: PublicMarketApplication[];
  total: number;
}

export async function fetchPublicMarketApplications(): Promise<PublicMarketApplicationsResponse> {
  const resp = await fetch(`${resolvePublicApiBase()}/api/v1/public/market/applications`, {
    cache: "no-store",
  });
  if (!resp.ok) {
    throw new Error(`Failed to load public application market: HTTP ${resp.status}`);
  }
  return resp.json() as Promise<PublicMarketApplicationsResponse>;
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
  const domain = process.env.NEXT_PUBLIC_PRIMARY_DOMAIN;
  if (domain && !domain.startsWith("__")) {
    const protocol = process.env.NEXT_PUBLIC_USE_HTTPS === "true" ? "https" : "http";
    return `${protocol}://${domain}`;
  }
  return "http://localhost:10000";
}
