use std::sync::Arc;

use agentcloud_api_client::ApiClient;
use agentcloud_services::WorkflowService;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmWorkflowService(pub(crate) WorkflowService);

#[wasm_bindgen]
impl WasmWorkflowService {
    pub(crate) fn new(client: Arc<ApiClient>) -> Self {
        Self(WorkflowService::new(client))
    }

    // -------- Connect-RPC (binary wire) --------

    #[wasm_bindgen(js_name = listWorkflowsConnect)]
    pub async fn list_workflows_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_workflows_connect(request).await
    }

    #[wasm_bindgen(js_name = getWorkflowConnect)]
    pub async fn get_workflow_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_workflow_connect(request).await
    }

    #[wasm_bindgen(js_name = createWorkflowConnect)]
    pub async fn create_workflow_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.create_workflow_connect(request).await
    }

    #[wasm_bindgen(js_name = updateWorkflowConnect)]
    pub async fn update_workflow_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.update_workflow_connect(request).await
    }

    #[wasm_bindgen(js_name = deleteWorkflowConnect)]
    pub async fn delete_workflow_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.delete_workflow_connect(request).await
    }

    #[wasm_bindgen(js_name = enableWorkflowConnect)]
    pub async fn enable_workflow_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.workflow_action_connect("enable", request).await
    }

    #[wasm_bindgen(js_name = disableWorkflowConnect)]
    pub async fn disable_workflow_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.workflow_action_connect("disable", request).await
    }

    #[wasm_bindgen(js_name = triggerWorkflowConnect)]
    pub async fn trigger_workflow_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.trigger_workflow_connect(request).await
    }

    #[wasm_bindgen(js_name = listWorkflowRunsConnect)]
    pub async fn list_workflow_runs_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_workflow_runs_connect(request).await
    }

    #[wasm_bindgen(js_name = cancelWorkflowRunConnect)]
    pub async fn cancel_workflow_run_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.cancel_workflow_run_connect(request).await
    }
}
