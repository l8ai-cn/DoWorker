use std::sync::Arc;

use agentcloud_api_client::ApiClient;
use agentcloud_services::PodService;
use wasm_bindgen::prelude::*;

// Networking-only wasm handle for the pod domain. The pod cache lives in the
// shared `AppState.pods` (reached via `WasmPodState`); this service exposes
// only the Connect-RPC `*_connect` surface.
#[wasm_bindgen]
pub struct WasmPodService(pub(crate) PodService);

#[wasm_bindgen]
impl WasmPodService {
    pub(crate) fn new(client: Arc<ApiClient>) -> Self {
        Self(PodService::new(client))
    }

    // -------- Connect-RPC (binary wire) --------
    //
    // Each `*_connect` method takes prost-encoded bytes (Uint8Array on the JS
    // side) and returns prost-encoded bytes — TS callers encode via
    // @bufbuild/protobuf .toBinary() and decode via .fromBinary().

    pub async fn list_pods_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_pods_connect(request_bytes).await
    }

    pub async fn get_pod_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_pod_connect(request_bytes).await
    }

    pub async fn create_pod_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.create_pod_connect(request_bytes).await
    }

    pub async fn terminate_pod_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.terminate_pod_connect(request_bytes).await
    }

    pub async fn update_pod_alias_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.update_pod_alias_connect(request_bytes).await
    }

    pub async fn update_pod_perpetual_connect(
        &self,
        request_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0.update_pod_perpetual_connect(request_bytes).await
    }

    pub async fn update_pod_preview_config_connect(
        &self,
        request_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0
            .update_pod_preview_config_connect(request_bytes)
            .await
    }

    pub async fn get_mobile_access_descriptor_connect(
        &self,
        request_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0
            .get_mobile_access_descriptor_connect(request_bytes)
            .await
    }

    pub async fn get_pod_connection_connect(
        &self,
        request_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0.get_pod_connection_connect(request_bytes).await
    }

    pub async fn send_pod_prompt_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.send_pod_prompt_connect(request_bytes).await
    }

    pub async fn list_pods_by_ticket_connect(
        &self,
        request_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0.list_pods_by_ticket_connect(request_bytes).await
    }
}
