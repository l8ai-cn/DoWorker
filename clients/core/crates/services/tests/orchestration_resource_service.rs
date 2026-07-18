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
async fn orchestration_resource_service_rejects_invalid_binary_requests_without_network() {
    let server = MockServer::start().await;
    let service = OrchestrationResourceService::new(Arc::new(ApiClient::new(
        server.uri(),
        Arc::new(TokenStore),
    )));

    assert!(service.validate_resource_connect(&[0xff]).await.is_err());
    assert!(service.plan_resource_connect(&[0xff]).await.is_err());
    assert!(service.get_resource_connect(&[0xff]).await.is_err());
    assert!(service
        .get_resource_capabilities_connect(&[0xff])
        .await
        .is_err());
    assert!(service.list_resources_connect(&[0xff]).await.is_err());
    assert!(service.export_resource_connect(&[0xff]).await.is_err());
    assert!(service.get_resource_plan_connect(&[0xff]).await.is_err());
    assert!(service
        .apply_binding_resource_plan_connect(&[0xff])
        .await
        .is_err());
    assert!(service
        .apply_worker_template_plan_connect(&[0xff])
        .await
        .is_err());
    assert!(service.apply_prompt_plan_connect(&[0xff]).await.is_err());

    assert_eq!(
        server.received_requests().await.unwrap_or_default().len(),
        0
    );
}

#[tokio::test]
async fn orchestration_resource_service_returns_decodable_binary_responses() {
    let server = MockServer::start().await;
    let service = OrchestrationResourceService::new(Arc::new(ApiClient::new(
        server.uri(),
        Arc::new(TokenStore),
    )));

    let validate_request = resource::ValidateResourceRequest {
        org_slug: "acme".into(),
        source: Some(sample_source()),
    };
    let validate_response = resource::ValidateResourceResponse {
        target: Some(sample_target("Widget")),
        operation: resource::ResourceOperation::Create as i32,
        canonical_json: br#"{"kind":"Widget"}"#.to_vec(),
        issues: vec![],
    };
    mount(
        &server,
        "/proto.orchestration_resource.v1.OrchestrationResourceService/ValidateResource",
        validate_request.encode_to_vec(),
        validate_response.encode_to_vec(),
    )
    .await;

    let plan_request = resource::PlanResourceRequest {
        org_slug: "acme".into(),
        source: Some(sample_source()),
    };
    let plan_response = resource::PlanResourceResponse {
        target: Some(sample_target("Widget")),
        operation: resource::ResourceOperation::Update as i32,
        canonical_json: br#"{"kind":"Widget"}"#.to_vec(),
        issues: vec![],
        plan: Some(sample_plan()),
    };
    mount(
        &server,
        "/proto.orchestration_resource.v1.OrchestrationResourceService/PlanResource",
        plan_request.encode_to_vec(),
        plan_response.encode_to_vec(),
    )
    .await;

    let get_request = resource::GetResourceRequest {
        org_slug: "acme".into(),
        target: Some(sample_target("Widget")),
    };
    let get_response = sample_resource();
    mount(
        &server,
        "/proto.orchestration_resource.v1.OrchestrationResourceService/GetResource",
        get_request.encode_to_vec(),
        get_response.encode_to_vec(),
    )
    .await;

    let list_request = resource::ListResourcesRequest {
        org_slug: "acme".into(),
        kind: Some("EnvironmentBundle".into()),
        offset: Some(0),
        limit: Some(10),
        environment_bundle_filter: Some(environment_bundle_filter()),
        model_binding_filter: None,
    };
    let list_response = resource::ListResourcesResponse {
        items: vec![sample_resource()],
        total: 1,
        limit: 10,
        offset: 0,
        applied_environment_bundle_filter: Some(environment_bundle_filter()),
        applied_model_binding_filter: None,
    };
    mount(
        &server,
        "/proto.orchestration_resource.v1.OrchestrationResourceService/ListResources",
        list_request.encode_to_vec(),
        list_response.encode_to_vec(),
    )
    .await;

    let export_request = resource::ExportResourceRequest {
        org_slug: "acme".into(),
        target: Some(sample_target("Widget")),
        revision: Some(7),
        format: resource::SourceFormat::Yaml as i32,
    };
    let export_response = resource::ExportResourceResponse {
        content: b"kind: Widget\n".to_vec(),
    };
    mount(
        &server,
        "/proto.orchestration_resource.v1.OrchestrationResourceService/ExportResource",
        export_request.encode_to_vec(),
        export_response.encode_to_vec(),
    )
    .await;

    let get_plan_request = resource::GetResourcePlanRequest {
        org_slug: "acme".into(),
        plan_id: "plan-1".into(),
    };
    let get_plan_response = sample_plan();
    mount(
        &server,
        "/proto.orchestration_resource.v1.OrchestrationResourceService/GetResourcePlan",
        get_plan_request.encode_to_vec(),
        get_plan_response.encode_to_vec(),
    )
    .await;

    let binding_apply_request = resource::ApplyBindingResourcePlanRequest {
        org_slug: "acme".into(),
        plan_id: "11111111-1111-4111-8111-111111111111".into(),
    };
    let binding_apply_response = sample_resource();
    mount(
        &server,
        "/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyBindingResourcePlan",
        binding_apply_request.encode_to_vec(),
        binding_apply_response.encode_to_vec(),
    )
    .await;

    let worker_apply_request = resource::ApplyWorkerTemplatePlanRequest {
        org_slug: "acme".into(),
        plan_id: "22222222-2222-4222-8222-222222222222".into(),
    };
    let worker_apply_response = resource::ApplyWorkerTemplatePlanResponse {
        resource: Some(sample_resource()),
        worker_spec_snapshot_id: 91,
    };
    mount(
        &server,
        "/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyWorkerTemplatePlan",
        worker_apply_request.encode_to_vec(),
        worker_apply_response.encode_to_vec(),
    )
    .await;

    let prompt_apply_request = resource::ApplyPromptPlanRequest {
        org_slug: "acme".into(),
        plan_id: "33333333-3333-4333-8333-333333333333".into(),
    };
    let prompt_apply_response = sample_resource();
    mount(
        &server,
        "/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyPromptPlan",
        prompt_apply_request.encode_to_vec(),
        prompt_apply_response.encode_to_vec(),
    )
    .await;

    let validate_bytes = service
        .validate_resource_connect(&validate_request.encode_to_vec())
        .await
        .unwrap();
    assert_eq!(
        resource::ValidateResourceResponse::decode(&*validate_bytes)
            .unwrap()
            .canonical_json,
        br#"{"kind":"Widget"}"#
    );

    let plan_bytes = service
        .plan_resource_connect(&plan_request.encode_to_vec())
        .await
        .unwrap();
    assert_eq!(
        resource::PlanResourceResponse::decode(&*plan_bytes)
            .unwrap()
            .plan
            .unwrap()
            .plan_id,
        "plan-1"
    );

    let get_bytes = service
        .get_resource_connect(&get_request.encode_to_vec())
        .await
        .unwrap();
    assert_eq!(resource::Resource::decode(&*get_bytes).unwrap().id, 42);

    let list_bytes = service
        .list_resources_connect(&list_request.encode_to_vec())
        .await
        .unwrap();
    let list_response = resource::ListResourcesResponse::decode(&*list_bytes).unwrap();
    assert_eq!(list_response.items[0].display_name, "Widget One");
    assert_eq!(
        list_response
            .applied_environment_bundle_filter
            .as_ref()
            .map(|filter| filter.worker_type.as_str()),
        Some("do-agent")
    );
    assert_eq!(
        list_response
            .applied_environment_bundle_filter
            .as_ref()
            .map(|filter| filter.target_name.as_str()),
        Some("DO_API_KEY")
    );

    let export_bytes = service
        .export_resource_connect(&export_request.encode_to_vec())
        .await
        .unwrap();
    assert_eq!(
        resource::ExportResourceResponse::decode(&*export_bytes)
            .unwrap()
            .content,
        b"kind: Widget\n"
    );

    let get_plan_bytes = service
        .get_resource_plan_connect(&get_plan_request.encode_to_vec())
        .await
        .unwrap();
    assert_eq!(
        resource::ResourcePlan::decode(&*get_plan_bytes)
            .unwrap()
            .plan_hash,
        "plan-hash"
    );

    let binding_apply_bytes = service
        .apply_binding_resource_plan_connect(&binding_apply_request.encode_to_vec())
        .await
        .unwrap();
    assert_eq!(
        resource::Resource::decode(&*binding_apply_bytes)
            .unwrap()
            .id,
        42
    );

    let worker_apply_bytes = service
        .apply_worker_template_plan_connect(&worker_apply_request.encode_to_vec())
        .await
        .unwrap();
    let worker_apply =
        resource::ApplyWorkerTemplatePlanResponse::decode(&*worker_apply_bytes).unwrap();
    assert_eq!(worker_apply.resource.unwrap().id, 42);
    assert_eq!(worker_apply.worker_spec_snapshot_id, 91);

    let prompt_apply_bytes = service
        .apply_prompt_plan_connect(&prompt_apply_request.encode_to_vec())
        .await
        .unwrap();
    assert_eq!(
        resource::Resource::decode(&*prompt_apply_bytes).unwrap().id,
        42
    );
}

async fn mount(server: &MockServer, procedure: &'static str, request: Vec<u8>, response: Vec<u8>) {
    Mock::given(method("POST"))
        .and(path(procedure))
        .and(body_bytes(request))
        .respond_with(ResponseTemplate::new(200).set_body_bytes(response))
        .mount(server)
        .await;
}

fn environment_bundle_filter() -> resource::EnvironmentBundleReferenceFilter {
    resource::EnvironmentBundleReferenceFilter {
        purpose: resource::EnvironmentBundlePurpose::Credential as i32,
        worker_type: "do-agent".into(),
        target_name: "DO_API_KEY".into(),
    }
}

fn sample_source() -> resource::ResourceSource {
    resource::ResourceSource {
        format: resource::SourceFormat::Json as i32,
        content: br#"{"kind":"Widget"}"#.to_vec(),
    }
}

fn sample_target(kind: &str) -> resource::ResourceTarget {
    resource::ResourceTarget {
        type_meta: Some(resource::TypeMeta {
            api_version: "orchestration.do/v1".into(),
            kind: kind.into(),
        }),
        namespace: "default".into(),
        name: "widget-1".into(),
    }
}

fn sample_resource() -> resource::Resource {
    resource::Resource {
        id: 42,
        identity: Some(resource::ResourceIdentity {
            target: Some(sample_target("Widget")),
            uid: "uid-42".into(),
        }),
        display_name: "Widget One".into(),
        labels: [("tier".into(), "gold".into())].into(),
        status_json: br#"{"ready":true}"#.to_vec(),
        revision: 7,
        generation: 3,
        resource_version: 9,
        created_by_id: 1,
        updated_by_id: 2,
        created_at: "2026-07-14T00:00:00Z".into(),
        updated_at: "2026-07-14T00:00:01Z".into(),
    }
}

fn sample_plan() -> resource::ResourcePlan {
    resource::ResourcePlan {
        plan_id: "plan-1".into(),
        operation: resource::ResourceOperation::Update as i32,
        target: Some(sample_target("Widget")),
        base: Some(resource::ResourceIdentity {
            target: Some(sample_target("Widget")),
            uid: "uid-42".into(),
        }),
        base_resource_version: 9,
        draft_hash: "draft-hash".into(),
        plan_hash: "plan-hash".into(),
        artifact_digest: "sha256:abc".into(),
        resolved_references: vec![],
        semantic_diff: vec![],
        issues: vec![],
        artifact_kind: "Widget".into(),
        options_revision: "rev-1".into(),
        created_at: "2026-07-14T00:00:00Z".into(),
        expires_at: "2026-07-14T01:00:00Z".into(),
        status: resource::PlanStatus::Pending as i32,
    }
}
