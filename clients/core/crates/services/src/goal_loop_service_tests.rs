use std::sync::Arc;

use agentsmesh_api_client::{ApiClient, AuthTokenStore};

use crate::GoalLoopService;

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
async fn loop_service_rejects_invalid_binary_requests() {
    let service = GoalLoopService::new(Arc::new(ApiClient::new(
        "http://unused".into(),
        Arc::new(TokenStore),
    )));

    let compile_error = service
        .compile_loop_program_connect(&[0xff])
        .await
        .unwrap_err();
    assert!(compile_error.starts_with("decode compile_loop_program request:"));

    let run_error = service.run_loop_program_connect(&[0xff]).await.unwrap_err();
    assert!(run_error.starts_with("decode run_loop_program request:"));
}
