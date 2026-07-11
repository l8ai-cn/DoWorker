use std::sync::Arc;

use agentsmesh_state::app_state::AppState;
use agentsmesh_state::expert_types::{ExpertEnvelope, ExpertListResponse};
use parking_lot::RwLock;
use wasm_bindgen::prelude::*;

// Expert has no proto/Connect coverage (REST+JSON backend), so — unlike the
// repo/pod/runner state views — this bridge folds/serves JSON strings rather
// than prost bytes. Networking stays on the TS `expertApi` (lightFetch); this
// view owns the canonical client cache (Rust SSOT).
#[wasm_bindgen]
pub struct WasmExpertState {
    state: Arc<RwLock<AppState>>,
}

impl WasmExpertState {
    pub(crate) fn from_runtime(state: Arc<RwLock<AppState>>) -> Self {
        Self { state }
    }
}

#[wasm_bindgen]
impl WasmExpertState {
    pub fn experts_json(&self) -> String {
        serde_json::to_string(self.state.read().experts.experts()).unwrap_or_else(|_| "[]".into())
    }

    pub fn total(&self) -> f64 {
        self.state.read().experts.total() as f64
    }

    pub fn current_expert_json(&self) -> JsValue {
        match self.state.read().experts.current_expert() {
            Some(e) => JsValue::from_str(&serde_json::to_string(e).unwrap_or_default()),
            None => JsValue::NULL,
        }
    }

    /// Fold a `{experts, total}` list response into the cache.
    pub fn apply_fetched_experts(&self, resp_json: &str) -> Result<(), JsValue> {
        let resp: ExpertListResponse = serde_json::from_str(resp_json)
            .map_err(|e| JsValue::from_str(&format!("decode experts: {e}")))?;
        self.state
            .write()
            .experts
            .set_experts(resp.experts, resp.total);
        Ok(())
    }

    /// Fold a single `{expert}` response into `current_expert`.
    pub fn apply_fetched_expert(&self, resp_json: &str) -> Result<(), JsValue> {
        let resp: ExpertEnvelope = serde_json::from_str(resp_json)
            .map_err(|e| JsValue::from_str(&format!("decode expert: {e}")))?;
        self.state
            .write()
            .experts
            .set_current_expert(Some(resp.expert));
        Ok(())
    }

    pub fn clear_current_expert(&self) {
        self.state.write().experts.set_current_expert(None);
    }

    pub fn remove_expert(&self, slug: &str) {
        self.state.write().experts.remove_expert(slug);
    }
}
