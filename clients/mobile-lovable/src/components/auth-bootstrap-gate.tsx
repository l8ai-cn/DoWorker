import { RefreshCw } from "lucide-react";
import { useEffect, useState, type ReactNode } from "react";
import { restoreAuthIdentity } from "@/lib/auth-store";

type BootstrapState = "loading" | "ready" | "error";

export function AuthBootstrapGate({ children }: { children: ReactNode }) {
  const [state, setState] = useState<BootstrapState>("loading");
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let active = true;
    void restoreAuthIdentity().then(
      () => active && setState("ready"),
      () => active && setState("error"),
    );
    return () => {
      active = false;
    };
  }, [attempt]);

  if (state === "ready") {
    return <>{children}</>;
  }

  if (state === "error") {
    return (
      <main className="flex min-h-screen flex-col items-center justify-center gap-4 px-6 text-center">
        <p className="text-sm text-muted-foreground">无法恢复登录状态。</p>
        <button
          type="button"
          aria-label="重试恢复登录状态"
          title="重试"
          onClick={() => {
            setState("loading");
            setAttempt((value) => value + 1);
          }}
          className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-input hover:bg-accent"
        >
          <RefreshCw className="h-4 w-4" />
        </button>
      </main>
    );
  }

  return <main className="min-h-screen bg-background" aria-busy="true" />;
}
