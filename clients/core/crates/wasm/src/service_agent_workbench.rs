use std::sync::Arc;

use agentsmesh_api_client::{AgentWorkbenchAccessScope, ApiClient};
use agentsmesh_state::app_state::AppState;
use agentsmesh_types::proto_agent_workbench_v2 as v2;
use parking_lot::RwLock;
use prost::Message;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmAgentWorkbenchService {
    pub(crate) client: Arc<ApiClient>,
    pub(crate) state: Arc<RwLock<AppState>>,
}

impl WasmAgentWorkbenchService {
    pub(crate) fn new(client: Arc<ApiClient>, state: Arc<RwLock<AppState>>) -> Self {
        Self { client, state }
    }
}

#[wasm_bindgen]
impl WasmAgentWorkbenchService {
    #[wasm_bindgen(js_name = getSessionSnapshotConnect)]
    pub async fn get_session_snapshot_connect(
        &self,
        org_slug: String,
        bearer_token: String,
        session_id: String,
    ) -> Result<Vec<u8>, String> {
        let access = access_scope(org_slug, bearer_token)?;
        let snapshot = self
            .client
            .get_agent_workbench_session_snapshot_connect(&access, &session_id)
            .await
            .map_err(|error| error.to_wire())?;
        self.state
            .write()
            .workbench
            .apply_snapshot(&snapshot)
            .map_err(|error| error.to_string())?;
        Ok(snapshot.encode_to_vec())
    }

    #[cfg(target_arch = "wasm32")]
    #[wasm_bindgen(js_name = streamSessionDeltasConnect)]
    pub async fn stream_session_deltas_connect(
        &self,
        org_slug: String,
        bearer_token: String,
        stream_base_url: String,
        session_id: String,
        replay_limit: u32,
        on_commit: js_sys::Function,
        on_error: js_sys::Function,
        on_close: js_sys::Function,
    ) -> Result<crate::agent_workbench_stream_wasm::WasmAgentWorkbenchStream, String> {
        let access = access_scope(org_slug, bearer_token)?;
        if stream_base_url.trim().is_empty() {
            return Err("agent workbench stream base URL is required".into());
        }
        let stream_client = Arc::new(self.client.clone_with_base_url(stream_base_url));
        crate::agent_workbench_stream_wasm::open_agent_workbench_stream(
            stream_client,
            self.state.clone(),
            access,
            session_id,
            replay_limit,
            on_commit,
            on_error,
            on_close,
        )
        .await
    }

    #[wasm_bindgen(js_name = executeCommandConnect)]
    pub async fn execute_command_connect(
        &self,
        org_slug: String,
        bearer_token: String,
        command_envelope_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        let access = access_scope(org_slug, bearer_token)?;
        let command = v2::CommandEnvelope::decode(command_envelope_bytes)
            .map_err(|error| format!("decode command envelope: {error}"))?;
        self.client
            .execute_agent_workbench_command_connect(&access, command)
            .await
            .map(|receipt| receipt.encode_to_vec())
            .map_err(|error| error.to_wire())
    }
}

fn access_scope(
    org_slug: String,
    bearer_token: String,
) -> Result<AgentWorkbenchAccessScope, String> {
    AgentWorkbenchAccessScope::new(org_slug, bearer_token).map_err(|error| error.to_wire())
}

pub(crate) fn session_cursor(
    state: &Arc<RwLock<AppState>>,
    session_id: &str,
) -> Result<v2::SessionCursor, String> {
    let state = state.read();
    let session = state
        .workbench
        .get_session(session_id)
        .ok_or_else(|| format!("agent workbench snapshot missing: {session_id}"))?;
    Ok(v2::SessionCursor {
        session_id: session.snapshot.session_id.clone(),
        stream_epoch: session.snapshot.stream_epoch.clone(),
        revision: session.snapshot.revision,
        sequence: session.snapshot.latest_sequence,
    })
}

pub(crate) fn apply_stream_batch(
    state: &Arc<RwLock<AppState>>,
    batch: &v2::SessionDeltaBatch,
) -> Result<bool, String> {
    let mut state = state.write();
    let before = state.workbench.revision(&batch.session_id);
    state
        .workbench
        .apply_delta_batch(batch)
        .map_err(|error| error.to_string())?;
    Ok(state.workbench.revision(&batch.session_id) != before)
}
