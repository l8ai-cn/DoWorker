use wasm_bindgen::prelude::*;

use crate::WasmPodService;

#[wasm_bindgen]
impl WasmPodService {
    pub async fn list_worker_create_options_connect(
        &self,
        request_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0
            .list_worker_create_options_connect(request_bytes)
            .await
    }

    pub async fn preflight_worker_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.preflight_worker_connect(request_bytes).await
    }

    pub async fn fill_worker_draft_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.fill_worker_draft_connect(request_bytes).await
    }
}
