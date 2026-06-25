export const RUNNERS_LIST_JSON = `{
  "runners": [
    {
      "id": 1,
      "name": "dev-runner-01",
      "status": "online",
      "version": "1.2.0",
      "os": "linux",
      "arch": "amd64",
      "labels": ["gpu", "high-memory"],
      "pod_count": 3,
      "max_pods": 10,
      "last_heartbeat_at": "2025-01-15T14:30:00Z",
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}`;

export const RUNNER_JSON = `{
  "runner": {
    "id": 1,
    "name": "dev-runner-01",
    "status": "online",
    "version": "1.2.0",
    "os": "linux",
    "arch": "amd64",
    "labels": ["gpu", "high-memory"],
    "pod_count": 3,
    "max_pods": 10,
    "last_heartbeat_at": "2025-01-15T14:30:00Z",
    "created_at": "2025-01-01T00:00:00Z"
  }
}`;

export const RUNNER_PODS_JSON = `{
  "pods": [
    {
      "key": "pod-abc123",
      "status": "running",
      "agent_slug": "claude-code",
      "prompt": "Fix the login bug",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "total": 8,
  "limit": 50,
  "offset": 0
}`;
