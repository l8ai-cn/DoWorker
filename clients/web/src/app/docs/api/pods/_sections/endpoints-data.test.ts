import { describe, expect, it } from "vitest";
import { detailEndpoints } from "./endpoints-data";

describe("Pod API documentation", () => {
  it("documents POST /pods as lineage-only resume", () => {
    const endpoint = detailEndpoints.find(
      ({ method, path }) => method === "POST" && path === "/pods",
    );
    const body = endpoint?.tables?.find(({ kind }) => kind === "body");

    expect(body?.rows.map(({ name }) => name)).toEqual([
      "source_pod_key",
      "resume_agent_session",
      "ticket_slug",
      "cols",
      "rows",
      "queue_if_offline",
      "queue_ttl_minutes",
    ]);
  });
});
