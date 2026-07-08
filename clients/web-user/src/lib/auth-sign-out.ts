import { logout } from "@/lib/accountsApi";
import {
  clearDoWorkerSession,
  markSessionLoggedOut,
} from "@/lib/do-worker/auth-session";
import { resetIdentity } from "@/lib/identity";

export async function signOutSession(accountsEnabled: boolean): Promise<void> {
  markSessionLoggedOut();
  if (accountsEnabled) {
    await logout();
  } else {
    clearDoWorkerSession();
  }
  resetIdentity();
  window.location.href = accountsEnabled ? "/login" : "/auth/logout";
}
