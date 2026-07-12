/**
 * Client for the ``accounts`` auth provider's HTTP API.
 *
 * Wraps the shared Connect authentication procedures in a small typed surface so the LoginPage,
 * MembersPage, and any future profile-management UI share one
 * source of truth for the request/response shapes.
 *
 * Do Worker stores JWT in localStorage (`do-worker-auth`, legacy `agentsmesh-auth`) after
 * login; authenticated procedures use Bearer auth.
 *
 * Errors: every helper resolves with a typed error object on
 * non-2xx instead of throwing, so the UI can render specific
 * messages (wrong password vs network failure vs server error)
 * without try/catch every call site.
 */

import { clearDoWorkerSession, readDoWorkerJWT } from "@/lib/do-worker/auth-session";

const LOGIN_PROCEDURE = "/proto.auth.v1.AuthService/Login";
const LOGOUT_PROCEDURE = "/proto.auth.v1.AuthSessionService/Logout";
const GET_ME_PROCEDURE = "/proto.user.v1.UserService/GetMe";
const LIST_MY_ORGS_PROCEDURE = "/proto.org.v1.OrgService/ListMyOrgs";

/** Body of AuthService.Login. */
export interface LoginRequest {
  username: string;
  password: string;
}

/** Successful AuthService.Login response. */
export interface LoginSuccess {
  ok: true;
  user: { id: string; is_admin: boolean };
  token: string;
  refresh_token?: string;
  expires_in: number;
  org_slug?: string;
}

/** Login failure — kept opaque on purpose (don't leak which check failed). */
export interface LoginFailure {
  ok: false;
  /** Short human-readable message safe to show in the form. */
  error: string;
  /** HTTP status, in case the UI wants to distinguish 401 vs 5xx. */
  status: number;
}

export type LoginResult = LoginSuccess | LoginFailure;

export async function listMyOrgSlug(token: string): Promise<string | null> {
  const res = await fetch(LIST_MY_ORGS_PROCEDURE, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "content-type": "application/json",
    },
    body: "{}",
  });
  if (!res.ok) {
    throw new Error(`Organization lookup failed with status ${res.status}`);
  }
  const data = (await res.json()) as unknown;
  if (!data || typeof data !== "object" || !Array.isArray((data as { items?: unknown }).items)) {
    throw new Error("Organization lookup returned an invalid response");
  }
  const first = (data as { items: unknown[] }).items[0];
  if (first === undefined) return null;
  if (
    !first ||
    typeof first !== "object" ||
    typeof (first as { slug?: unknown }).slug !== "string"
  ) {
    throw new Error("Organization lookup returned an invalid organization");
  }
  return (first as { slug: string }).slug;
}

/** Shape of UserService.GetMe when authenticated. */
export interface CurrentAccount {
  id: string;
  is_admin: boolean;
  created_at: number | null;
  last_login_at: number | null;
}

/**
 * AuthService.Login — verify username + password using the shared public
 * authentication protocol.
 *
 * :param body: Login credentials.
 * :returns: Discriminated union — ``ok: true`` with the user info,
 *     or ``ok: false`` with an error message.
 */
export async function login(body: LoginRequest): Promise<LoginResult> {
  let res: Response;
  try {
    res = await fetch(LOGIN_PROCEDURE, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(body),
    });
  } catch {
    return {
      ok: false,
      error: "Could not reach the server. Check your connection.",
      status: 0,
    };
  }

  if (res.ok) {
    const data = (await res.json()) as {
      token: string;
      refreshToken?: string;
      expiresIn: string | number;
      user?: { id: string | number; isSystemAdmin?: boolean };
    };
    if (!data.token || !data.user || !Number.isFinite(Number(data.expiresIn))) {
      return { ok: false, error: "Invalid login response.", status: 502 };
    }
    return {
      ok: true,
      token: data.token,
      refresh_token: data.refreshToken,
      expires_in: Number(data.expiresIn),
      user: { id: String(data.user.id), is_admin: Boolean(data.user.isSystemAdmin) },
    };
  }

  // The route returns 401 for both unknown-user and wrong-password.
  // Surface the server's message when it's a 4xx, generic for 5xx.
  let message = "Login failed.";
  if (res.status >= 500) {
    message = "Server error. Try again in a moment.";
  } else {
    try {
      const data = (await res.json()) as { error?: string };
      if (data.error) {
        message = data.error;
      }
    } catch {
      // Body wasn't JSON; keep the generic message.
    }
  }
  return { ok: false, error: message, status: res.status };
}

/**
 * AuthSessionService.Logout — revoke the current bearer token.
 */
export async function logout(): Promise<void> {
  const headers: HeadersInit = {};
  const jwt = readDoWorkerJWT();
  if (jwt) headers.Authorization = `Bearer ${jwt}`;
  try {
    await fetch(LOGOUT_PROCEDURE, {
      method: "POST",
      headers: { ...headers, "content-type": "application/json" },
      body: "{}",
    });
  } catch {
    // Network error — clear local session anyway.
  }
  clearDoWorkerSession();
}

/**
 * UserService.GetMe — fetch the current user.
 *
 * :returns: The current :class:`CurrentAccount`, or ``null`` if
 *     unauthenticated.
 */
export async function getMe(): Promise<CurrentAccount | null> {
  const headers: HeadersInit = {};
  const jwt = readDoWorkerJWT();
  if (jwt) headers.Authorization = `Bearer ${jwt}`;
  let res: Response;
  try {
    res = await fetch(GET_ME_PROCEDURE, {
      method: "POST",
      headers: { ...headers, "content-type": "application/json" },
      body: "{}",
      cache: "no-store",
    });
  } catch {
    return null;
  }
  if (res.ok) {
    const data = (await res.json()) as {
      id?: string | number;
      isSystemAdmin?: boolean;
      createdAt?: string;
      lastLoginAt?: string;
    };
    if (data.id === undefined) return null;
    return {
      id: String(data.id),
      is_admin: Boolean(data.isSystemAdmin),
      created_at: toEpochSeconds(data.createdAt),
      last_login_at: toEpochSeconds(data.lastLoginAt),
    };
  }
  return null;
}

function toEpochSeconds(value: string | undefined): number | null {
  if (!value) return null;
  const epoch = Date.parse(value);
  return Number.isNaN(epoch) ? null : Math.floor(epoch / 1000);
}

/** Body of POST /auth/register. */
export interface RegisterRequest {
  invite: string;
  username: string;
  password: string;
}

/**
 * POST /auth/register — redeem an invite token and create the user.
 *
 * Same response shape as :func:`login` (cookie set on success) so
 * the calling page can navigate straight to ``/`` after.
 */
export async function register(body: RegisterRequest): Promise<LoginResult> {
  let res: Response;
  try {
    res = await fetch("/auth/register", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(body),
    });
  } catch {
    return {
      ok: false,
      error: "Could not reach the server.",
      status: 0,
    };
  }
  if (res.ok) {
    const data = (await res.json()) as Omit<LoginSuccess, "ok">;
    return { ok: true, ...data };
  }
  let message = "Registration failed.";
  if (res.status >= 500) {
    message = "Server error. Try again in a moment.";
  } else {
    try {
      const data = (await res.json()) as { error?: string };
      if (data.error) message = data.error;
    } catch {
      // pass
    }
  }
  return { ok: false, error: message, status: res.status };
}

/** Body of POST /auth/users/me/password (self-serve password change). */
export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

/** Result of a self-serve password change. */
export type ChangePasswordResult = { ok: true } | { ok: false; error: string };

/**
 * POST /auth/users/me/password — change the signed-in user's own password.
 *
 * Requires the current password (the server re-verifies it). Returns
 * 204 on success. Maps the server's status codes to user-facing
 * messages: 401 → wrong current password, 400 → account has no
 * password (header/OIDC identity), 5xx → server error.
 *
 * :param body: ``{old_password, new_password}``.
 * :returns: ``{ok: true}`` or ``{ok: false, error}``.
 */
export async function changePassword(body: ChangePasswordRequest): Promise<ChangePasswordResult> {
  let res: Response;
  try {
    res = await fetch("/auth/users/me/password", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(body),
    });
  } catch {
    return { ok: false, error: "Could not reach the server." };
  }
  if (res.ok) return { ok: true };
  if (res.status === 401) {
    return { ok: false, error: "Current password is incorrect." };
  }
  if (res.status >= 500) {
    return { ok: false, error: "Server error. Try again in a moment." };
  }
  let message = "Could not change password.";
  try {
    const data = (await res.json()) as { error?: string };
    if (data.error) message = data.error;
  } catch {
    // pass
  }
  return { ok: false, error: message };
}

/** Body of POST /auth/setup (first-run admin claim). */
export interface SetupRequest {
  username: string;
  password: string;
}

/**
 * POST /auth/setup — claim the first admin on a fresh instance.
 *
 * Only valid while no account exists (the server 409s once one does).
 * Same success shape as :func:`login` (cookie set on success) so the
 * page can navigate to ``/`` after.
 */
export async function setup(body: SetupRequest): Promise<LoginResult> {
  let res: Response;
  try {
    res = await fetch("/auth/setup", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(body),
    });
  } catch {
    return { ok: false, error: "Could not reach the server.", status: 0 };
  }
  if (res.ok) {
    const data = (await res.json()) as Omit<LoginSuccess, "ok">;
    return { ok: true, ...data };
  }
  if (res.status === 409) {
    return {
      ok: false,
      error: "This server already has an admin. Sign in instead.",
      status: 409,
    };
  }
  let message = "Could not create the admin account.";
  if (res.status >= 500) {
    message = "Server error. Try again in a moment.";
  } else {
    try {
      const data = (await res.json()) as { error?: string };
      if (data.error) message = data.error;
    } catch {
      // pass
    }
  }
  return { ok: false, error: message, status: res.status };
}

// ── Admin: members management ──────────────────────────────────────

/**
 * A user row as returned by ``GET /auth/users``.
 *
 * Same shape as :class:`CurrentAccount` plus ``has_password``
 * (so the UI can render a "External login" badge for header/OIDC
 * rows that haven't been converted to accounts).
 */
export interface AccountListEntry {
  id: string;
  is_admin: boolean;
  created_at: number | null;
  last_login_at: number | null;
  has_password: boolean;
}

/** Successful response from ``POST /auth/invite``. */
export interface InviteCreated {
  ok: true;
  token: string;
  register_url: string;
  expires_at: number;
  is_admin: boolean;
}

/** Successful response from ``POST /auth/users/{id}/reset``. */
export interface PasswordReset {
  ok: true;
  id: string;
  new_password: string;
}

/** Generic admin operation failure. */
export interface AdminFailure {
  ok: false;
  error: string;
  status: number;
}

/**
 * Wrap a generic admin response, mapping non-2xx to typed failure.
 *
 * Centralized so each admin call site has the same error shape.
 * Network failures collapse to ``status: 0`` per the convention
 * already established by :func:`login`.
 */
async function _admin<T extends { ok: true }>(
  doFetch: () => Promise<Response>,
  toSuccess: (body: unknown) => Omit<T, "ok">,
): Promise<T | AdminFailure> {
  let res: Response;
  try {
    res = await doFetch();
  } catch {
    return { ok: false, error: "Could not reach the server.", status: 0 };
  }
  if (res.ok) {
    const body = await res.json();
    return { ok: true, ...toSuccess(body) } as T;
  }
  let message = `Request failed (${res.status}).`;
  if (res.status === 403) message = "Admin permission required.";
  else if (res.status === 404) message = "Not found.";
  try {
    const data = (await res.json()) as { error?: string };
    if (data.error) message = data.error;
  } catch {
    // Body wasn't JSON; keep the generic message.
  }
  return { ok: false, error: message, status: res.status };
}

/**
 * GET /auth/users — admin-only listing of every account.
 *
 * Returns ``null`` on 403 / network error so the caller can fall
 * back gracefully (e.g. hide the members page entirely for
 * non-admins instead of throwing).
 */
export async function listUsers(): Promise<AccountListEntry[] | null> {
  let res: Response;
  try {
    res = await fetch("/auth/users", { cache: "no-store" });
  } catch {
    return null;
  }
  if (!res.ok) return null;
  const data = (await res.json()) as { users: AccountListEntry[] };
  return data.users;
}

/**
 * POST /auth/invite — mint a single-use invite token (admin only).
 *
 * :param isAdmin: Whether the resulting user is created with admin
 *     rights. Defaults false; the modal flips it via a checkbox.
 * :returns: The new token + the URL to share, or a typed failure.
 */
export async function createInvite(isAdmin: boolean): Promise<InviteCreated | AdminFailure> {
  return _admin<InviteCreated>(
    () =>
      fetch("/auth/invite", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ is_admin: isAdmin }),
      }),
    (body) => {
      const b = body as {
        token: string;
        register_url: string;
        expires_at: number;
        is_admin: boolean;
      };
      return {
        token: b.token,
        register_url: b.register_url,
        expires_at: b.expires_at,
        is_admin: b.is_admin,
      };
    },
  );
}

/**
 * DELETE /auth/users/{id} — remove a user (admin only).
 *
 * Server rejects self-delete and bootstrap-admin delete — those
 * surface as 400 with explanatory ``error`` strings, propagated
 * through :type:`AdminFailure`.
 */
export async function deleteUser(userId: string): Promise<{ ok: true } | AdminFailure> {
  let res: Response;
  try {
    res = await fetch(`/auth/users/${encodeURIComponent(userId)}`, {
      method: "DELETE",
    });
  } catch {
    return { ok: false, error: "Could not reach the server.", status: 0 };
  }
  if (res.status === 204) return { ok: true };
  let message = `Delete failed (${res.status}).`;
  try {
    const data = (await res.json()) as { error?: string };
    if (data.error) message = data.error;
  } catch {
    // pass
  }
  return { ok: false, error: message, status: res.status };
}

/**
 * POST /auth/users/{id}/reset — admin-issued password reset.
 *
 * Returns the freshly generated plaintext exactly once. The admin
 * is responsible for DM-ing it to the user out-of-band.
 */
export async function resetUserPassword(userId: string): Promise<PasswordReset | AdminFailure> {
  return _admin<PasswordReset>(
    () =>
      fetch(`/auth/users/${encodeURIComponent(userId)}/reset`, {
        method: "POST",
      }),
    (body) => {
      const b = body as { id: string; new_password: string };
      return { id: b.id, new_password: b.new_password };
    },
  );
}
