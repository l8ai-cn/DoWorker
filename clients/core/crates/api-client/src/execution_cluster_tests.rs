use std::sync::Arc;

use agentcloud_types::proto_execution_cluster_v1 as cluster;
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
        Some("dev-org".into())
    }
}

#[tokio::test]
async fn execution_cluster_methods_use_binary_connect_wire() {
    let server = MockServer::start().await;
    let list_request = cluster::ListExecutionClustersRequest {
        org_slug: "dev-org".into(),
    };
    let list_response = cluster::ListExecutionClustersResponse {
        items: vec![cluster::ExecutionCluster {
            id: 12,
            slug: "local".into(),
            ..Default::default()
        }],
    };
    let command_request = cluster::CreateRegistrationCommandRequest {
        org_slug: "dev-org".into(),
        cluster_id: 12,
        node_name: "mac-studio".into(),
    };
    let command_response = cluster::CreateRegistrationCommandResponse {
        command: "runner register --server https://example.test --token secret".into(),
        expires_at: "2026-07-12T12:15:00Z".into(),
    };
    mount(
        &server,
        "/proto.execution_cluster.v1.ExecutionClusterService/ListExecutionClusters",
        list_request.encode_to_vec(),
        list_response.encode_to_vec(),
    )
    .await;
    mount(
        &server,
        "/proto.execution_cluster.v1.ExecutionClusterService/CreateRegistrationCommand",
        command_request.encode_to_vec(),
        command_response.encode_to_vec(),
    )
    .await;

    let client = ApiClient::new(server.uri(), Arc::new(TokenStore));

    assert_eq!(
        client
            .list_execution_clusters_connect(&list_request)
            .await
            .unwrap()
            .items[0]
            .slug,
        "local"
    );
    assert_eq!(
        client
            .create_execution_cluster_registration_command_connect(&command_request)
            .await
            .unwrap()
            .expires_at,
        "2026-07-12T12:15:00Z"
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
