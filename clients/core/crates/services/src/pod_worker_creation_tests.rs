use std::sync::Arc;

use agentsmesh_api_client::{ApiClient, AuthTokenStore};

use crate::PodService;

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
async fn worker_creation_service_rejects_invalid_binary_requests() {
    let service = PodService::new(Arc::new(ApiClient::new(
        "http://unused".into(),
        Arc::new(TokenStore),
    )));

    let options_error = service
        .list_worker_create_options_connect(&[0xff])
        .await
        .unwrap_err();
    assert!(options_error.starts_with("decode list_worker_create_options request:"));

    let preflight_error = service.preflight_worker_connect(&[0xff]).await.unwrap_err();
    assert!(preflight_error.starts_with("decode preflight_worker request:"));

    let fill_error = service
        .fill_worker_draft_connect(&[0xff])
        .await
        .unwrap_err();
    assert!(fill_error.starts_with("decode fill_worker_draft request:"));
}
