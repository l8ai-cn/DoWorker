use crate::{connect_call, ApiClient, ApiError};
use agentcloud_types::proto_execution_cluster_v1 as cluster;

impl ApiClient {
    pub async fn list_execution_clusters_connect(
        &self,
        request: &cluster::ListExecutionClustersRequest,
    ) -> Result<cluster::ListExecutionClustersResponse, ApiError> {
        connect_call(
            self,
            "/proto.execution_cluster.v1.ExecutionClusterService/ListExecutionClusters",
            request,
        )
        .await
    }

    pub async fn create_execution_cluster_registration_command_connect(
        &self,
        request: &cluster::CreateRegistrationCommandRequest,
    ) -> Result<cluster::CreateRegistrationCommandResponse, ApiError> {
        connect_call(
            self,
            "/proto.execution_cluster.v1.ExecutionClusterService/CreateRegistrationCommand",
            request,
        )
        .await
    }
}
