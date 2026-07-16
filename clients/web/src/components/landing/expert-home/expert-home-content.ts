import workerRuntimeCatalog from "@/generated/worker-runtime-catalog.json";

export const solutionDomains = [
  { id: "enterprise-agent-supply", messageId: "enterpriseAgentSupply" },
  { id: "opc-incubation", messageId: "opcIncubation" },
  {
    id: "higher-education-digital-employees",
    messageId: "higherEducationDigitalEmployees",
  },
] as const;

export const workerTypes = workerRuntimeCatalog.workers.map(({ slug, name }) => ({ slug, name }));

export const capabilityGroups = [
  { id: "agent-factory", level: "implemented" },
  { id: "agent-market", level: "implemented" },
  { id: "collaboration-workspace", level: "implemented" },
  { id: "automation", level: "implemented" },
  { id: "governance", level: "implemented" },
] as const;

export const marketplaceApplications = [
  { slug: "software-delivery-expert", messageId: "softwareDelivery" },
  { slug: "multi-worker-orchestrator", messageId: "multiWorker" },
  { slug: "dual-repo-sync-expert", messageId: "dualRepo" },
] as const;

export type CapabilityLevel = "implemented" | "composable" | "planned";

export interface LocalizedSolution {
  id: (typeof solutionDomains)[number]["id"];
  title: string;
  description: string;
  chain: string;
  outcome: string;
  action: string;
}

export interface LocalizedCapability {
  id: (typeof capabilityGroups)[number]["id"];
  level: CapabilityLevel;
  title: string;
  description: string;
}

export interface LocalizedContentItem {
  id: string;
  title: string;
  description: string;
}

export interface LocalizedMarketApplication {
  slug: (typeof marketplaceApplications)[number]["slug"];
  title: string;
  description: string;
}

export interface LocalizedTrustItem extends LocalizedContentItem {
  status: string;
}
