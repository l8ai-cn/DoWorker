use agentsmesh_types::{Organization, User};
use serde::{Deserialize, Serialize};

use crate::error::AuthError;
use crate::manager::{now_unix_secs, AuthManager};
use crate::state::LEGACY_STORAGE_KEY;

const REFRESH_LEAD_SECS: i64 = 60;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "kind", rename_all = "snake_case")]
pub enum BootstrapResult {
    Anonymous,
    Authenticated {
        user: User,
        current_org: Option<Organization>,
    },
    AnonymousAfterCleanup {
        reason: BootstrapCleanupReason,
    },
}

/// First disqualifying check encountered during `bootstrap()`. The
/// protocol short-circuits at the first failure (legacy purge → schema
/// validity → base_url match → token freshness → identity 401), so this
/// enum carries one *primary* reason — never a composite. UI / telemetry
/// should treat the variants as mutually exclusive labels for the step
/// where bootstrap decided to drop local state.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum BootstrapCleanupReason {
    /// Stored `base_url` doesn't match the manager's current `base_url`.
    /// Happens after a server switch when an old namespaced blob survives.
    BaseUrlMismatch,
    /// Access token expired (or about to within `REFRESH_LEAD_SECS`) and
    /// the refresh attempt against `/auth/refresh` itself failed (network,
    /// 401 from server, malformed response).
    TokenExpiredAndRefreshFailed,
    /// `/users/me` or `/users/me/organizations` returned 401 even though
    /// the local token wasn't yet expired — server has revoked it.
    UnauthorizedFromIdentityCall,
    /// `serde_json::from_str::<PersistedSession>` failed — disk blob was
    /// truncated, hand-edited, or written by an incompatible schema.
    StorageCorrupt,
    /// One-shot migration: legacy global `agentsmesh-auth` key existed
    /// alongside (or instead of) the new namespaced key. Always purged
    /// on first bootstrap of a v0.32+ build.
    LegacyDataPurged,
}

impl AuthManager {
    pub async fn bootstrap(&self) -> BootstrapResult {
        let purged_legacy = self.storage.get(LEGACY_STORAGE_KEY).is_some();
        if purged_legacy {
            self.storage.remove(LEGACY_STORAGE_KEY);
        }

        let session_json = match self.storage.get(&self.session_key()) {
            Some(j) => j,
            None => {
                return if purged_legacy {
                    BootstrapResult::AnonymousAfterCleanup {
                        reason: BootstrapCleanupReason::LegacyDataPurged,
                    }
                } else {
                    BootstrapResult::Anonymous
                };
            }
        };

        let restored: crate::state::PersistedSession = match serde_json::from_str(&session_json) {
            Ok(s) => s,
            Err(e) => {
                tracing::warn!("auth bootstrap: storage corrupt: {e}");
                return self.cleanup(BootstrapCleanupReason::StorageCorrupt);
            }
        };

        if !restored.base_url.is_empty() && restored.base_url != self.base_url {
            tracing::warn!(
                "auth bootstrap: base_url mismatch (stored={}, current={})",
                restored.base_url,
                self.base_url,
            );
            return self.cleanup(BootstrapCleanupReason::BaseUrlMismatch);
        }
        if restored.access_token.is_empty() {
            return self.cleanup(BootstrapCleanupReason::StorageCorrupt);
        }

        self.write_state().restore_persisted(restored.clone());

        if restored.expires_at <= now_unix_secs() + REFRESH_LEAD_SECS {
            if let Err(e) = self.refresh_token().await {
                tracing::warn!("auth bootstrap: refresh failed: {e}");
                return self.cleanup(BootstrapCleanupReason::TokenExpiredAndRefreshFailed);
            }
        }

        let user = match self.fetch_me().await {
            Ok(u) => u,
            Err(AuthError::Server { status: 401, .. }) | Err(AuthError::NotAuthenticated) => {
                return self.cleanup(BootstrapCleanupReason::UnauthorizedFromIdentityCall);
            }
            Err(e) => {
                // Transient error (network, 5xx) — keep session, present anonymous
                // for now so caller can retry; storage is not wiped.
                tracing::warn!("auth bootstrap: identity transient failure: {e}");
                return BootstrapResult::Anonymous;
            }
        };

        let orgs = match self.fetch_organizations().await {
            Ok(o) => o,
            Err(AuthError::Server { status: 401, .. }) | Err(AuthError::NotAuthenticated) => {
                // Token rejected mid-bootstrap (server revoked between
                // /users/me and /users/me/organizations). Same cleanup
                // semantics as the identity-call 401 path.
                return self.cleanup(BootstrapCleanupReason::UnauthorizedFromIdentityCall);
            }
            Err(e) => {
                // Transient (5xx, network) — we already have user, so the
                // dashboard can render with empty orgs and let the user
                // retry. Don't wipe the session.
                tracing::warn!("auth bootstrap: orgs transient failure: {e}");
                Vec::new()
            }
        };

        let current_org = restored
            .current_org_slug
            .as_ref()
            .and_then(|slug| orgs.iter().find(|o| &o.slug == slug).cloned())
            .or_else(|| orgs.first().cloned());

        if let Some(ref org) = current_org {
            self.set_current_org(Some(org.clone()));
        }

        BootstrapResult::Authenticated { user, current_org }
    }

    fn cleanup(&self, reason: BootstrapCleanupReason) -> BootstrapResult {
        self.reset_local();
        BootstrapResult::AnonymousAfterCleanup { reason }
    }
}
