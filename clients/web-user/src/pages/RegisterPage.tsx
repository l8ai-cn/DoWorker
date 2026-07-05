import { useEffect, useState, type FormEvent } from "react";
import { useSearchParams } from "@/lib/routing";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AuthFormAlert } from "@/components/auth/AuthFormAlert";
import { AuthPageShell } from "@/components/auth/AuthPageShell";
import { register as registerRequest } from "@/lib/accountsApi";

const MIN_PASSWORD_LENGTH = 8;

export function RegisterPage() {
  const [params] = useSearchParams();
  const invite = params.get("invite") ?? "";

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const missingInvite = invite === "";

  useEffect(() => {
    const el = document.getElementById("register-username");
    if (el instanceof HTMLInputElement) el.focus();
  }, []);

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (submitting) return;
    setError(null);

    if (password !== confirm) {
      setError("Passwords don't match.");
      return;
    }
    if (password.length < MIN_PASSWORD_LENGTH) {
      setError(`Password must be at least ${MIN_PASSWORD_LENGTH} characters.`);
      return;
    }

    setSubmitting(true);
    const result = await registerRequest({ invite, username, password });
    if (result.ok) {
      window.location.href = "/";
      return;
    }
    setSubmitting(false);
    setError(result.error);
  }

  return (
    <AuthPageShell
      title="Create your account"
      description="You were invited to join this Do Worker server."
    >
      {missingInvite ? (
        <AuthFormAlert message="This page needs an invite token in the URL — open the link your admin sent you." />
      ) : (
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <label htmlFor="register-username" className="text-sm font-medium leading-none">
              Username
            </label>
            <Input
              id="register-username"
              type="text"
              autoComplete="username"
              value={username}
              onChange={(e) => setUsername(e.target.value.toLowerCase())}
              disabled={submitting}
              required
              pattern="[a-z0-9][a-z0-9._\-]{0,63}(@[a-z0-9.\-]+\.[a-z]{2,})?"
              title="Lowercase letters, digits, dots, hyphens, underscores (or a lowercase email)"
            />
            <p className="text-xs text-muted-foreground">
              Lowercase letters, digits, dots, hyphens, underscores — or a lowercase email.
            </p>
          </div>

          <div className="space-y-1.5">
            <label htmlFor="register-password" className="text-sm font-medium leading-none">
              Password
            </label>
            <Input
              id="register-password"
              type="password"
              autoComplete="new-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={submitting}
              required
              minLength={MIN_PASSWORD_LENGTH}
            />
          </div>

          <div className="space-y-1.5">
            <label htmlFor="register-confirm" className="text-sm font-medium leading-none">
              Confirm password
            </label>
            <Input
              id="register-confirm"
              type="password"
              autoComplete="new-password"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              disabled={submitting}
              required
              minLength={MIN_PASSWORD_LENGTH}
            />
          </div>

          {error !== null ? <AuthFormAlert message={error} /> : null}

          <Button
            type="submit"
            className="w-full"
            disabled={submitting || password.length < MIN_PASSWORD_LENGTH || username.length === 0}
          >
            {submitting ? "Creating…" : "Create account"}
          </Button>
        </form>
      )}
    </AuthPageShell>
  );
}
