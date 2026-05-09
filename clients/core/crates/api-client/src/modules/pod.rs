use crate::ApiClient;
use crate::error::ApiError;
use agentsmesh_types::*;

impl ApiClient {
    pub async fn list_pods(
        &self,
        status: Option<&str>,
        runner_id: Option<i64>,
        created_by_id: Option<i64>,
        limit: Option<i64>,
        offset: Option<i64>,
    ) -> Result<PodListResponse, ApiError> {
        let mut path = self.org_path("/pods");
        let mut params = Vec::new();
        if let Some(v) = status {
            params.push(format!("status={v}"));
        }
        if let Some(v) = runner_id {
            params.push(format!("runner_id={v}"));
        }
        if let Some(v) = created_by_id {
            params.push(format!("created_by_id={v}"));
        }
        if let Some(v) = limit {
            params.push(format!("limit={v}"));
        }
        if let Some(v) = offset {
            params.push(format!("offset={v}"));
        }
        if !params.is_empty() {
            path = format!("{path}?{}", params.join("&"));
        }
        self.get(&path).await
    }

    pub async fn get_pod(&self, key: &str) -> Result<Pod, ApiError> {
        self.get_resource(&self.org_path(&format!("/pods/{key}")), "pod").await
    }

    pub async fn create_pod(&self, data: &CreatePodRequest) -> Result<CreatePodResponse, ApiError> {
        self.post(&self.org_path("/pods"), data).await
    }

    pub async fn terminate_pod(&self, key: &str) -> Result<EmptyResponse, ApiError> {
        self.post(
            &self.org_path(&format!("/pods/{key}/terminate")),
            &serde_json::json!({}),
        )
        .await
    }

    pub async fn get_pod_connection_info(
        &self,
        key: &str,
    ) -> Result<PodConnectionInfo, ApiError> {
        self.get(&self.org_path(&format!("/pods/{key}/connect")))
            .await
    }

    pub async fn get_pod_relay_connection(
        &self,
        key: &str,
    ) -> Result<PodConnectionInfo, ApiError> {
        self.get(&self.org_path(&format!("/pods/{key}/relay/connect")))
            .await
    }

    pub async fn update_pod_alias(
        &self,
        key: &str,
        data: &UpdatePodAliasRequest,
    ) -> Result<Pod, ApiError> {
        self.patch_resource(&self.org_path(&format!("/pods/{key}/alias")), data, "pod").await
    }
}
