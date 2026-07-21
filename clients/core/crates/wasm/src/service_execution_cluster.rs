use std::sync::Arc;

use agentcloud_api_client::ApiClient;
use agentcloud_services::ExecutionClusterService;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmExecutionClusterService(ExecutionClusterService);

#[wasm_bindgen]
impl WasmExecutionClusterService {
    pub(crate) fn new(client: Arc<ApiClient>) -> Self {
        Self(ExecutionClusterService::new(client))
    }

    #[wasm_bindgen(js_name = listExecutionClustersConnect)]
    pub async fn list_execution_clusters_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_execution_clusters_connect(request).await
    }

    #[wasm_bindgen(js_name = createRegistrationCommandConnect)]
    pub async fn create_registration_command_connect(
        &self,
        request: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0.create_registration_command_connect(request).await
    }
}
