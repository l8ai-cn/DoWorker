use std::sync::Arc;

use agentsmesh_state::app_state::AppState;
use agentsmesh_state::loop_state::LoopRunData;
use agentsmesh_types::proto_loop_state_v1::{
    ClearCurrentLoopRequest, ClearLoopRunsRequest, InsertLoopRunRequest,
    PatchLoopFromActionRequest, PatchLoopRunStatusRequest, ReplaceCachedLoopsRequest,
    ReplaceCachedRunsRequest, SetCurrentLoopRequest,
};
use agentsmesh_types::proto_loop_v1::{ListLoopsResponse, ListRunsResponse, Loop as ProtoLoop};
use parking_lot::RwLock;
use prost::Message;
use wasm_bindgen::prelude::*;

use crate::state_loop_proto::{loop_from_proto, loop_to_proto, run_from_proto, run_to_proto};

#[wasm_bindgen]
pub struct WasmLoopState {
    state: Arc<RwLock<AppState>>,
}

fn decode_err<E: std::fmt::Display>(e: E) -> JsValue {
    JsValue::from_str(&format!("decode: {e}"))
}

impl WasmLoopState {
    pub(crate) fn from_runtime(state: Arc<RwLock<AppState>>) -> Self {
        Self { state }
    }
}

#[wasm_bindgen]
impl WasmLoopState {
    #[wasm_bindgen(constructor)]
    pub fn new() -> Self {
        Self {
            state: Arc::new(RwLock::new(AppState::with_storage(
                crate::new_memory_backend(),
            ))),
        }
    }

    pub fn loops_json(&self) -> String {
        serde_json::to_string(self.state.read().loops.get_loops()).unwrap_or_default()
    }

    pub fn current_loop_json(&self) -> JsValue {
        match self.state.read().loops.get_current_loop() {
            Some(l) => JsValue::from_str(&serde_json::to_string(l).unwrap_or_default()),
            None => JsValue::NULL,
        }
    }

    pub fn runs_json(&self) -> String {
        serde_json::to_string(self.state.read().loops.get_runs()).unwrap_or_default()
    }

    pub fn get_loop_by_slug_json(&self, slug: &str) -> JsValue {
        match self.state.read().loops.get_loop_by_slug(slug) {
            Some(l) => JsValue::from_str(&serde_json::to_string(l).unwrap_or_default()),
            None => JsValue::NULL,
        }
    }

    // Read side (B, zero-JSON): prost-encode state into the same wrappers the
    // mutators decode, so the shared selectors decode bytes uniformly.
    pub fn loops_bytes(&self) -> Vec<u8> {
        let loops = self
            .state
            .read()
            .loops
            .get_loops()
            .iter()
            .map(loop_to_proto)
            .collect();
        ReplaceCachedLoopsRequest { loops }.encode_to_vec()
    }

    pub fn runs_bytes(&self) -> Vec<u8> {
        let runs = self
            .state
            .read()
            .loops
            .get_runs()
            .iter()
            .map(run_to_proto)
            .collect();
        ReplaceCachedRunsRequest { runs }.encode_to_vec()
    }

    pub fn current_loop_bytes(&self) -> Vec<u8> {
        match self.state.read().loops.get_current_loop() {
            Some(l) => SetCurrentLoopRequest {
                r#loop: Some(loop_to_proto(l)),
            }
            .encode_to_vec(),
            None => Vec::new(),
        }
    }

    // Fetch→state (B): decode wire ListLoops/ListRuns response + fold into state
    // via loop_from_proto/run_from_proto — no TS loopToProtoLoop/fromProtoLoop.
    pub fn apply_fetched_loops(&self, resp_bytes: &[u8]) -> Result<(), JsValue> {
        let resp = ListLoopsResponse::decode(resp_bytes).map_err(decode_err)?;
        let loops = resp.items.into_iter().map(loop_from_proto).collect();
        self.state.write().loops.set_loops(loops);
        Ok(())
    }

    pub fn apply_fetched_runs(&self, resp_bytes: &[u8]) -> Result<(), JsValue> {
        let resp = ListRunsResponse::decode(resp_bytes).map_err(decode_err)?;
        let runs = resp.items.into_iter().map(run_from_proto).collect();
        self.state.write().loops.set_runs(runs);
        Ok(())
    }

    pub fn apply_appended_runs(&self, resp_bytes: &[u8]) -> Result<(), JsValue> {
        let resp = ListRunsResponse::decode(resp_bytes).map_err(decode_err)?;
        let runs: Vec<LoopRunData> = resp.items.into_iter().map(run_from_proto).collect();
        self.state.write().loops.append_runs(runs);
        Ok(())
    }

    pub fn set_current_loop(&self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let req = SetCurrentLoopRequest::decode(req_bytes).map_err(decode_err)?;
        let loop_data = req.r#loop.map(loop_from_proto);
        self.state.write().loops.set_current_loop(loop_data);
        Ok(())
    }

    // Fetch→state (B): decode the full wire GetLoop response (Loop) + fold via
    // loop_from_proto — no TS loopToProtoLoop round-trip (which dropped the
    // proto-only fields the lossy LoopData cannot carry).
    pub fn apply_fetched_current_loop(&self, resp_bytes: &[u8]) -> Result<(), JsValue> {
        let proto = ProtoLoop::decode(resp_bytes).map_err(decode_err)?;
        self.state
            .write()
            .loops
            .set_current_loop(Some(loop_from_proto(proto)));
        Ok(())
    }

    pub fn clear_current_loop(&self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let _ = ClearCurrentLoopRequest::decode(req_bytes).map_err(decode_err)?;
        self.state.write().loops.set_current_loop(None);
        Ok(())
    }

    pub fn patch_loop_from_action(&self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let req = PatchLoopFromActionRequest::decode(req_bytes).map_err(decode_err)?;
        let loop_data = req
            .r#loop
            .ok_or_else(|| JsValue::from_str("missing loop"))?;
        self.state
            .write()
            .loops
            .update_loop(&req.slug, loop_from_proto(loop_data));
        Ok(())
    }

    pub fn insert_loop_run(&self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let req = InsertLoopRunRequest::decode(req_bytes).map_err(decode_err)?;
        let run = req.run.ok_or_else(|| JsValue::from_str("missing run"))?;
        self.state.write().loops.add_run(run_from_proto(run));
        Ok(())
    }

    pub fn patch_loop_run_status(&self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let req = PatchLoopRunStatusRequest::decode(req_bytes).map_err(decode_err)?;
        self.state
            .write()
            .loops
            .update_run_status(req.run_id, &req.status);
        Ok(())
    }

    pub fn clear_loop_runs(&self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let _ = ClearLoopRunsRequest::decode(req_bytes).map_err(decode_err)?;
        self.state.write().loops.clear_runs();
        Ok(())
    }
}
