use agentsmesh_state::app_state::AppState;
use agentsmesh_types::proto_app_state_v1::DispatchEventRequest;
use prost::Message;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmAppState {
    inner: AppState,
}

fn decode_err<E: std::fmt::Display>(e: E) -> JsValue {
    JsValue::from_str(&format!("decode: {e}"))
}

#[wasm_bindgen]
impl WasmAppState {
    #[wasm_bindgen(constructor)]
    pub fn new() -> Self {
        Self {
            inner: AppState::with_storage(crate::new_memory_backend()),
        }
    }

    pub fn dispatch_event(&mut self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let req = DispatchEventRequest::decode(req_bytes).map_err(decode_err)?;
        if let Ok(event) = serde_json::from_str(&req.event_json) {
            self.inner.dispatch(&event);
        }
        Ok(())
    }

    pub fn pods_json(&self) -> String {
        serde_json::to_string(self.inner.pods.pods()).unwrap_or_default()
    }

    pub fn channels_json(&self) -> String {
        serde_json::to_string(self.inner.channels.get_channels()).unwrap_or_default()
    }

    pub fn runners_json(&self) -> String {
        serde_json::to_string(self.inner.runners.runners()).unwrap_or_default()
    }

    pub fn tickets_json(&self) -> String {
        serde_json::to_string(self.inner.tickets.get_tickets()).unwrap_or_default()
    }

    pub fn workflows_json(&self) -> String {
        serde_json::to_string(self.inner.workflows.get_workflows()).unwrap_or_default()
    }

    pub fn mesh_json(&self) -> String {
        match self.inner.mesh.topology() {
            Some(t) => serde_json::to_string(t).unwrap_or_default(),
            None => "null".to_string(),
        }
    }
}
