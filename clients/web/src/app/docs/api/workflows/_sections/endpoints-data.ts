import type { EndpointRow } from "../../_components/endpoint-summary-table";

export const summaryRows: EndpointRow[] = [
  {
    method: "GET",
    path: "/workflows",
    scope: "workflows:read",
    descKey: "docs.api.workflows.endpoints.list",
  },
  {
    method: "POST",
    path: "/workflows",
    scope: "workflows:write",
    descKey: "docs.api.workflows.endpoints.create",
  },
  {
    method: "GET",
    path: "/workflows/:slug",
    scope: "workflows:read",
    descKey: "docs.api.workflows.endpoints.get",
  },
  {
    method: "PUT",
    path: "/workflows/:slug",
    scope: "workflows:write",
    descKey: "docs.api.workflows.endpoints.update",
  },
  {
    method: "DELETE",
    path: "/workflows/:slug",
    scope: "workflows:write",
    descKey: "docs.api.workflows.endpoints.delete",
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

export const createWorkflowFields = [
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
