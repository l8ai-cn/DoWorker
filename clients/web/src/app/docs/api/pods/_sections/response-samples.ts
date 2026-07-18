export const PODS_LIST_JSON = `{
  "pods": [
    {
      "key": "pod-abc123",
      "status": "running",
      "agent_slug": "claude-code",
      "runner_id": "550e8400-e29b-41d4-a716-446655440000",
      "prompt": "Fix the login bug",
      "repository_id": 1,
      "branch": "main",
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T10:35:00Z"
    }
  ],
  "total": 42,
  "limit": 20,
  "offset": 0
}`;

export const POD_JSON = `{
  "pod": {
    "key": "pod-abc123",
    "status": "running",
    "agent_slug": "claude-code",
    "runner_id": "550e8400-e29b-41d4-a716-446655440000",
    "prompt": "Fix the login bug",
    "repository_id": 1,
    "branch": "main",
    "ticket_slug": "AM-42",
    "channel_id": 5,
    "sandbox_type": "worktree",
    "auto_close": false,
    "pod_timeout_minutes": 60,
    "max_turns": 100,
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:35:00Z"
  }
}`;

export const POD_RESUME_JSON = `{
  "pod": {
    "pod_key": "pod-xyz789",
    "status": "initializing",
    "agent_slug": "claude-code",
    "source_pod_key": "pod-abc123",
    "created_at": "2025-01-15T10:30:00Z"
  }
}`;

export const TERMINATE_JSON = `{
  "message": "Pod terminated"
}`;
