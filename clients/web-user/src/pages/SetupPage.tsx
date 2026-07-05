import { useEffect, useState, type FormEvent } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AuthFormAlert } from "@/components/auth/AuthFormAlert";
import { AuthPageShell } from "@/components/auth/AuthPageShell";
import { setup as setupRequest } from "@/lib/accountsApi";

const MIN_PASSWORD_LENGTH = 8;

export function SetupPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const el = document.getElementById("setup-username");
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
    const result = await setupRequest({ username, password });
    if (result.ok) {
      window.location.href = "/";
      return;
    }
    setSubmitting(false);
    if (result.status === 409) {
      window.location.href = "/login";
      return;
    }
    setError(result.error);
  }

  return (
    <AuthPageShell
      title="Create the admin account"
      description="First run — pick the username and password for this server's admin."
    >
      <form onSubmit={onSubmit} className="space-y-4">
        <div className="space-y-1.5">
          <label htmlFor="setup-username" className="text-sm font-medium leading-none">
            Username
          </label>
          <Input
            id="setup-username"
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
          <label htmlFor="setup-password" className="text-sm font-medium leading-none">
            Password
          </label>
          <Input
            id="setup-password"
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
          <label htmlFor="setup-confirm" className="text-sm font-medium leading-none">
            Confirm password
          </label>
          <Input
            id="setup-confirm"
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
          {submitting ? "Creating…" : "Create admin"}
        </Button>
      </form>
    </AuthPageShell>
  );
}
