import { describe, expect, it } from "vitest";
import { summaryRows } from "./endpoints-data";

describe("Workflow API documentation", () => {
  it("documents runtime endpoints without legacy definition CRUD", () => {
    expect(summaryRows.map(({ method, path }) => `${method} ${path}`)).toEqual([
      "GET /workflows",
      "GET /workflows/:slug",
      "POST /workflows/:slug/enable",
      "POST /workflows/:slug/disable",
      "POST /workflows/:slug/trigger",
      "GET /workflows/:slug/runs",
      "GET /workflows/:slug/runs/:run_id",
      "POST /workflows/:slug/runs/:run_id/cancel",
    ]);
  });
});
