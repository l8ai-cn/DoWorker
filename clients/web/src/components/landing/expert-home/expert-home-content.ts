import workerRuntimeCatalog from "@/generated/worker-runtime-catalog.json";

export const solutionDomains = [
  { id: "cross-border-commerce", messageId: "crossBorderCommerce" },
  { id: "ai-education", messageId: "aiEducation" },
  { id: "digital-employees", messageId: "aiPartners" },
  { id: "marketplace", messageId: "marketplace" },
] as const;

export const workerTypes = workerRuntimeCatalog.workers.map(({ slug, name }) => ({ slug, name }));

export const capabilityGroups = [
  { id: "programming", level: "implemented" },
  { id: "research", level: "composable" },
  { id: "documents", level: "composable" },
  { id: "office", level: "composable" },
  { id: "content", level: "composable" },
  { id: "screenwriting", level: "composable" },
  { id: "image", level: "composable" },
  { id: "audio", level: "composable" },
  { id: "video", level: "composable" },
  { id: "data", level: "composable" },
  { id: "education", level: "composable" },
  { id: "industryMarket", level: "planned" },
] as const;

export const marketplaceApplications = [
  { slug: "software-delivery-expert", messageId: "softwareDelivery" },
  { slug: "multi-worker-orchestrator", messageId: "multiWorker" },
  { slug: "dual-repo-sync-expert", messageId: "dualRepo" },
] as const;

export type CapabilityLevel = (typeof capabilityGroups)[number]["level"];

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
