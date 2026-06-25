import type { EndpointRow } from "../../_components/endpoint-summary-table";
import type { DetailSpec } from "../../_components/endpoint-detail";
import {
  RUNNERS_LIST_JSON,
  RUNNER_JSON,
  RUNNER_PODS_JSON,
} from "./response-samples";

export const summaryRows: EndpointRow[] = [
  {
    method: "GET",
    path: "/runners",
    scope: "runners:read",
    descKey: "docs.api.runners.endpoints.list",
  },
  {
    method: "GET",
    path: "/runners/:id",
    scope: "runners:read",
    descKey: "docs.api.runners.endpoints.get",
  },
  {
    method: "GET",
    path: "/runners/available",
    scope: "runners:read",
    descKey: "docs.api.runners.endpoints.available",
  },
  {
    method: "GET",
    path: "/runners/:id/pods",
    scope: "runners:read",
    descKey: "docs.api.runners.endpoints.pods",
  },
];

export const detailEndpoints: DetailSpec[] = [
  {
    method: "GET",
    path: "/runners",
    descKey: "docs.api.runners.details.listRunners.description",
    response: RUNNERS_LIST_JSON,
    tables: [],
  },
  {
    method: "GET",
    path: "/runners/:id",
    descKey: "docs.api.runners.details.getRunner.description",
    response: RUNNER_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "id", type: "integer", required: true, descKey: "docs.api.runners.details.getRunner.params.id" },
        ],
      },
    ],
  },
  {
    method: "GET",
    path: "/runners/available",
    descKey: "docs.api.runners.details.availableRunners.description",
    response: RUNNERS_LIST_JSON,
    tables: [],
  },
  {
    method: "GET",
    path: "/runners/:id/pods",
    descKey: "docs.api.runners.details.runnerPods.description",
    response: RUNNER_PODS_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "id", type: "integer", required: true, descKey: "docs.api.runners.details.runnerPods.params.id" },
        ],
      },
      {
        kind: "query",
        withDefault: true,
        rows: [
          { name: "status", type: "string", required: false, descKey: "docs.api.runners.details.runnerPods.params.status" },
          { name: "limit", type: "integer", required: false, default: "50", descKey: "docs.api.runners.details.runnerPods.params.limit" },
          { name: "offset", type: "integer", required: false, default: "0", descKey: "docs.api.runners.details.runnerPods.params.offset" },
        ],
      },
    ],
  },
];
