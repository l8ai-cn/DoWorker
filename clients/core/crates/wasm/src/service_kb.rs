use std::sync::Arc;

use agentsmesh_api_client::ApiClient;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmKnowledgeBaseService {
    client: Arc<ApiClient>,
}

impl WasmKnowledgeBaseService {
    /// Crate-local accessor used by service_kb_connect.rs to forward to the
    /// underlying api-client `*_connect` methods.
    pub(crate) fn client_ref(&self) -> &ApiClient {
        &self.client
    }
}

#[wasm_bindgen]
impl WasmKnowledgeBaseService {
    pub(crate) fn new(client: Arc<ApiClient>) -> Self {
        Self { client }
    }
}
