import type { EndpointRow } from "../../_components/endpoint-summary-table";
import type { DetailSpec } from "../../_components/endpoint-detail";
import {
  TICKET_JSON,
  LIST_JSON,
  BOARD_JSON,
  STATUS_JSON,
  DELETE_JSON,
} from "./response-samples";

export const summaryRows: EndpointRow[] = [
  {
    method: "GET",
    path: "/tickets",
    scope: "tickets:read",
    descKey: "docs.api.tickets.endpoints.list",
  },
  {
    method: "GET",
    path: "/tickets/board",
    scope: "tickets:read",
    descKey: "docs.api.tickets.endpoints.board",
  },
  {
    method: "GET",
    path: "/tickets/:slug",
    scope: "tickets:read",
    descKey: "docs.api.tickets.endpoints.get",
  },
  {
    method: "POST",
    path: "/tickets",
    scope: "tickets:write",
    descKey: "docs.api.tickets.endpoints.create",
  },
  {
    method: "PUT",
    path: "/tickets/:slug",
    scope: "tickets:write",
    descKey: "docs.api.tickets.endpoints.update",
  },
  {
    method: "PATCH",
    path: "/tickets/:slug/status",
    scope: "tickets:write",
    descKey: "docs.api.tickets.endpoints.updateStatus",
  },
  {
    method: "DELETE",
    path: "/tickets/:slug",
    scope: "tickets:write",
    descKey: "docs.api.tickets.endpoints.delete",
  },
];

export const detailEndpoints: DetailSpec[] = [
  {
    method: "GET",
    path: "/tickets",
    descKey: "docs.api.tickets.details.listTickets.description",
    response: LIST_JSON,
    tables: [
      {
        kind: "query",
        withDefault: true,
        rows: [
          { name: "repository_id", type: "integer", required: false, descKey: "docs.api.tickets.details.listTickets.params.repository_id" },
          { name: "status", type: "string", required: false, descKey: "docs.api.tickets.details.listTickets.params.status" },
          { name: "type", type: "string", required: false, descKey: "docs.api.tickets.details.listTickets.params.type" },
          { name: "assignee_id", type: "integer", required: false, descKey: "docs.api.tickets.details.listTickets.params.assignee_id" },
          { name: "labels", type: "string", required: false, descKey: "docs.api.tickets.details.listTickets.params.labels" },
          { name: "limit", type: "integer", required: false, default: "20", descKey: "docs.api.tickets.details.listTickets.params.limit" },
          { name: "offset", type: "integer", required: false, default: "0", descKey: "docs.api.tickets.details.listTickets.params.offset" },
        ],
      },
    ],
  },
  {
    method: "GET",
    path: "/tickets/board",
    descKey: "docs.api.tickets.details.getBoard.description",
    response: BOARD_JSON,
    tables: [
      {
        kind: "query",
        rows: [
          { name: "repository_id", type: "integer", required: false, descKey: "docs.api.tickets.details.getBoard.params.repository_id" },
        ],
      },
    ],
  },
  {
    method: "GET",
    path: "/tickets/:slug",
    descKey: "docs.api.tickets.details.getTicket.description",
    response: TICKET_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "slug", type: "string", required: true, descKey: "docs.api.tickets.details.getTicket.params.slug" },
        ],
      },
    ],
  },
  {
    method: "POST",
    path: "/tickets",
    descKey: "docs.api.tickets.details.createTicket.description",
    response: TICKET_JSON,
    tables: [
      {
        kind: "body",
        rows: [
          { name: "type", type: "string", required: true, descKey: "docs.api.tickets.details.createTicket.fields.type" },
          { name: "title", type: "string", required: true, descKey: "docs.api.tickets.details.createTicket.fields.title" },
          { name: "priority", type: "string", required: false, descKey: "docs.api.tickets.details.createTicket.fields.priority" },
          { name: "status", type: "string", required: false, descKey: "docs.api.tickets.details.createTicket.fields.status" },
          { name: "assignee_id", type: "integer", required: false, descKey: "docs.api.tickets.details.createTicket.fields.assignee_id" },
          { name: "repository_id", type: "integer", required: false, descKey: "docs.api.tickets.details.createTicket.fields.repository_id" },
          { name: "labels", type: "string[]", required: false, descKey: "docs.api.tickets.details.createTicket.fields.labels" },
          { name: "parent_slug", type: "string", required: false, descKey: "docs.api.tickets.details.createTicket.fields.parent_slug" },
        ],
      },
    ],
  },
  {
    method: "PUT",
    path: "/tickets/:slug",
    descKey: "docs.api.tickets.details.updateTicket.description",
    response: TICKET_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "slug", type: "string", required: true, descKey: "docs.api.tickets.details.updateTicket.params.slug" },
        ],
      },
      {
        kind: "body",
        rows: [
          { name: "title", type: "string", required: false, descKey: "docs.api.tickets.details.updateTicket.fields.title" },
          { name: "type", type: "string", required: false, descKey: "docs.api.tickets.details.updateTicket.fields.type" },
          { name: "priority", type: "string", required: false, descKey: "docs.api.tickets.details.updateTicket.fields.priority" },
          { name: "status", type: "string", required: false, descKey: "docs.api.tickets.details.updateTicket.fields.status" },
          { name: "assignee_id", type: "integer", required: false, descKey: "docs.api.tickets.details.updateTicket.fields.assignee_id" },
          { name: "labels", type: "string[]", required: false, descKey: "docs.api.tickets.details.updateTicket.fields.labels" },
          { name: "parent_slug", type: "string", required: false, descKey: "docs.api.tickets.details.updateTicket.fields.parent_slug" },
        ],
      },
    ],
  },
  {
    method: "PATCH",
    path: "/tickets/:slug/status",
    descKey: "docs.api.tickets.details.updateStatus.description",
    response: STATUS_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "slug", type: "string", required: true, descKey: "docs.api.tickets.details.updateStatus.params.slug" },
        ],
      },
      {
        kind: "body",
        rows: [
          { name: "status", type: "string", required: true, descKey: "docs.api.tickets.details.updateStatus.fields.status" },
        ],
      },
    ],
  },
  {
    method: "DELETE",
    path: "/tickets/:slug",
    descKey: "docs.api.tickets.details.deleteTicket.description",
    response: DELETE_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "slug", type: "string", required: true, descKey: "docs.api.tickets.details.deleteTicket.params.slug" },
        ],
      },
    ],
  },
];
