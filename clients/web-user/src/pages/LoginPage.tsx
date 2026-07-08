import { useEffect, useState, type FormEvent } from "react";
import { useSearchParams } from "@/lib/routing";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AuthFormAlert } from "@/components/auth/AuthFormAlert";
import { useI18n } from "@/i18n/I18nProvider";
import { AuthPageShell } from "@/components/auth/AuthPageShell";
import { DevCredentialsPanel } from "@/components/auth/DevCredentialsPanel";
import { getMe, login as loginRequest } from "@/lib/accountsApi";
import { hostFetch } from "@/lib/do-worker";
import { persistDoWorkerSession } from "@/lib/do-worker-auth";
import { sanitizeReturnTo } from "@/lib/auth-return-to";

const LAST_USERNAME_KEY = "do-worker.lastLoginUsername";

function readLastUsername(): string {
  try {
    return window.localStorage.getItem(LAST_USERNAME_KEY) ?? "";
  } catch {
    return "";
  }
}

function rememberUsername(value: string): void {
  try {
    window.localStorage.setItem(LAST_USERNAME_KEY, value);
  } catch {
    // Best-effort — see readLastUsername.
  }
}

export function LoginPage() {
  const { t } = useI18n();
  const [params] = useSearchParams();
  const returnTo = sanitizeReturnTo(params.get("return_to"));
  const magicError = params.get("magic");

  const [username, setUsername] = useState(readLastUsername);
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(
    magicError === "expired"
      ? "That sign-in link has expired. Enter your password to sign in."
      : magicError === "missing"
        ? "That sign-in link is no longer valid. Enter your password to sign in."
        : null,
  );

  useEffect(() => {
    void (async () => {
      const account = await getMe();
      if (account !== null) {
        window.location.href = returnTo;
      }
    })();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const targetId = username ? "login-password" : "login-username";
    const el = document.getElementById(targetId);
    if (el instanceof HTMLInputElement) {
      el.focus();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (submitting) return;
    setSubmitting(true);
    setError(null);

    const result = await loginRequest({
      username: username.trim(),
      password,
    });
    if (result.ok) {
      rememberUsername(username.trim());
      let orgSlug = typeof result.org_slug === "string" ? result.org_slug : null;
      if (!orgSlug) {
        try {
          const meRes = await hostFetch("/v1/me", {
            headers: { Authorization: `Bearer ${result.token}` },
            cache: "no-store",
          });
          if (meRes.ok) {
            const me = (await meRes.json()) as { org_slug?: string };
            if (typeof me.org_slug === "string" && me.org_slug) {
              orgSlug = me.org_slug;
            }
          }
        } catch {
          // Best-effort — resolveIdentity backfills on next load.
        }
      }
      persistDoWorkerSession({
        accessToken: result.token,
        expiresIn: result.expires_in,
        orgSlug,
      });
      window.location.href = returnTo;
      return;
    }
    setSubmitting(false);
    setError(result.error);
  }

  return (
    <AuthPageShell
      title={t.composer.signIn}
      description={t.auth.welcome}
      footer={
        <DevCredentialsPanel
          disabled={submitting}
          onPick={(user, pass) => {
            setUsername(user);
            setPassword(pass);
            setError(null);
          }}
        />
      }
    >
      <form onSubmit={onSubmit} className="space-y-4">
        <div className="space-y-1.5">
          <label htmlFor="login-username" className="text-sm font-medium leading-none">
            {t.composer.username}
          </label>
          <Input
            id="login-username"
            type="text"
            autoComplete="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            disabled={submitting}
            required
          />
          <p className="text-xs text-muted-foreground">
            Email or username (dev: <code className="font-mono">devuser</code> or{" "}
            <code className="font-mono">admin</code>).
          </p>
        </div>

        <div className="space-y-1.5">
          <label htmlFor="login-password" className="text-sm font-medium leading-none">
            {t.composer.password}
          </label>
          <Input
            id="login-password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={submitting}
            required
          />
        </div>

        {error !== null ? <AuthFormAlert message={error} /> : null}

        <Button type="submit" className="w-full" disabled={submitting || password.length === 0}>
          {submitting ? t.composer.signingIn : t.composer.signIn}
        </Button>
      </form>
    </AuthPageShell>
  );
}
