use std::sync::Arc;

use agentsmesh_api_client::{ApiClient, AuthTokenStore};
use agentsmesh_services::OrchestrationResourceService;
use orchestration_resource_proto::proto::orchestration_resource::v1 as resource;
use prost::Message;
use wiremock::matchers::{body_bytes, method, path};
use wiremock::{Mock, MockServer, ResponseTemplate};

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
async fn expert_apply_service_decodes_and_encodes_typed_messages() {
    let server = MockServer::start().await;
    let service = OrchestrationResourceService::new(Arc::new(ApiClient::new(
        server.uri(),
        Arc::new(TokenStore),
    )));
    assert!(service.apply_expert_plan_connect(&[0xff]).await.is_err());

    let request = resource::ApplyExpertPlanRequest {
        org_slug: "acme".into(),
        plan_id: "11111111-1111-4111-8111-111111111111".into(),
    };
    let response = resource::ApplyExpertPlanResponse {
        resource: None,
        expert_id: 81,
        worker_spec_snapshot_id: 91,
        resource_revision: 3,
    };
    Mock::given(method("POST"))
        .and(path(
            "/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyExpertPlan",
        ))
        .and(body_bytes(request.encode_to_vec()))
        .respond_with(ResponseTemplate::new(200).set_body_bytes(response.encode_to_vec()))
        .mount(&server)
        .await;

    let bytes = service
        .apply_expert_plan_connect(&request.encode_to_vec())
        .await
        .expect("apply expert");
    let applied = resource::ApplyExpertPlanResponse::decode(&*bytes).expect("decode response");

    assert_eq!(applied.expert_id, 81);
    assert_eq!(applied.worker_spec_snapshot_id, 91);
    assert_eq!(applied.resource_revision, 3);
}
