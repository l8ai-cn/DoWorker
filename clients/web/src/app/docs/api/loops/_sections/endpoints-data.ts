import type { EndpointRow } from "../../_components/endpoint-summary-table";

export const summaryRows: EndpointRow[] = [
  {
    method: "GET",
    path: "/loops",
    scope: "loops:read",
    descKey: "docs.api.loops.endpoints.list",
  },
  {
    method: "POST",
    path: "/loops",
    scope: "loops:write",
    descKey: "docs.api.loops.endpoints.create",
  },
  {
    method: "GET",
    path: "/loops/:slug",
    scope: "loops:read",
    descKey: "docs.api.loops.endpoints.get",
  },
  {
    method: "PUT",
    path: "/loops/:slug",
    scope: "loops:write",
    descKey: "docs.api.loops.endpoints.update",
  },
  {
    method: "DELETE",
    path: "/loops/:slug",
    scope: "loops:write",
    descKey: "docs.api.loops.endpoints.delete",
  },
  {
    method: "POST",
    path: "/loops/:slug/enable",
    scope: "loops:write",
    descKey: "docs.api.loops.endpoints.enable",
  },
  {
    method: "POST",
    path: "/loops/:slug/disable",
    scope: "loops:write",
    descKey: "docs.api.loops.endpoints.disable",
  },
  {
    method: "POST",
    path: "/loops/:slug/trigger",
    scope: "loops:write",
    descKey: "docs.api.loops.endpoints.trigger",
  },
  {
    method: "GET",
    path: "/loops/:slug/runs",
    scope: "loops:read",
    descKey: "docs.api.loops.endpoints.listRuns",
  },
  {
    method: "GET",
    path: "/loops/:slug/runs/:run_id",
    scope: "loops:read",
    descKey: "docs.api.loops.endpoints.getRun",
  },
  {
    method: "POST",
    path: "/loops/:slug/runs/:run_id/cancel",
    scope: "loops:write",
    descKey: "docs.api.loops.endpoints.cancelRun",
  },
];

export const createLoopFields = [
  "name",
  "description",
  "agent_slug",
  "custom_agent_slug",
  "prompt_template",
  "prompt_variables",
  "repository_id",
  "runner_id",
  "branch_name",
  "execution_mode",
  "cron_expression",
  "sandbox_strategy",
  "session_persistence",
  "concurrency_policy",
  "max_concurrent_runs",
  "timeout_minutes",
  "callback_url",
  "autopilot_config",
] as const;
