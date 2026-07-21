use agentcloud_types::proto_agent_workbench_v2 as v2;

use crate::agent_workbench_state::{
    AgentWorkbenchError, AgentWorkbenchSession, AgentWorkbenchState, ProjectionStatus, ResyncReason,
};
use crate::agent_workbench_test_fixtures::{
    batch, event, receipt, receipt_event, snapshot, timeline_event, timeline_item,
};
use crate::app_state::AppState;

#[test]
fn two_event_batch_commits_once() {
    let mut state = loaded_state();
    let events = vec![
        timeline_event(10, "message-1", 5, false),
        receipt_event(11, 5, "command-1", v2::CommandReceiptState::Running),
    ];
    state
        .apply_delta_batch(&batch(4, 5, 10, events, "batch-1"))
        .unwrap();

    let session = session(&state);
    assert_eq!(state.revision("session-1"), Some(2));
    assert_eq!(state.revision("missing"), None);
    assert_eq!(session.snapshot.revision, 5);
    assert_eq!(session.snapshot.latest_sequence, 11);
    assert_eq!(session.snapshot.history.len(), 1);
    assert_eq!(
        session.snapshot.command_receipts[0].state,
        v2::CommandReceiptState::Running as i32
    );
}

#[test]
fn duplicate_digest_is_idempotent_but_conflict_fails() {
    let mut state = loaded_state();
    let original = batch(
        4,
        5,
        10,
        vec![timeline_event(10, "message-1", 5, false)],
        "batch-1",
    );
    state.apply_delta_batch(&original).unwrap();
    state.apply_delta_batch(&original).unwrap();
    assert_eq!(session(&state).commit_revision, 2);

    let mut conflict = original.clone();
    conflict.digest = "batch-conflict".into();
    assert!(matches!(
        state.apply_delta_batch(&conflict),
        Err(AgentWorkbenchError::DigestConflict { .. })
    ));
    let session = session(&state);
    assert_eq!(session.commit_revision, 2);
    assert_eq!(session.status, ProjectionStatus::ResyncRequired);
    assert_eq!(session.resync_reason, Some(ResyncReason::DigestConflict));
}

#[test]
fn continuity_failures_preserve_snapshot_and_require_resync() {
    let mut epoch = batch(4, 5, 10, vec![status_event(10, 5)], "epoch");
    epoch.stream_epoch = "epoch-2".into();
    epoch.events[0].envelope.as_mut().unwrap().stream_epoch = "epoch-2".into();
    let cases = [
        (
            batch(4, 5, 11, vec![status_event(11, 5)], "gap"),
            ResyncReason::SequenceGap,
        ),
        (epoch, ResyncReason::StreamEpochChanged),
        (
            batch(3, 4, 10, vec![status_event(10, 4)], "stale"),
            ResyncReason::BaseRevisionMismatch,
        ),
    ];

    for (delta, expected) in cases {
        let mut state = loaded_state();
        let before = session(&state).snapshot.clone();
        assert!(matches!(
            state.apply_delta_batch(&delta),
            Err(AgentWorkbenchError::ResyncRequired { reason, .. }) if reason == expected
        ));
        let session = session(&state);
        assert_eq!(session.snapshot, before);
        assert_eq!(session.commit_revision, 1);
        assert_eq!(session.status, ProjectionStatus::ResyncRequired);
        assert_eq!(session.resync_reason, Some(expected));
    }
}

#[test]
fn snapshot_validation_matches_v2_contract_and_is_atomic() {
    let mut state = loaded_state();
    let original = session(&state).snapshot.clone();
    for value in invalid_snapshots() {
        assert!(matches!(
            state.apply_snapshot(&value),
            Err(AgentWorkbenchError::InvalidPayload { .. })
        ));
        let session = session(&state);
        assert_eq!(session.snapshot, original);
        assert_eq!(session.commit_revision, 1);
    }
}

#[test]
fn snapshots_reject_stale_cursors_and_earliest_epoch_after_33_rotations() {
    let mut state = AgentWorkbenchState::new();
    let initial = snapshot();
    state.apply_snapshot(&initial).unwrap();
    state.apply_snapshot(&initial).unwrap();
    assert_eq!(state.revision("session-1"), Some(1));

    let mut cursor_conflict = initial.clone();
    cursor_conflict.status = v2::SessionStatus::Idle as i32;
    assert_snapshot_error(&mut state, cursor_conflict, "snapshot_cursor_conflict");

    let mut stale = initial.clone();
    stale.revision -= 1;
    assert_snapshot_error(&mut state, stale, "snapshot_stale");
    assert_eq!(session(&state).snapshot, initial);

    let next_epoch = epoch_snapshot("epoch-2");
    state.apply_snapshot(&next_epoch).unwrap();
    assert_eq!(session(&state).snapshot.stream_epoch, "epoch-2");
    assert_snapshot_error(&mut state, initial.clone(), "snapshot_epoch_stale");
    assert_eq!(session(&state).snapshot.stream_epoch, "epoch-2");

    for index in 3..=34 {
        let epoch = format!("epoch-{index}");
        state.apply_snapshot(&epoch_snapshot(&epoch)).unwrap();
        assert_eq!(session(&state).snapshot.stream_epoch, epoch);
    }
    assert_eq!(state.revision("session-1"), Some(34));
    assert_snapshot_error(&mut state, initial, "snapshot_epoch_stale");
    assert_eq!(session(&state).snapshot.stream_epoch, "epoch-34");
}

#[test]
fn same_cursor_snapshot_refreshes_server_digest_after_deltas() {
    let mut state = AgentWorkbenchState::new();
    state.apply_snapshot(&snapshot()).unwrap();
    state
        .apply_delta_batch(&batch(4, 5, 10, vec![status_event(10, 5)], "batch-5"))
        .unwrap();
    let commit_revision = session(&state).commit_revision;
    let mut authoritative = session(&state).snapshot.clone();
    authoritative.digest = Some("sha256:authoritative".into());

    state.apply_snapshot(&authoritative).unwrap();

    assert_eq!(
        session(&state).snapshot.digest.as_deref(),
        Some("sha256:authoritative"),
    );
    assert_eq!(session(&state).commit_revision, commit_revision + 1);
}

#[test]
fn digest_window_is_bounded_and_evicted_replay_requires_resync() {
    let mut state = AgentWorkbenchState::new();
    let mut initial = snapshot();
    initial.revision = 0;
    initial.latest_sequence = 0;
    state.apply_snapshot(&initial).unwrap();
    let first = sequential_batch(1);
    for revision in 1..=257 {
        state
            .apply_delta_batch(&sequential_batch(revision))
            .unwrap();
    }

    assert!(matches!(
        state.apply_delta_batch(&first),
        Err(AgentWorkbenchError::ResyncRequired {
            reason: ResyncReason::BaseRevisionMismatch,
            ..
        })
    ));
}

#[test]
fn app_state_initializes_workbench_without_replacing_acp() {
    let mut state = AppState::new();
    assert!(state.workbench.get_session("missing").is_none());
    assert!(state.acp.get_session("missing").is_none());
    state.workbench.apply_snapshot(&snapshot()).unwrap();
    state.reset_for_org_switch();
    assert!(state.workbench.get_session("session-1").is_none());
}

pub(crate) fn loaded_state() -> AgentWorkbenchState {
    let mut state = AgentWorkbenchState::new();
    let mut value = snapshot();
    value.command_receipts = vec![receipt("command-1", v2::CommandReceiptState::Accepted)];
    state.apply_snapshot(&value).unwrap();
    state
}

pub(crate) fn receipt_changed(
    sequence: u64,
    revision: u64,
    value: v2::CommandReceipt,
) -> v2::AgentEvent {
    let changed = v2::CommandReceiptChanged {
        receipt: Some(value),
    };
    event(
        sequence,
        revision,
        "receipt-change",
        v2::agent_event::Event::CommandReceiptChanged(changed),
    )
}

pub(crate) fn apply_one(
    state: &mut AgentWorkbenchState,
    base_revision: u64,
    sequence: u64,
    event: v2::AgentEvent,
    digest: &str,
) -> Result<(), AgentWorkbenchError> {
    let batch = batch(
        base_revision,
        base_revision + 1,
        sequence,
        vec![event],
        digest,
    );
    state.apply_delta_batch(&batch)
}

pub(crate) fn unsupported_history(snapshot: &v2::SessionSnapshot) -> &v2::UnsupportedValue {
    let content = snapshot.history[1].content.as_ref().unwrap();
    match content.content.as_ref().unwrap() {
        v2::timeline_item_content::Content::Unsupported(value) => value,
        _ => panic!("unsupported event missing"),
    }
}

fn session(state: &AgentWorkbenchState) -> &AgentWorkbenchSession {
    state.get_session("session-1").unwrap()
}

fn assert_snapshot_error(
    state: &mut AgentWorkbenchState,
    value: v2::SessionSnapshot,
    reason: &str,
) {
    match state.apply_snapshot(&value) {
        Err(AgentWorkbenchError::InvalidPayload { reason: actual }) => assert_eq!(actual, reason),
        result => panic!("unexpected snapshot result: {result:?}"),
    }
}

fn epoch_snapshot(epoch: &str) -> v2::SessionSnapshot {
    let mut value = snapshot();
    value.stream_epoch = epoch.into();
    value.revision = 0;
    value.latest_sequence = 0;
    value
}

fn invalid_snapshots() -> Vec<v2::SessionSnapshot> {
    let changes: [fn(&mut v2::SessionSnapshot); 12] = [
        |value| value.stream_epoch.clear(),
        |value| value.revision = value.latest_sequence + 1,
        |value| value.latest_sequence = 0,
        |value| {
            let mut item = timeline_item("item-1", 9);
            item.content = Some(v2::TimelineItemContent::default());
            value.history = vec![item];
        },
        |value| {
            value.history = vec![timeline_item("item-1", 9), timeline_item("item-2", 9)];
        },
        |value| {
            let mut receipt = receipt("bad", v2::CommandReceiptState::Received);
            receipt.payload_digest.clear();
            value.command_receipts = vec![receipt];
        },
        |value| {
            value.command_receipts = vec![receipt("dup", v2::CommandReceiptState::Received); 2];
        },
        |value| {
            let mut grant = grant("grant-1");
            grant.session_id = "session-2".into();
            value.grants = vec![grant];
        },
        |value| {
            value.permission_requests = vec![v2::PermissionRequest {
                permission_request_id: "permission-1".into(),
                state: v2::PermissionRequestState::Pending as i32,
                ..Default::default()
            }];
        },
        |value| {
            let mut request = permission_request("permission-1");
            request.state = v2::PermissionRequestState::Resolved as i32;
            request.resolution = Some(v2::PermissionResolution {
                permission_request_id: "other".into(),
                decision: v2::PermissionDecision::Unspecified as i32,
                ..Default::default()
            });
            value.permission_requests = vec![request];
        },
        |value| {
            let mut resource = terminal_resource("resource-1");
            resource.status = v2::SessionResourceStatus::Unspecified as i32;
            value.resources = vec![resource];
        },
        |value| {
            let mut artifact = artifact("artifact-1");
            artifact.status = v2::ArtifactStatus::Unspecified as i32;
            value.artifacts = vec![artifact];
        },
    ];
    changes.into_iter().map(changed_snapshot).collect()
}

fn changed_snapshot(change: fn(&mut v2::SessionSnapshot)) -> v2::SessionSnapshot {
    let mut value = snapshot();
    change(&mut value);
    value
}

pub(crate) fn status_event(sequence: u64, revision: u64) -> v2::AgentEvent {
    let changed = v2::SessionStatusChanged {
        status: v2::SessionStatus::Running as i32,
        ..Default::default()
    };
    event(
        sequence,
        revision,
        &format!("status-{sequence}"),
        v2::agent_event::Event::SessionStatusChanged(changed),
    )
}

pub(crate) fn permission_request(id: &str) -> v2::PermissionRequest {
    let approval = v2::PermissionApproval {
        title: "Approve".into(),
        ..Default::default()
    };
    v2::PermissionRequest {
        permission_request_id: id.into(),
        state: v2::PermissionRequestState::Pending as i32,
        request: Some(v2::permission_request::Request::Approval(approval)),
        ..Default::default()
    }
}

pub(crate) fn terminal_resource(id: &str) -> v2::SessionResource {
    let terminal = v2::TerminalResource {
        writable: true,
        control_mode: v2::TerminalControlMode::Surface as i32,
        ..Default::default()
    };
    v2::SessionResource {
        resource_id: id.into(),
        label: "Terminal".into(),
        status: v2::SessionResourceStatus::Ready as i32,
        resource: Some(v2::session_resource::Resource::Terminal(terminal)),
        ..Default::default()
    }
}

pub(crate) fn artifact(id: &str) -> v2::ArtifactDescriptor {
    v2::ArtifactDescriptor {
        artifact_id: id.into(),
        revision: 1,
        filename: "result.txt".into(),
        media_type: "text/plain".into(),
        status: v2::ArtifactStatus::Ready as i32,
        ..Default::default()
    }
}

fn grant(id: &str) -> v2::AuthorizationGrant {
    v2::AuthorizationGrant {
        grant_id: id.into(),
        session_id: "session-1".into(),
        ..Default::default()
    }
}

fn sequential_batch(revision: u64) -> v2::SessionDeltaBatch {
    batch(
        revision - 1,
        revision,
        revision,
        vec![status_event(revision, revision)],
        &format!("batch-{revision}"),
    )
}
