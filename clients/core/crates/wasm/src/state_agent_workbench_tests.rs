use std::sync::Arc;

use agentcloud_state::app_state::AppState;
use agentcloud_types::proto_agent_workbench_v2 as v2;
use parking_lot::RwLock;
use prost::Message;

use crate::api::WasmApiClient;
use crate::service_agent_workbench::{
    apply_stream_batch, session_cursor, WasmAgentWorkbenchService,
};
use crate::state_agent_workbench::WasmAgentWorkbenchState;

#[test]
fn state_reads_canonical_snapshot_bytes_and_revision() {
    let state = Arc::new(RwLock::new(AppState::new()));
    let view = WasmAgentWorkbenchState::from_runtime(state.clone());
    assert_eq!(view.snapshot_bytes("missing"), None);
    assert_eq!(view.revision("missing"), None);
    assert_eq!(view.projection_status("missing"), None);
    assert_eq!(view.resync_reason("missing"), None);

    let snapshot = snapshot();
    state.write().workbench.apply_snapshot(&snapshot).unwrap();

    let encoded = view.snapshot_bytes("session-1").unwrap();
    assert_eq!(
        v2::SessionSnapshot::decode(encoded.as_slice()).unwrap(),
        snapshot
    );
    assert_eq!(view.revision("session-1"), Some(1));
    assert_eq!(
        view.projection_status("session-1").as_deref(),
        Some("ready")
    );
    assert_eq!(view.resync_reason("session-1"), None);
}

#[test]
fn stream_cursor_comes_from_canonical_snapshot() {
    let state = loaded_state();
    assert_eq!(
        session_cursor(&state, "session-1").unwrap(),
        v2::SessionCursor {
            session_id: "session-1".into(),
            stream_epoch: "epoch-1".into(),
            revision: 4,
            sequence: 9,
        }
    );
    assert!(session_cursor(&state, "missing").is_err());
}

#[test]
fn stream_batch_reports_only_commits_and_surfaces_gaps() {
    let state = loaded_state();
    let view = WasmAgentWorkbenchState::from_runtime(state.clone());
    let batch = status_batch(4, 5, 10);
    assert_eq!(apply_stream_batch(&state, &batch).unwrap(), true);
    assert_eq!(apply_stream_batch(&state, &batch).unwrap(), false);

    let gap = status_batch(5, 6, 12);
    let error = apply_stream_batch(&state, &gap).unwrap_err();
    assert!(error.contains("SequenceGap"));
    assert_eq!(state.read().workbench.revision("session-1"), Some(2));
    assert_eq!(
        view.projection_status("session-1").as_deref(),
        Some("resync_required")
    );
    assert_eq!(
        view.resync_reason("session-1").as_deref(),
        Some("sequence_gap")
    );
}

#[test]
fn api_exposes_agent_workbench_factory_state_and_service_methods() {
    let _ = WasmApiClient::create_agent_workbench_service;
    let _ = WasmApiClient::get_agent_workbench_state;
    let _ = WasmAgentWorkbenchService::get_session_snapshot_connect;
    let _ = WasmAgentWorkbenchService::execute_command_connect;
}

fn loaded_state() -> Arc<RwLock<AppState>> {
    let state = Arc::new(RwLock::new(AppState::new()));
    state.write().workbench.apply_snapshot(&snapshot()).unwrap();
    state
}

fn snapshot() -> v2::SessionSnapshot {
    v2::SessionSnapshot {
        session_id: "session-1".into(),
        stream_epoch: "epoch-1".into(),
        revision: 4,
        latest_sequence: 9,
        status: v2::SessionStatus::Running as i32,
        ..Default::default()
    }
}

fn status_batch(base_revision: u64, revision: u64, sequence: u64) -> v2::SessionDeltaBatch {
    v2::SessionDeltaBatch {
        session_id: "session-1".into(),
        stream_epoch: "epoch-1".into(),
        base_revision,
        revision,
        first_sequence: sequence,
        last_sequence: sequence,
        events: vec![v2::AgentEvent {
            envelope: Some(v2::EventEnvelope {
                session_id: "session-1".into(),
                stream_epoch: "epoch-1".into(),
                revision,
                sequence,
                item_id: format!("status-{revision}"),
                created_at: "2026-07-16T00:00:00Z".into(),
                ..Default::default()
            }),
            event: Some(v2::agent_event::Event::SessionStatusChanged(
                v2::SessionStatusChanged {
                    status: v2::SessionStatus::Running as i32,
                    error: None,
                },
            )),
        }],
        digest: format!("batch-{revision}"),
    }
}
