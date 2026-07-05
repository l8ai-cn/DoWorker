import { useI18n } from "@/i18n/I18nProvider";
import { Button } from "@/components/ui/button";
import { isCurrentServerLocal } from "@/lib/serverOrigin";

interface DevCredentialsPanelProps {
  disabled?: boolean;
  onPick: (username: string, password: string) => void;
}

const DEV_ACCOUNTS = [
  { username: "devuser", password: "AdminAb123456", label: "Dev user" },
  { username: "admin", password: "Ab123456", label: "Admin" },
] as const;

export function DevCredentialsPanel({ disabled, onPick }: DevCredentialsPanelProps) {
  const { t } = useI18n();
  if (!isCurrentServerLocal()) return null;

  const accounts = [
    { ...DEV_ACCOUNTS[0], label: t.auth.devUser },
    { ...DEV_ACCOUNTS[1], label: t.auth.admin },
  ] as const;

  return (
    <div className="space-y-2">
      <p className="text-center text-xs text-muted-foreground">{t.auth.devAccounts}</p>
      <div className="space-y-2 rounded-lg border border-border/60 bg-muted/40 px-3 py-2.5 text-xs">
        {accounts.map((account) => (
          <div key={account.username} className="flex items-center justify-between gap-2">
            <span className="text-muted-foreground">
              <span className="font-medium text-foreground">{account.label}</span>{" "}
              <code className="font-mono">{account.username}</code>
              <span className="mx-1 text-border-strong">/</span>
              <code className="font-mono">{account.password}</code>
            </span>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="h-7 shrink-0 px-2.5 text-xs"
              disabled={disabled}
              onClick={() => onPick(account.username, account.password)}
            >
              {t.auth.use}
            </Button>
          </div>
        ))}
      </div>
    </div>
  );
}
