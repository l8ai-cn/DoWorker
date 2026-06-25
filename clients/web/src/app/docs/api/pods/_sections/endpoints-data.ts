import type { EndpointRow } from "../../_components/endpoint-summary-table";
import type { DetailSpec } from "../../_components/endpoint-detail";
import {
  PODS_LIST_JSON,
  POD_JSON,
  POD_CREATE_JSON,
  TERMINATE_JSON,
} from "./response-samples";

export const summaryRows: EndpointRow[] = [
  {
    method: "GET",
    path: "/pods",
    scope: "pods:read",
    descKey: "docs.api.pods.endpoints.list",
  },
  {
    method: "GET",
    path: "/pods/:key",
    scope: "pods:read",
    descKey: "docs.api.pods.endpoints.get",
  },
  {
    method: "POST",
    path: "/pods",
    scope: "pods:write",
    descKey: "docs.api.pods.endpoints.create",
  },
  {
    method: "POST",
    path: "/pods/:key/terminate",
    scope: "pods:write",
    descKey: "docs.api.pods.endpoints.terminate",
  },
];

export const detailEndpoints: DetailSpec[] = [
  {
    method: "GET",
    path: "/pods",
    descKey: "docs.api.pods.details.listPods.description",
    response: PODS_LIST_JSON,
    tables: [
      {
        kind: "query",
        withDefault: true,
        rows: [
          { name: "status", type: "string", required: false, descKey: "docs.api.pods.details.listPods.params.status" },
          { name: "limit", type: "integer", required: false, default: "20", descKey: "docs.api.pods.details.listPods.params.limit" },
          { name: "offset", type: "integer", required: false, default: "0", descKey: "docs.api.pods.details.listPods.params.offset" },
        ],
      },
    ],
  },
  {
    method: "GET",
    path: "/pods/:key",
    descKey: "docs.api.pods.details.getPod.description",
    response: POD_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "key", type: "string", required: true, descKey: "docs.api.pods.details.getPod.params.key" },
        ],
      },
    ],
  },
  {
    method: "POST",
    path: "/pods",
    descKey: "docs.api.pods.details.createPod.description",
    response: POD_CREATE_JSON,
    tables: [
      {
        kind: "body",
        rows: [
          { name: "agent_slug", type: "string", required: true, descKey: "docs.api.pods.details.createPod.fields.agent_slug" },
          { name: "agentfile_layer", type: "string", required: false, desc: "AgentFile Layer — SSOT for PROMPT, MODE, CONFIG, REPO, BRANCH, CREDENTIAL" },
          { name: "runner_id", type: "integer", required: false, descKey: "docs.api.pods.details.createPod.fields.runner_id" },
          { name: "ticket_slug", type: "string", required: false, descKey: "docs.api.pods.details.createPod.fields.ticket_slug" },
          { name: "alias", type: "string", required: false, desc: "Display name for the pod (max 100 chars)" },
        ],
      },
    ],
  },
  {
    method: "POST",
    path: "/pods/:key/terminate",
    descKey: "docs.api.pods.details.terminatePod.description",
    response: TERMINATE_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "key", type: "string", required: true, descKey: "docs.api.pods.details.terminatePod.params.key" },
        ],
      },
    ],
  },
];
