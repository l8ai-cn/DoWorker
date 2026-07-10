use std::sync::Arc;

use agentsmesh_api_client::ApiClient;
use agentsmesh_services::AIResourceService;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmAIResourceService(AIResourceService);

#[wasm_bindgen]
impl WasmAIResourceService {
    pub(crate) fn new(client: Arc<ApiClient>) -> Self {
        Self(AIResourceService::new(client))
    }

    #[wasm_bindgen(js_name = getCatalogConnect)]
    pub async fn get_catalog_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_catalog_connect(request).await
    }
    #[wasm_bindgen(js_name = listPersonalConnectionsConnect)]
    pub async fn list_personal_connections_connect(
        &self,
        request: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0.list_personal_connections_connect(request).await
    }
    #[wasm_bindgen(js_name = listOrganizationConnectionsConnect)]
    pub async fn list_organization_connections_connect(
        &self,
        request: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0.list_organization_connections_connect(request).await
    }
    #[wasm_bindgen(js_name = listPersonalEffectiveResourcesConnect)]
    pub async fn list_personal_effective_resources_connect(
        &self,
        request: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0
            .list_personal_effective_resources_connect(request)
            .await
    }
    #[wasm_bindgen(js_name = listOrganizationEffectiveResourcesConnect)]
    pub async fn list_organization_effective_resources_connect(
        &self,
        request: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0
            .list_organization_effective_resources_connect(request)
            .await
    }
    #[wasm_bindgen(js_name = createPersonalConnectionConnect)]
    pub async fn create_personal_connection_connect(
        &self,
        request: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0.create_personal_connection_connect(request).await
    }
    #[wasm_bindgen(js_name = createOrganizationConnectionConnect)]
    pub async fn create_organization_connection_connect(
        &self,
        request: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0.create_organization_connection_connect(request).await
    }
    #[wasm_bindgen(js_name = updateConnectionConnect)]
    pub async fn update_connection_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.update_connection_connect(request).await
    }
    #[wasm_bindgen(js_name = rotateConnectionCredentialsConnect)]
    pub async fn rotate_connection_credentials_connect(
        &self,
        request: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0.rotate_connection_credentials_connect(request).await
    }
    #[wasm_bindgen(js_name = setConnectionEnabledConnect)]
    pub async fn set_connection_enabled_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.set_connection_enabled_connect(request).await
    }
    #[wasm_bindgen(js_name = validateConnectionConnect)]
    pub async fn validate_connection_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.validate_connection_connect(request).await
    }
    #[wasm_bindgen(js_name = deleteConnectionConnect)]
    pub async fn delete_connection_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.delete_connection_connect(request).await
    }
    #[wasm_bindgen(js_name = createResourceConnect)]
    pub async fn create_resource_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.create_resource_connect(request).await
    }
    #[wasm_bindgen(js_name = updateResourceConnect)]
    pub async fn update_resource_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.update_resource_connect(request).await
    }
    #[wasm_bindgen(js_name = setResourceEnabledConnect)]
    pub async fn set_resource_enabled_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.set_resource_enabled_connect(request).await
    }
    #[wasm_bindgen(js_name = deleteResourceConnect)]
    pub async fn delete_resource_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.delete_resource_connect(request).await
    }
    #[wasm_bindgen(js_name = setDefaultConnect)]
    pub async fn set_default_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.set_default_connect(request).await
    }
}
