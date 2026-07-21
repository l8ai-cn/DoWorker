use std::sync::Arc;

use agentcloud_state::app_state::AppState;
use agentcloud_types::proto_goalloop_v1 as lp;
use parking_lot::RwLock;
use prost::Message;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmLoopBuilderState {
    state: Arc<RwLock<AppState>>,
}

fn decode_err<E: std::fmt::Display>(error: E) -> JsValue {
    JsValue::from_str(&format!("decode: {error}"))
}

impl WasmLoopBuilderState {
    pub(crate) fn from_runtime(state: Arc<RwLock<AppState>>) -> Self {
        Self { state }
    }
}

#[wasm_bindgen]
impl WasmLoopBuilderState {
    #[wasm_bindgen(constructor)]
    pub fn new() -> Self {
        Self {
            state: Arc::new(RwLock::new(AppState::with_storage(
                crate::new_memory_backend(),
            ))),
        }
    }

    pub fn snapshot_bytes(&self) -> Vec<u8> {
        self.state.read().loop_builder.snapshot().encode_to_vec()
    }

    pub fn set_source(&self, source: String, active_editor: String) {
        self.state
            .write()
            .loop_builder
            .set_source(source, active_editor);
    }

    pub fn set_active_editor(&self, active_editor: String) {
        self.state
            .write()
            .loop_builder
            .set_active_editor(active_editor);
    }

    pub fn apply_compile_response(&self, response: &[u8]) -> Result<(), JsValue> {
        let response = lp::CompileLoopProgramResponse::decode(response).map_err(decode_err)?;
        self.state.write().loop_builder.apply_compile(response);
        Ok(())
    }

    pub fn apply_ai_draft_response(&self, response: &[u8]) -> Result<bool, JsValue> {
        let response = lp::CompileLoopProgramResponse::decode(response).map_err(decode_err)?;
        Ok(self.state.write().loop_builder.apply_ai_draft(response))
    }

    pub fn apply_run_response(&self, response: &[u8]) -> Result<(), JsValue> {
        let run = lp::GoalLoop::decode(response).map_err(decode_err)?;
        self.state.write().loop_builder.apply_run(run);
        Ok(())
    }

    pub fn reset(&self) {
        self.state.write().loop_builder.reset();
    }
}

impl Default for WasmLoopBuilderState {
    fn default() -> Self {
        Self::new()
    }
}
