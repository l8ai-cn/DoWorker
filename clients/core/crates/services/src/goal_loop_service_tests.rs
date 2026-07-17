use std::sync::Arc;

use agentsmesh_api_client::{ApiClient, AuthTokenStore};
use agentsmesh_types::proto_goalloop_v1 as lp;
use prost::Message;
use wiremock::matchers::{body_bytes, method, path};
use wiremock::{Mock, MockServer, ResponseTemplate};

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

    let generate_error = service
        .generate_loop_program_connect(&[0xff])
        .await
        .unwrap_err();
    assert!(generate_error.starts_with("decode generate_loop_program request:"));

    let repair_error = service
        .repair_loop_program_connect(&[0xff])
        .await
        .unwrap_err();
    assert!(repair_error.starts_with("decode repair_loop_program request:"));

    let run_error = service.run_loop_program_connect(&[0xff]).await.unwrap_err();
    assert!(run_error.starts_with("decode run_loop_program request:"));
}

#[tokio::test]
async fn loop_service_forwards_generate_request_without_diagnostics() {
    let server = MockServer::start().await;
    let request = lp::GenerateLoopProgramRequest {
        org_slug: "acme".into(),
        prompt: "制作专业 PPT".into(),
        current_source: "loop current {}".into(),
        model_resource_id: 42,
        locale: "zh-CN".into(),
        revision: 7,
    };
    let response = lp::CompileLoopProgramResponse {
        canonical_source: "loop generated {}".into(),
        revision: 7,
        ..Default::default()
    };
    Mock::given(method("POST"))
        .and(path(
            "/proto.goalloop.v1.GoalLoopService/GenerateLoopProgram",
        ))
        .and(body_bytes(request.encode_to_vec()))
        .respond_with(ResponseTemplate::new(200).set_body_bytes(response.encode_to_vec()))
        .mount(&server)
        .await;

    let service =
        GoalLoopService::new(Arc::new(ApiClient::new(server.uri(), Arc::new(TokenStore))));
    let bytes = service
        .generate_loop_program_connect(&request.encode_to_vec())
        .await
        .unwrap();
    let decoded = lp::CompileLoopProgramResponse::decode(&*bytes).unwrap();

    assert_eq!(decoded.canonical_source, "loop generated {}");
    assert_eq!(decoded.revision, 7);
}

#[tokio::test]
async fn loop_service_forwards_targeted_repair_request() {
    let server = MockServer::start().await;
    let request = lp::RepairLoopProgramRequest {
        org_slug: "acme".into(),
        source: "loop current {}".into(),
        model_resource_id: 42,
        locale: "zh-CN".into(),
        revision: 7,
        diagnostic_code: "loop.value.out-of-range".into(),
        node_id: "n-limits".into(),
        field_path: "limits.iterations".into(),
        prompt: "保持预算严格".into(),
    };
    let response = lp::RepairLoopProgramResponse {
        proposal: Some(lp::CompileLoopProgramResponse {
            canonical_source: "loop repaired {}".into(),
            revision: 7,
            ..Default::default()
        }),
        patch: Some(lp::LoopIntegerPatch {
            node_id: "n-limits".into(),
            field_path: "limits.iterations".into(),
            old_value: 100,
            new_value: 20,
        }),
    };
    Mock::given(method("POST"))
        .and(path(
            "/proto.goalloop.v1.GoalLoopService/RepairLoopProgram",
        ))
        .and(body_bytes(request.encode_to_vec()))
        .respond_with(ResponseTemplate::new(200).set_body_bytes(response.encode_to_vec()))
        .mount(&server)
        .await;

    let service =
        GoalLoopService::new(Arc::new(ApiClient::new(server.uri(), Arc::new(TokenStore))));
    let bytes = service
        .repair_loop_program_connect(&request.encode_to_vec())
        .await
        .unwrap();
    let decoded = lp::RepairLoopProgramResponse::decode(&*bytes).unwrap();

    assert_eq!(
        decoded.proposal.unwrap().canonical_source,
        "loop repaired {}"
    );
    assert_eq!(decoded.patch.unwrap().new_value, 20);
}
