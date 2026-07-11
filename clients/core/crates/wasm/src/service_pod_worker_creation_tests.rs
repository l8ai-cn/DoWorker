use std::sync::Arc;

use agentsmesh_api_client::{ApiClient, AuthTokenStore};

use crate::WasmPodService;

struct TokenStore;

impl AuthTokenStore for TokenStore {
    fn get_token(&self) -> Option<String> {
        None
    }

    fn get_refresh_token(&self) -> Option<String> {
        None
    }

    fn set_tokens(&self, _: String, _: String, _: Option<i64>) {}

    fn clear_tokens(&self) {}

    fn get_current_org_slug(&self) -> Option<String> {
        Some("acme".into())
    }
}

#[tokio::test]
async fn worker_creation_wasm_methods_forward_binary_errors() {
    let service = WasmPodService::new(Arc::new(ApiClient::new(
        "http://unused".into(),
        Arc::new(TokenStore),
    )));

    assert!(service
        .list_worker_create_options_connect(&[0xff])
        .await
        .unwrap_err()
        .starts_with("decode list_worker_create_options request:"));
    assert!(service
        .preflight_worker_connect(&[0xff])
        .await
        .unwrap_err()
        .starts_with("decode preflight_worker request:"));
    assert!(service
        .fill_worker_draft_connect(&[0xff])
        .await
        .unwrap_err()
        .starts_with("decode fill_worker_draft request:"));
}
