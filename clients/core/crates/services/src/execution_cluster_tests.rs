use std::sync::Arc;

use agentcloud_api_client::{ApiClient, AuthTokenStore};
use agentcloud_types::proto_execution_cluster_v1 as cluster;
use prost::Message;
use wiremock::matchers::{body_bytes, method, path};
use wiremock::{Mock, MockServer, ResponseTemplate};

use crate::ExecutionClusterService;

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
async fn execution_cluster_service_forwards_binary_connect_response() {
    let server = MockServer::start().await;
    let request = cluster::ListExecutionClustersRequest {
        org_slug: "dev-org".into(),
    };
    let response = cluster::ListExecutionClustersResponse {
        items: vec![cluster::ExecutionCluster {
            id: 12,
            slug: "online".into(),
            ..Default::default()
        }],
    };
    Mock::given(method("POST"))
        .and(path(
            "/proto.execution_cluster.v1.ExecutionClusterService/ListExecutionClusters",
        ))
        .and(body_bytes(request.encode_to_vec()))
        .respond_with(ResponseTemplate::new(200).set_body_bytes(response.encode_to_vec()))
        .mount(&server)
        .await;

    let service =
        ExecutionClusterService::new(Arc::new(ApiClient::new(server.uri(), Arc::new(TokenStore))));
    let bytes = service
        .list_execution_clusters_connect(&request.encode_to_vec())
        .await
        .unwrap();
    let decoded = cluster::ListExecutionClustersResponse::decode(&*bytes).unwrap();

    assert_eq!(decoded.items[0].slug, "online");
}
