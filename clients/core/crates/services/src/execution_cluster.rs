use std::sync::Arc;

use agentcloud_api_client::ApiClient;
use agentcloud_types::proto_execution_cluster_v1 as cluster;
use prost::Message;

pub struct ExecutionClusterService {
    client: Arc<ApiClient>,
}

impl ExecutionClusterService {
    pub fn new(client: Arc<ApiClient>) -> Self {
        Self { client }
    }

    pub async fn list_execution_clusters_connect(&self, bytes: &[u8]) -> Result<Vec<u8>, String> {
        let request = cluster::ListExecutionClustersRequest::decode(bytes)
            .map_err(|error| format!("decode list_execution_clusters request: {error}"))?;
        let response = self
            .client
            .list_execution_clusters_connect(&request)
            .await
            .map_err(crate::wire)?;
        Ok(response.encode_to_vec())
    }

    pub async fn create_registration_command_connect(
        &self,
        bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        let request = cluster::CreateRegistrationCommandRequest::decode(bytes)
            .map_err(|error| format!("decode create_registration_command request: {error}"))?;
        let response = self
            .client
            .create_execution_cluster_registration_command_connect(&request)
            .await
            .map_err(crate::wire)?;
        Ok(response.encode_to_vec())
    }
}
