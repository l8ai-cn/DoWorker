export const REPOS_LIST_JSON = `{
  "repositories": [
    {
      "id": 1,
      "name": "agentsmesh",
      "full_name": "org/agentsmesh",
      "provider": "gitlab",
      "url": "https://gitlab.com/org/agentsmesh",
      "default_branch": "main",
      "created_at": "2025-01-01T00:00:00Z"
    }
  ]
}`;

export const REPO_JSON = `{
  "repository": {
    "id": 1,
    "name": "agentsmesh",
    "full_name": "org/agentsmesh",
    "provider": "gitlab",
    "url": "https://gitlab.com/org/agentsmesh",
    "default_branch": "main",
    "created_at": "2025-01-01T00:00:00Z"
  }
}`;

export const BRANCHES_JSON = `{
  "branches": [
    "main",
    "develop",
    "feature/auth",
    "fix/login-bug"
  ]
}`;

export const MERGE_REQUESTS_JSON = `{
  "merge_requests": [
    {
      "id": 101,
      "title": "Add JWT authentication",
      "state": "opened",
      "source_branch": "feature/auth",
      "target_branch": "main",
      "author": "john.doe",
      "url": "https://gitlab.com/org/repo/-/merge_requests/101",
      "created_at": "2025-01-14T09:00:00Z"
    }
  ]
}`;
