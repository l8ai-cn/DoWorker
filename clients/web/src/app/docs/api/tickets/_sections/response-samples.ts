export const TICKET_JSON = `{
  "ticket": {
    "id": 1,
    "slug": "AM-42",
    "type": "feature",
    "title": "Implement user authentication",
    "status": "in_progress",
    "priority": "high",
    "assignee_id": 10,
    "repository_id": 1,
    "labels": ["backend", "auth"],
    "parent_slug": null,
    "created_at": "2025-01-10T08:00:00Z",
    "updated_at": "2025-01-15T14:20:00Z"
  }
}`;

export const LIST_JSON = `{
  "tickets": [
    {
      "id": 1,
      "slug": "AM-42",
      "type": "feature",
      "title": "Implement user authentication",
      "status": "in_progress",
      "priority": "high",
      "assignee_id": 10,
      "repository_id": 1,
      "labels": ["backend", "auth"],
      "created_at": "2025-01-10T08:00:00Z",
      "updated_at": "2025-01-15T14:20:00Z"
    }
  ],
  "total": 156,
  "limit": 20,
  "offset": 0
}`;

export const BOARD_JSON = `{
  "board": {
    "columns": [
      {
        "status": "open",
        "tickets": [
          {
            "id": 2,
            "slug": "AM-43",
            "title": "Fix CSS layout issue",
            "type": "bug",
            "priority": "medium",
            "assignee_id": 5
          }
        ]
      },
      {
        "status": "in_progress",
        "tickets": []
      }
    ]
  }
}`;

export const STATUS_JSON = `{
  "message": "Status updated"
}`;

export const DELETE_JSON = `{
  "message": "Ticket deleted"
}`;
