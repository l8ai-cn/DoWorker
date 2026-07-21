use std::sync::Arc;

use agentcloud_api_client::ApiClient;
use agentcloud_services::MeshService;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmMeshService(pub(crate) MeshService);

#[wasm_bindgen]
impl WasmMeshService {
    pub(crate) fn new(client: Arc<ApiClient>) -> Self {
        Self(MeshService::new(client))
    }

    /// Networking-only: returns a prost-encoded `ReplaceTopologyRequest` the
    /// caller feeds to the mesh state surface (getMeshState().replace_topology).
    pub async fn fetch_topology(&self) -> Result<Vec<u8>, String> {
        self.0.fetch_topology().await
    }
}
