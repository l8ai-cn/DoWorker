import { readDoWorkerJWT } from "./do-worker/auth-session";
import type { ServerInfo } from "./do-worker/server-info";

/** True when the deploy exposes sign-in or the client holds a session token. */
export function hasAuthSession(info: ServerInfo | "loading"): boolean {
  if (info === "loading") return false;
  if (info.login_url !== null || info.accounts_enabled) return true;
  return readDoWorkerJWT() !== null;
}
