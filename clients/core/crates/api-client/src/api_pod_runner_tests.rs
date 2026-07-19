use std::sync::Arc;

use agentsmesh_types::proto_pod_v1 as pod;
use prost::Message;
use wiremock::matchers::{body_bytes, method, path};
use wiremock::{Mock, MockServer, ResponseTemplate};

use crate::{ApiClient, AuthTokenStore};

struct TokenStore;

impl AuthTokenStore for TokenStore {
    fn get_token(&self) -> Option<String> {
        Some("token".into())
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
async fn worker_creation_connect_methods_use_binary_wire() {
    let server = MockServer::start().await;
    let options_request = pod::ListWorkerCreateOptionsRequest {
        org_slug: "acme".into(),
        worker_type_slug: Some("codex".into()),
        compute_target_id: Some(11),
        deployment_mode: Some("pooled".into()),
    };
    let options_response = pod::ListWorkerCreateOptionsResponse {
        revision: "rev-1".into(),
        ..Default::default()
    };
    mount_binary_rpc(
        &server,
        "/proto.pod.v1.PodService/ListWorkerCreateOptions",
        options_request.encode_to_vec(),
        options_response.encode_to_vec(),
    )
    .await;

    let preflight_request = pod::PreflightWorkerRequest {
        org_slug: "acme".into(),
        draft: Some(worker_draft()),
    };
    let preflight_response = pod::PreflightWorkerResponse {
        options_revision: "rev-1".into(),
        resolved_spec_json: Some(r#"{"version":1}"#.into()),
        ..Default::default()
    };
    mount_binary_rpc(
        &server,
        "/proto.pod.v1.PodService/PreflightWorker",
        preflight_request.encode_to_vec(),
        preflight_response.encode_to_vec(),
    )
    .await;

    let fill_request = pod::FillWorkerDraftRequest {
        org_slug: "acme".into(),
        prompt: "Create a coding worker".into(),
        current_draft: Some(worker_draft()),
        generation_model_resource_id: 77,
    };
    let fill_response = pod::FillWorkerDraftResponse {
        draft: Some(worker_draft()),
        issues: Vec::new(),
    };
    mount_binary_rpc(
        &server,
        "/proto.pod.v1.PodService/FillWorkerDraft",
        fill_request.encode_to_vec(),
        fill_response.encode_to_vec(),
    )
    .await;

    let client = ApiClient::new(server.uri(), Arc::new(TokenStore));
    assert_eq!(
        client
            .list_worker_create_options_connect(&options_request)
            .await
            .unwrap()
            .revision,
        "rev-1"
    );
    assert_eq!(
        client
            .preflight_worker_connect(&preflight_request)
            .await
            .unwrap()
            .resolved_spec_json
            .as_deref(),
        Some(r#"{"version":1}"#)
    );
    assert_eq!(
        client
            .fill_worker_draft_connect(&fill_request)
            .await
            .unwrap()
            .draft
            .unwrap()
            .worker_type_slug,
        "codex"
    );
}

async fn mount_binary_rpc(
    server: &MockServer,
    procedure: &'static str,
    request: Vec<u8>,
    response: Vec<u8>,
) {
    Mock::given(method("POST"))
        .and(path(procedure))
        .and(body_bytes(request))
        .respond_with(ResponseTemplate::new(200).set_body_bytes(response))
        .mount(server)
        .await;
}

fn worker_draft() -> pod::WorkerSpecDraft {
    pod::WorkerSpecDraft {
        worker_type_slug: "codex".into(),
        options_revision: "rev-1".into(),
        ..Default::default()
    }
}
