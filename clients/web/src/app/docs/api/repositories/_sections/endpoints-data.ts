import type { EndpointRow } from "../../_components/endpoint-summary-table";
import type { DetailSpec } from "../../_components/endpoint-detail";
import {
  REPOS_LIST_JSON,
  REPO_JSON,
  BRANCHES_JSON,
  MERGE_REQUESTS_JSON,
} from "./response-samples";

export const summaryRows: EndpointRow[] = [
  {
    method: "GET",
    path: "/repositories",
    scope: "repos:read",
    descKey: "docs.api.repositories.endpoints.list",
  },
  {
    method: "GET",
    path: "/repositories/:id",
    scope: "repos:read",
    descKey: "docs.api.repositories.endpoints.get",
  },
  {
    method: "GET",
    path: "/repositories/:id/branches",
    scope: "repos:read",
    descKey: "docs.api.repositories.endpoints.branches",
  },
  {
    method: "GET",
    path: "/repositories/:id/merge-requests",
    scope: "repos:read",
    descKey: "docs.api.repositories.endpoints.mergeRequests",
  },
];

export const detailEndpoints: DetailSpec[] = [
  {
    method: "GET",
    path: "/repositories",
    descKey: "docs.api.repositories.details.listRepos.description",
    response: REPOS_LIST_JSON,
    tables: [],
  },
  {
    method: "GET",
    path: "/repositories/:id",
    descKey: "docs.api.repositories.details.getRepo.description",
    response: REPO_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "id", type: "integer", required: true, descKey: "docs.api.repositories.details.getRepo.params.id" },
        ],
      },
    ],
  },
  {
    method: "GET",
    path: "/repositories/:id/branches",
    descKey: "docs.api.repositories.details.listBranches.description",
    response: BRANCHES_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "id", type: "integer", required: true, descKey: "docs.api.repositories.details.listBranches.params.id" },
        ],
      },
      {
        kind: "query",
        rows: [
          { name: "access_token", type: "string", required: true, descKey: "docs.api.repositories.details.listBranches.params.access_token" },
        ],
      },
    ],
  },
  {
    method: "GET",
    path: "/repositories/:id/merge-requests",
    descKey: "docs.api.repositories.details.listMergeRequests.description",
    response: MERGE_REQUESTS_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "id", type: "integer", required: true, descKey: "docs.api.repositories.details.listMergeRequests.params.id" },
        ],
      },
      {
        kind: "query",
        rows: [
          { name: "branch", type: "string", required: false, descKey: "docs.api.repositories.details.listMergeRequests.params.branch" },
          { name: "state", type: "string", required: false, descKey: "docs.api.repositories.details.listMergeRequests.params.state" },
        ],
      },
    ],
  },
];
