import { logout } from "@/lib/accountsApi";
import {
  clearAgentCloudSession,
  markSessionLoggedOut,
} from "@/lib/agent-cloud/auth-session";
import { resetIdentity } from "@/lib/identity";

export async function signOutSession(accountsEnabled: boolean): Promise<void> {
  markSessionLoggedOut();
  if (accountsEnabled) {
    await logout();
  } else {
    clearAgentCloudSession();
  }
  resetIdentity();
  window.location.href = accountsEnabled ? "/login" : "/auth/logout";
}
