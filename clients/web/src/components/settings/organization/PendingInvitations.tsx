import { Button } from "@/components/ui/button";
import type { Invitation } from "@/lib/api/facade/invitationConnect";
import type { TranslationFn } from "./GeneralSettings";

interface PendingInvitationsProps {
  invitations: Invitation[];
  loading: boolean;
  resendingId: number | null;
  t: TranslationFn;
  onResend: (invitationId: number) => void;
  onRevoke: (invitationId: number) => void;
}

export function PendingInvitations({
  invitations, loading, resendingId, t, onResend, onRevoke,
}: PendingInvitationsProps) {
  if (loading || invitations.length === 0) return null;

  return (
    <div className="mt-6">
      <h3 className="text-sm font-semibold text-muted-foreground mb-3">
        {t("settings.members.pendingInvitations")}
      </h3>
      <div className="space-y-3">
        {invitations.map((invitation) => {
          const id = Number(invitation.id);
          return (
          <div key={id}
            className="flex items-center justify-between p-4 surface-card border-dashed">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-muted/50 flex items-center justify-center text-sm font-medium text-muted-foreground">
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <rect width="20" height="16" x="2" y="4" rx="2" />
                  <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" />
                </svg>
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <span className="font-medium">{invitation.email}</span>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${getRoleBadgeColor(invitation.role)}`}>
                    {invitation.role}
                  </span>
                </div>
                <p className="text-xs text-muted-foreground">
                  {t("settings.members.pendingExpires", { date: formatExpiryDate(invitation.expiresAt) })}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="ghost" size="sm" onClick={() => onResend(id)}
                disabled={resendingId === id}>
                {resendingId === id
                  ? t("settings.members.resending")
                  : t("settings.members.resend")}
              </Button>
              <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive"
                onClick={() => onRevoke(id)}>
                {t("settings.members.revoke")}
              </Button>
            </div>
          </div>
        );
        })}
      </div>
    </div>
  );
}

function getRoleBadgeColor(role: string) {
  switch (role) {
    case "owner": return "bg-accent text-primary";
    case "admin": return "bg-info-bg text-info";
    default: return "bg-muted text-muted-foreground";
  }
}

function formatExpiryDate(dateStr: string) {
  const date = new Date(dateStr);
  return date.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}
