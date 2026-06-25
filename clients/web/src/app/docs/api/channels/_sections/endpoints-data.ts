import type { EndpointRow } from "../../_components/endpoint-summary-table";
import type { DetailSpec } from "../../_components/endpoint-detail";
import {
  CHANNEL_JSON,
  CHANNELS_LIST_JSON,
  MESSAGES_JSON,
  SEND_MESSAGE_JSON,
} from "./response-samples";

export const summaryRows: EndpointRow[] = [
  {
    method: "GET",
    path: "/channels",
    scope: "channels:read",
    descKey: "docs.api.channels.endpoints.list",
  },
  {
    method: "GET",
    path: "/channels/:id",
    scope: "channels:read",
    descKey: "docs.api.channels.endpoints.get",
  },
  {
    method: "GET",
    path: "/channels/:id/messages",
    scope: "channels:read",
    descKey: "docs.api.channels.endpoints.messages",
  },
  {
    method: "POST",
    path: "/channels",
    scope: "channels:write",
    descKey: "docs.api.channels.endpoints.create",
  },
  {
    method: "PUT",
    path: "/channels/:id",
    scope: "channels:write",
    descKey: "docs.api.channels.endpoints.update",
  },
  {
    method: "POST",
    path: "/channels/:id/messages",
    scope: "channels:write",
    descKey: "docs.api.channels.endpoints.sendMessage",
  },
];

export const detailEndpoints: DetailSpec[] = [
  {
    method: "GET",
    path: "/channels",
    descKey: "docs.api.channels.details.listChannels.description",
    response: CHANNELS_LIST_JSON,
    tables: [
      {
        kind: "query",
        rows: [
          { name: "repository_id", type: "integer", required: false, descKey: "docs.api.channels.details.listChannels.params.repository_id" },
          { name: "ticket_slug", type: "string", required: false, descKey: "docs.api.channels.details.listChannels.params.ticket_slug" },
          { name: "include_archived", type: "boolean", required: false, descKey: "docs.api.channels.details.listChannels.params.include_archived" },
        ],
      },
    ],
  },
  {
    method: "GET",
    path: "/channels/:id",
    descKey: "docs.api.channels.details.getChannel.description",
    response: CHANNEL_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "id", type: "integer", required: true, descKey: "docs.api.channels.details.getChannel.params.id" },
        ],
      },
    ],
  },
  {
    method: "GET",
    path: "/channels/:id/messages",
    descKey: "docs.api.channels.details.getMessages.description",
    response: MESSAGES_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "id", type: "integer", required: true, descKey: "docs.api.channels.details.getMessages.params.id" },
        ],
      },
      {
        kind: "query",
        withDefault: true,
        rows: [
          { name: "limit", type: "integer", required: false, default: "50", descKey: "docs.api.channels.details.getMessages.params.limit" },
        ],
      },
    ],
  },
  {
    method: "POST",
    path: "/channels",
    descKey: "docs.api.channels.details.createChannel.description",
    response: CHANNEL_JSON,
    tables: [
      {
        kind: "body",
        rows: [
          { name: "name", type: "string", required: true, descKey: "docs.api.channels.details.createChannel.fields.name" },
          { name: "description", type: "string", required: false, descKey: "docs.api.channels.details.createChannel.fields.description" },
          { name: "repository_id", type: "integer", required: false, descKey: "docs.api.channels.details.createChannel.fields.repository_id" },
          { name: "ticket_slug", type: "string", required: false, descKey: "docs.api.channels.details.createChannel.fields.ticket_slug" },
          { name: "document", type: "string", required: false, descKey: "docs.api.channels.details.createChannel.fields.document" },
        ],
      },
    ],
  },
  {
    method: "PUT",
    path: "/channels/:id",
    descKey: "docs.api.channels.details.updateChannel.description",
    response: CHANNEL_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "id", type: "integer", required: true, descKey: "docs.api.channels.details.updateChannel.params.id" },
        ],
      },
      {
        kind: "body",
        rows: [
          { name: "name", type: "string", required: false, descKey: "docs.api.channels.details.updateChannel.fields.name" },
          { name: "description", type: "string", required: false, descKey: "docs.api.channels.details.updateChannel.fields.description" },
          { name: "document", type: "string", required: false, descKey: "docs.api.channels.details.updateChannel.fields.document" },
        ],
      },
    ],
  },
  {
    method: "POST",
    path: "/channels/:id/messages",
    descKey: "docs.api.channels.details.sendMessage.description",
    response: SEND_MESSAGE_JSON,
    tables: [
      {
        kind: "path",
        rows: [
          { name: "id", type: "integer", required: true, descKey: "docs.api.channels.details.sendMessage.params.id" },
        ],
      },
      {
        kind: "body",
        rows: [
          { name: "content", type: "string", required: true, descKey: "docs.api.channels.details.sendMessage.fields.content" },
          { name: "pod_key", type: "string", required: false, descKey: "docs.api.channels.details.sendMessage.fields.pod_key" },
        ],
      },
    ],
  },
];
