use std::sync::Arc;

use agentcloud_api_client::ApiClient;
use agentcloud_services::AutopilotService;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmAutopilotService(pub(crate) AutopilotService);

#[wasm_bindgen]
impl WasmAutopilotService {
    pub(crate) fn new(client: Arc<ApiClient>) -> Self {
        Self(AutopilotService::new(client))
    }

    // -------- Connect-RPC (binary wire) --------

    #[wasm_bindgen(js_name = listAutopilotsConnect)]
    pub async fn list_autopilots_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_autopilots_connect(request).await
    }

    #[wasm_bindgen(js_name = getAutopilotConnect)]
    pub async fn get_autopilot_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_autopilot_connect(request).await
    }

    #[wasm_bindgen(js_name = createAutopilotConnect)]
    pub async fn create_autopilot_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.create_autopilot_connect(request).await
    }

    #[wasm_bindgen(js_name = pauseAutopilotConnect)]
    pub async fn pause_autopilot_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.action_autopilot_connect("pause", request).await
    }

    #[wasm_bindgen(js_name = resumeAutopilotConnect)]
    pub async fn resume_autopilot_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.action_autopilot_connect("resume", request).await
    }

    #[wasm_bindgen(js_name = stopAutopilotConnect)]
    pub async fn stop_autopilot_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.action_autopilot_connect("stop", request).await
    }

    #[wasm_bindgen(js_name = takeoverAutopilotConnect)]
    pub async fn takeover_autopilot_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.action_autopilot_connect("takeover", request).await
    }

    #[wasm_bindgen(js_name = handbackAutopilotConnect)]
    pub async fn handback_autopilot_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.action_autopilot_connect("handback", request).await
    }

    #[wasm_bindgen(js_name = approveAutopilotConnect)]
    pub async fn approve_autopilot_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.approve_autopilot_connect(request).await
    }

    #[wasm_bindgen(js_name = getIterationsConnect)]
    pub async fn get_iterations_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_iterations_connect(request).await
    }
}
