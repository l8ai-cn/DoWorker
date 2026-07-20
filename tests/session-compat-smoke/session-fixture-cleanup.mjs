const API = process.env.SESSION_COMPAT_API_URL || "http://localhost:10015";
const ORG = "dev-org";

export function trackSmokeSession(sessionIDs, session) {
  if (!(sessionIDs instanceof Set)) {
    throw new Error("sessionIDs must be a Set");
  }
  const id = typeof session === "string" ? session : session?.id;
  if (typeof id !== "string" || id.trim() === "") {
    throw new Error("smoke fixture session must have an id");
  }
  sessionIDs.add(id);
  return session;
}

export function untrackSmokeSession(sessionIDs, sessionID) {
  if (!(sessionIDs instanceof Set)) {
    throw new Error("sessionIDs must be a Set");
  }
  sessionIDs.delete(sessionID);
}

export async function cleanupSmokeSessions(token, sessionIDs) {
  if (typeof token !== "string" || token.trim() === "") {
    throw new Error("cleanup requires an authenticated token");
  }
  if (!(sessionIDs instanceof Set)) {
    throw new Error("sessionIDs must be a Set");
  }

  const failures = [];
  let deleted = 0;
  for (const sessionID of [...sessionIDs].reverse()) {
    try {
      const response = await fetch(
        `${API}/v1/sessions/${encodeURIComponent(sessionID)}?delete_branch=false`,
        {
          method: "DELETE",
          headers: {
            Authorization: `Bearer ${token}`,
            "X-Organization-Slug": ORG,
          },
        },
      );
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      sessionIDs.delete(sessionID);
      deleted += 1;
    } catch (error) {
      failures.push(`${sessionID}: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  if (failures.length > 0) {
    throw new Error(`smoke fixture cleanup failed: ${failures.join("; ")}`);
  }
  return deleted;
}
