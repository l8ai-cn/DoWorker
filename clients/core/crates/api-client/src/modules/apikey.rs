use crate::ApiClient;
use crate::error::ApiError;
use agentsmesh_types::*;

impl ApiClient {
    pub async fn list_api_keys(&self) -> Result<ApiKeyListResponse, ApiError> {
        self.get(&self.org_path("/api-keys")).await
    }

    pub async fn get_api_key(&self, id: i64) -> Result<ApiKey, ApiError> {
        self.get_resource(&self.org_path(&format!("/api-keys/{id}")), "api_key").await
    }

    pub async fn create_api_key(
        &self,
        data: &CreateApiKeyRequest,
    ) -> Result<CreateApiKeyResponse, ApiError> {
        self.post(&self.org_path("/api-keys"), data).await
    }

    pub async fn update_api_key(
        &self,
        id: i64,
        data: &UpdateApiKeyRequest,
    ) -> Result<ApiKey, ApiError> {
        self.put_resource(&self.org_path(&format!("/api-keys/{id}")), data, "api_key").await
    }

    pub async fn delete_api_key(&self, id: i64) -> Result<EmptyResponse, ApiError> {
        self.delete(&self.org_path(&format!("/api-keys/{id}")))
            .await
    }

    pub async fn revoke_api_key(&self, id: i64) -> Result<EmptyResponse, ApiError> {
        self.post(
            &self.org_path(&format!("/api-keys/{id}/revoke")),
            &serde_json::json!({}),
        )
        .await
    }
}
