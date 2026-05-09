/* eslint-disable react-hooks/rules-of-hooks -- Playwright fixtures use function
   arguments named `use` which happen to trigger the React Hooks rule. */
import { test as base, expect, BrowserContext, Page } from "@playwright/test";
import { getApiBaseUrl, getWebBaseUrl, TEST_USER, TEST_ORG_SLUG } from "../helpers/env";

// E2E fixtures for Block Store specs. Provides:
//   - `token` / `workspaceID`: authenticated session + default workspace id
//   - `authenticatedPage`: a page that's already logged in (JWT in localStorage)
//   - `api`: tiny fetch wrapper with Authorization header pre-applied
//
// Tests should prefer driving state via `api` (deterministic) and asserting
// via the UI (via `authenticatedPage`). Direct-from-UI setup is slower and
// flakier when the backend is also under test.
//
// Distinct from `api.fixture.ts` (used by API-level specs that already share
// the suite-wide storageState login): blockstore specs need an isolated
// per-test workspace + JWT-in-localStorage seeding for the Rust auth manager
// + an in-page proxy so /api/* and ws:/api/* hit the backend host directly
// (Bazel sandboxes the Next.js dev server so its API_PROXY_TARGET is dropped).

const API_BASE = getApiBaseUrl();
const ORG_SLUG = TEST_ORG_SLUG;
const DEV_EMAIL = TEST_USER.email;
const DEV_PASSWORD = TEST_USER.password;

interface BlockstoreFixtures {
  token: string;
  workspaceID: string;
  /**
   * A freshly-provisioned workspace for this single test. Root block created
   * server-side. Isolates the test from accumulated writes in the shared
   * default workspace, so assertions that depend on "state count" (semantic
   * search ranking, slash-menu contents, etc.) stay stable run-over-run.
   * Cleaned up automatically after the test finishes (best-effort DELETE on
   * the workspace id) so the dev DB doesn't accumulate orphan rows across
   * repeated runs. Failures during teardown are swallowed — we never want a
   * cleanup error to mask the test's real assertion failure.
   */
  isolatedWorkspace: { id: string; rootID: string };
  api: ApiClient;
  authenticatedPage: Page;
}

interface ApiClient {
  post<T = unknown>(path: string, body: unknown): Promise<T>;
  get<T = unknown>(path: string): Promise<T>;
}

async function login(): Promise<string> {
  const res = await fetch(`${API_BASE}/api/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email: DEV_EMAIL, password: DEV_PASSWORD }),
  });
  if (!res.ok) throw new Error(`login failed: ${res.status}`);
  const json = (await res.json()) as { token: string };
  return json.token;
}

// Shared token cache — fixtures are instantiated once per test, and dev's
// 20/min /auth/login rate limit would 429 a 20+ spec run otherwise. One
// token lives for the whole Playwright process; it's refreshed on token
// expiry (reject with 401 retry path), which the REST wrapper handles.
let cachedToken: Promise<string> | null = null;
function sharedLogin(): Promise<string> {
  if (!cachedToken) cachedToken = login();
  return cachedToken;
}

// Shared workspace id cache — ensureWorkspace is idempotent server-side but
// cheap to skip entirely.
let cachedWorkspaceID: Promise<string> | null = null;
function sharedEnsureWorkspace(token: string): Promise<string> {
  if (!cachedWorkspaceID) cachedWorkspaceID = ensureWorkspace(token);
  return cachedWorkspaceID;
}

async function ensureWorkspace(token: string): Promise<string> {
  const res = await fetch(
    `${API_BASE}/api/v1/orgs/${ORG_SLUG}/blocks/workspaces/default`,
    {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "X-Organization-Slug": ORG_SLUG,
      },
    },
  );
  if (!res.ok) throw new Error(`ensureWorkspace failed: ${res.status}`);
  const json = (await res.json()) as { id: string };
  return json.id;
}

// provisionIsolatedWorkspace creates a brand-new workspace each call, which
// a test uses instead of the shared default when it needs a clean slate.
// Slug randomisation is aggressive because Playwright's workers + reruns can
// share the same millisecond; collisions surface as 409 which is explicit.
async function provisionIsolatedWorkspace(
  token: string,
): Promise<{ id: string; rootID: string }> {
  const slug = `e2e-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  const res = await fetch(`${API_BASE}/api/v1/orgs/${ORG_SLUG}/blocks/workspaces`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Organization-Slug": ORG_SLUG,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ slug, name: slug }),
  });
  if (!res.ok) throw new Error(`provisionIsolatedWorkspace failed: ${res.status}: ${await res.text()}`);
  const json = (await res.json()) as { id: string; root_block_id: string };
  return { id: json.id, rootID: json.root_block_id };
}

// tearDownIsolatedWorkspace deletes the workspace + every row it owns via
// the backend's DELETE /blocks/workspaces/:ws_id. Intentionally best-effort:
// a teardown error must never convert a green test into a red one, so we
// log (via console.warn) and swallow. The worst case is one leaked row;
// periodic resets still catch it.
async function tearDownIsolatedWorkspace(token: string, wsID: string): Promise<void> {
  try {
    await fetch(`${API_BASE}/api/v1/orgs/${ORG_SLUG}/blocks/workspaces/${wsID}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${token}`,
        "X-Organization-Slug": ORG_SLUG,
      },
    });
  } catch (err) {
    // eslint-disable-next-line no-console
    console.warn(`[e2e] workspace teardown failed for ${wsID}:`, err);
  }
}

function makeApi(token: string): ApiClient {
  const headers = {
    Authorization: `Bearer ${token}`,
    "X-Organization-Slug": ORG_SLUG,
    "Content-Type": "application/json",
  };
  return {
    async post<T>(path: string, body: unknown): Promise<T> {
      const res = await fetch(`${API_BASE}${path}`, {
        method: "POST",
        headers,
        body: JSON.stringify(body ?? {}),
      });
      if (!res.ok) {
        throw new Error(`${path} → ${res.status}: ${await res.text()}`);
      }
      return res.json() as Promise<T>;
    },
    async get<T>(path: string): Promise<T> {
      const res = await fetch(`${API_BASE}${path}`, { headers });
      if (!res.ok) {
        throw new Error(`${path} → ${res.status}: ${await res.text()}`);
      }
      return res.json() as Promise<T>;
    },
  };
}

// installApiProxy intercepts /api/* requests at the Playwright layer and
// re-issues them server-side against the real backend. Bazel sandboxes the
// Next.js dev server so .env.local with API_PROXY_TARGET isn't picked up;
// without this, /api/* hits Next.js's 404 page. We can't use
// route.continue({url:…}) for cross-origin redirect — same-origin requests
// blocked by the browser as ERR_BLOCKED_BY_CLIENT.
async function installApiProxy(target: BrowserContext | Page): Promise<void> {
  // Match /api/* but EXCLUDE the realtime WS endpoint. WS upgrades can't
  // be handled by route.fulfill (no body), and Playwright's route.continue
  // for WS is unreliable — letting the request pass through unintercepted
  // is the only way to keep the in-page WebSocket constructor patch (below)
  // in charge of redirecting the upgrade to the backend host.
  await target.route(/\/api\/(?!.*\/ws\/).+/, async (route) => {
    const orig = route.request();
    const url = new URL(orig.url());
    const upstream = `${API_BASE}${url.pathname}${url.search}`;
    const headers = { ...orig.headers() };
    delete headers["host"];
    delete headers["origin"];
    delete headers["referer"];
    try {
      const res = await fetch(upstream, {
        method: orig.method(),
        headers,
        body: ["GET", "HEAD"].includes(orig.method())
          ? undefined
          : orig.postData() ?? undefined,
      });
      const respHeaders: Record<string, string> = {};
      res.headers.forEach((v, k) => {
        const lk = k.toLowerCase();
        if (lk === "content-encoding" || lk === "transfer-encoding") return;
        respHeaders[k] = v;
      });
      const buf = Buffer.from(await res.arrayBuffer());
      await route.fulfill({ status: res.status, headers: respHeaders, body: buf });
    } catch (err) {
      // eslint-disable-next-line no-console
      console.warn(`[e2e] API proxy failed for ${upstream}:`, err);
      await route.abort();
    }
  });
  // Realtime path: Next.js dev server can't proxy WS through rewrites, so
  // the page tries ws://localhost:WEB_PORT/api/v1/.../ws/events and dies.
  // Override the WebSocket constructor in-page to point any /api/* WS URLs
  // straight at the backend host. We use a Proxy on the native class so
  // callers (web_sys + JS) get a real WebSocket instance and prototype chain.
  const wsHost = API_BASE.replace(/^https?:\/\//, "").replace(/\/$/, "");
  const wsScheme = API_BASE.startsWith("https") ? "wss" : "ws";
  await target.addInitScript(
    ({ host, scheme }) => {
      const Orig = window.WebSocket;
      const Patched = new Proxy(Orig, {
        construct(t, args) {
          const [url, protocols] = args as [unknown, unknown];
          let s = typeof url === "string" ? url : String(url);
          try {
            const u = new URL(s, window.location.href);
            if (u.pathname.startsWith("/api/")) {
              u.host = host;
              u.protocol = scheme + ":";
              s = u.toString();
            }
          } catch {
            /* fall through */
          }
          return protocols !== undefined
            ? Reflect.construct(t, [s, protocols])
            : Reflect.construct(t, [s]);
        },
      });
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (window as any).WebSocket = Patched;
    },
    { host: wsHost, scheme: wsScheme },
  );
}

// Inject the JWT into localStorage before Next.js boots so the Rust auth
// manager (clients/core/crates/auth/src/state.rs) restores from storage and
// hydrates as logged-in on first render. Storage is namespaced per
// `base_url`: `agentsmesh-auth/<url_slug>/session`. The slug derivation
// (`http_localhost_25357` from `http://localhost:25357`) mirrors Rust's
// `url_slug()` in state.rs — non-alphanumerics → `_`, max 64 chars.
function urlSlug(url: string): string {
  const u = new URL(url);
  const port = u.port ? `_${u.port}` : "";
  const raw = `${u.protocol.replace(":", "")}_${u.hostname.toLowerCase()}${port}`;
  return raw.replace(/[^a-zA-Z0-9]/g, "_").slice(0, 64);
}

async function seedAuth(target: BrowserContext | Page, token: string): Promise<void> {
  const me = await fetch(`${API_BASE}/api/v1/users/me`, {
    headers: { Authorization: `Bearer ${token}` },
  }).then((r) => r.json());
  const orgs = await fetch(`${API_BASE}/api/v1/orgs`, {
    headers: { Authorization: `Bearer ${token}` },
  }).then((r) => r.json());
  const current = (orgs.organizations ?? []).find((o: { slug: string }) => o.slug === ORG_SLUG);

  // Page-side base_url is window.location.origin (the web port, not the API
  // port) — that's what wasm-core.ts feeds into `new WasmAuthManager(baseUrl)`.
  // Storage key + base_url field must agree, otherwise bootstrap's
  // BaseUrlMismatch cleanup wipes the session and the test redirects to /login.
  const webOrigin = getWebBaseUrl();
  const storageKey = `agentsmesh-auth/${urlSlug(webOrigin)}/session`;

  await target.addInitScript(
    ({ tok, user, org, key, baseUrl }) => {
      const session = {
        access_token: tok,
        refresh_token: "",
        // Long-lived: 24 h matches the JWT exp the dev backend hands out.
        // bootstrap's `near-expiry` lead is 60 s, so we never trigger refresh.
        expires_at: Math.floor(Date.now() / 1000) + 86400,
        base_url: baseUrl,
        user_id: user?.id ?? 0,
        current_org_slug: org?.slug ?? null,
        schema_version: 1,
      };
      window.localStorage.setItem(key, JSON.stringify(session));
    },
    { tok: token, user: me.user, org: current, key: storageKey, baseUrl: webOrigin },
  );
}

export const test = base.extend<BlockstoreFixtures>({
  token: async ({}, use) => {
    const tok = await sharedLogin();
    await use(tok);
  },
  workspaceID: async ({ token }, use) => {
    const ws = await sharedEnsureWorkspace(token);
    await use(ws);
  },
  isolatedWorkspace: async ({ token }, use) => {
    const ws = await provisionIsolatedWorkspace(token);
    try {
      await use(ws);
    } finally {
      await tearDownIsolatedWorkspace(token, ws.id);
    }
  },
  api: async ({ token }, use) => {
    await use(makeApi(token));
  },
  authenticatedPage: async ({ page, token, workspaceID }, use) => {
    void workspaceID; // ensures default workspace exists before the page loads
    await installApiProxy(page);
    await seedAuth(page, token);
    await use(page);
  },
});

export { expect, installApiProxy, seedAuth };
export const orgSlug = ORG_SLUG;
export const apiBase = API_BASE;
