use std::sync::Arc;

use agentsmesh_api_client::ApiClient;
use agentsmesh_services::OrchestrationResourceService;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmOrchestrationResourceService(OrchestrationResourceService);

#[wasm_bindgen]
impl WasmOrchestrationResourceService {
    pub(crate) fn new(client: Arc<ApiClient>) -> Self {
        Self(OrchestrationResourceService::new(client))
    }

    #[wasm_bindgen(js_name = validateResourceConnect)]
    pub async fn validate_resource_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.validate_resource_connect(request).await
    }

    #[wasm_bindgen(js_name = planResourceConnect)]
    pub async fn plan_resource_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.plan_resource_connect(request).await
    }

    #[wasm_bindgen(js_name = getResourceConnect)]
    pub async fn get_resource_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_resource_connect(request).await
    }

    #[wasm_bindgen(js_name = listResourcesConnect)]
    pub async fn list_resources_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_resources_connect(request).await
    }

    #[wasm_bindgen(js_name = exportResourceConnect)]
    pub async fn export_resource_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.export_resource_connect(request).await
    }

    #[wasm_bindgen(js_name = getResourcePlanConnect)]
    pub async fn get_resource_plan_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_resource_plan_connect(request).await
    }
}
