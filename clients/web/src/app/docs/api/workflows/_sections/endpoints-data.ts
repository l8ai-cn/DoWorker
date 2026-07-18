import type { EndpointRow } from "../../_components/endpoint-summary-table";

export const summaryRows: EndpointRow[] = [
  {
    method: "GET",
    path: "/workflows",
    scope: "workflows:read",
    descKey: "docs.api.workflows.endpoints.list",
  },
  {
    method: "GET",
    path: "/workflows/:slug",
    scope: "workflows:read",
    descKey: "docs.api.workflows.endpoints.get",
  },
  {
    method: "POST",
    path: "/workflows/:slug/enable",
    scope: "workflows:write",
    descKey: "docs.api.workflows.endpoints.enable",
  },
  {
    method: "POST",
    path: "/workflows/:slug/disable",
    scope: "workflows:write",
    descKey: "docs.api.workflows.endpoints.disable",
  },
  {
    method: "POST",
    path: "/workflows/:slug/trigger",
    scope: "workflows:write",
    descKey: "docs.api.workflows.endpoints.trigger",
  },
  {
    method: "GET",
    path: "/workflows/:slug/runs",
    scope: "workflows:read",
    descKey: "docs.api.workflows.endpoints.listRuns",
  },
  {
    method: "GET",
    path: "/workflows/:slug/runs/:run_id",
    scope: "workflows:read",
    descKey: "docs.api.workflows.endpoints.getRun",
  },
  {
    method: "POST",
    path: "/workflows/:slug/runs/:run_id/cancel",
    scope: "workflows:write",
    descKey: "docs.api.workflows.endpoints.cancelRun",
  },
];
