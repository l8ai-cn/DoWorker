use serde::{Deserialize, Serialize};
use tracing::warn;

use crate::client::ApiClient;

#[derive(Serialize)]
struct RefreshRequest {
    refresh_token: String,
}

#[derive(Deserialize)]
struct RefreshResponse {
    token: String,
    refresh_token: String,
    expires_in: Option<i64>,
}

impl ApiClient {
    pub(crate) async fn handle_token_refresh(&self, failed_token: Option<&str>) -> bool {
        let _guard = self.refresh_lock.lock().await;

        let current_token = self.auth_store.get_token();
        if current_token.as_deref() != failed_token {
            return current_token.is_some();
        }

        let Some(refresh_token) = self.auth_store.get_refresh_token() else {
            warn!("no refresh token available");
            self.auth_store.clear_tokens();
            return false;
        };

        let url = format!("{}/api/v1/auth/refresh", self.base_url);
        let result = self
            .http
            .post(&url)
            .header("Content-Type", "application/json")
            .json(&RefreshRequest { refresh_token })
            .send()
            .await;

        match result {
            Ok(resp) if resp.status().is_success() => match resp.json::<RefreshResponse>().await {
                Ok(data) => {
                    self.auth_store.set_tokens(data.token, data.refresh_token, data.expires_in);
                    true
                }
                Err(e) => {
                    warn!("failed to parse refresh response: {e}");
                    self.auth_store.clear_tokens();
                    false
                }
            },
            Ok(resp) => {
                warn!("token refresh failed with status: {}", resp.status());
                self.auth_store.clear_tokens();
                false
            }
            Err(e) => {
                warn!("token refresh network error: {e}");
                self.auth_store.clear_tokens();
                false
            }
        }
    }
}
