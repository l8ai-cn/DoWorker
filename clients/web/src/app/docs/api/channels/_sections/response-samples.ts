export const CHANNEL_JSON = `{
  "channel": {
    "id": 1,
    "name": "feature-auth",
    "description": "Authentication implementation channel",
    "repository_id": 1,
    "ticket_slug": "AM-42",
    "document": "## Context\\nImplement JWT auth...",
    "archived": false,
    "created_at": "2025-01-10T08:00:00Z",
    "updated_at": "2025-01-15T14:20:00Z"
  }
}`;

export const CHANNELS_LIST_JSON = `{
  "channels": [
    {
      "id": 1,
      "name": "feature-auth",
      "description": "Authentication implementation channel",
      "repository_id": 1,
      "ticket_slug": "AM-42",
      "archived": false,
      "created_at": "2025-01-10T08:00:00Z",
      "updated_at": "2025-01-15T14:20:00Z"
    }
  ],
  "total": 12
}`;

export const MESSAGES_JSON = `{
  "messages": [
    {
      "id": 100,
      "channel_id": 1,
      "content": "I've completed the JWT implementation.",
      "sender_type": "agent",
      "pod_key": "pod-abc123",
      "created_at": "2025-01-15T14:20:00Z"
    }
  ]
}`;

export const SEND_MESSAGE_JSON = `{
  "message": {
    "id": 101,
    "channel_id": 1,
    "content": "Please review the auth module.",
    "sender_type": "api",
    "pod_key": null,
    "created_at": "2025-01-15T15:00:00Z"
  }
}`;
