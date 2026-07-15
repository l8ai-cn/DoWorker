import { describe, expect, it } from "vitest";
import workerRuntimeCatalog from "@/generated/worker-runtime-catalog.json";

import {
  capabilityGroups,
  marketplaceApplications,
  solutionDomains,
  workerTypes,
} from "../expert-home-content";

describe("expert homepage content contract", () => {
  it("exposes the four approved solution domains", () => {
    expect(solutionDomains.map(({ id }) => id)).toEqual([
      "cross-border-commerce",
      "ai-education",
      "digital-employees",
      "marketplace",
    ]);
  });

  it("lists every formal worker type from the product catalog", () => {
    expect(workerTypes).toEqual(
      workerRuntimeCatalog.workers.map(({ slug, name }) => ({ slug, name })),
    );
  });

  it("keeps the marketplace aligned with implemented expert applications", () => {
    expect(marketplaceApplications.map(({ slug }) => slug)).toEqual([
      "software-delivery-expert",
      "multi-worker-orchestrator",
      "dual-repo-sync-expert",
    ]);
  });

  it("labels capabilities by implementation maturity", () => {
    expect(new Set(capabilityGroups.map(({ level }) => level))).toEqual(
      new Set(["implemented", "composable", "planned"]),
    );
  });
});
