use std::sync::Arc;

use agentsmesh_state::agent_workbench_state::{ProjectionStatus, ResyncReason};
use agentsmesh_state::app_state::AppState;
use parking_lot::RwLock;
use prost::Message;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmAgentWorkbenchState {
    state: Arc<RwLock<AppState>>,
}

impl WasmAgentWorkbenchState {
    pub(crate) fn from_runtime(state: Arc<RwLock<AppState>>) -> Self {
        Self { state }
    }
}

#[wasm_bindgen]
impl WasmAgentWorkbenchState {
    #[wasm_bindgen(js_name = snapshotBytes)]
    pub fn snapshot_bytes(&self, session_id: &str) -> Option<Vec<u8>> {
        self.state
            .read()
            .workbench
            .get_session(session_id)
            .map(|session| session.snapshot.encode_to_vec())
    }

    pub fn revision(&self, session_id: &str) -> Option<u64> {
        self.state.read().workbench.revision(session_id)
    }

    #[wasm_bindgen(js_name = projectionStatus)]
    pub fn projection_status(&self, session_id: &str) -> Option<String> {
        let state = self.state.read();
        let session = state.workbench.get_session(session_id)?;
        Some(
            match session.status {
                ProjectionStatus::Ready => "ready",
                ProjectionStatus::ResyncRequired => "resync_required",
            }
            .into(),
        )
    }

    #[wasm_bindgen(js_name = resyncReason)]
    pub fn resync_reason(&self, session_id: &str) -> Option<String> {
        let state = self.state.read();
        let reason = state.workbench.get_session(session_id)?.resync_reason?;
        Some(
            match reason {
                ResyncReason::StreamEpochChanged => "stream_epoch_changed",
                ResyncReason::BaseRevisionMismatch => "base_revision_mismatch",
                ResyncReason::SequenceGap => "sequence_gap",
                ResyncReason::DigestConflict => "digest_conflict",
            }
            .into(),
        )
    }
}
