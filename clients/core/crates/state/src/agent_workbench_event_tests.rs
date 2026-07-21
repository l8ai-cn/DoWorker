use agentcloud_types::proto_agent_workbench_v2 as v2;

use crate::agent_workbench_state::{AgentWorkbenchError, AgentWorkbenchState};
use crate::agent_workbench_state_tests::{
    apply_one, artifact, loaded_state, permission_request, receipt_changed, terminal_resource,
    unsupported_history,
};
use crate::agent_workbench_test_fixtures::{
    batch, envelope, event, receipt, receipt_event, snapshot, timeline_event, timeline_item,
};

#[test]
fn projects_all_v2_event_payloads_losslessly() {
    let mut state = AgentWorkbenchState::new();
    let mut initial = snapshot();
    initial.history.push(timeline_item("message-1", 9));
    state.apply_snapshot(&initial).unwrap();
    let unsupported = unsupported();
    let events = vec![
        timeline_event(10, "message-1", 5, true),
        permission_requested(11),
        permission_resolved(12),
        resource_changed(13),
        terminal_lease_changed(14),
        artifact_changed(15),
        capabilities_changed(16),
        failed_status(17),
        event(
            18,
            5,
            "unsupported",
            v2::agent_event::Event::Unsupported(unsupported.clone()),
        ),
    ];
    state
        .apply_delta_batch(&batch(4, 5, 10, events, "all-events"))
        .unwrap();

    let snapshot = &state.get_session("session-1").unwrap().snapshot;
    assert_eq!(snapshot.history[0].envelope.as_ref().unwrap().sequence, 10);
    let request = &snapshot.permission_requests[0];
    assert_eq!(request.state, v2::PermissionRequestState::Resolved as i32);
    assert_eq!(
        request.resolution.as_ref().unwrap().decision,
        v2::PermissionDecision::Accept as i32
    );
    let terminal = match snapshot.resources[0].resource.as_ref().unwrap() {
        v2::session_resource::Resource::Terminal(value) => value,
        _ => panic!("terminal resource missing"),
    };
    assert_eq!(terminal.lease.as_ref().unwrap().lease_id, "lease-1");
    assert_eq!(snapshot.artifacts[0].artifact_id, "artifact-1");
    assert_eq!(
        snapshot.capabilities.as_ref().unwrap().protocol_version,
        "2"
    );
    assert_eq!(snapshot.status, v2::SessionStatus::Failed as i32);
    assert_eq!(snapshot.error.as_ref().unwrap().code, "agent_failed");
    assert_eq!(unsupported_history(snapshot), &unsupported);
}

#[test]
fn preserves_artifact_revision_history_across_live_deltas() {
    let mut state = AgentWorkbenchState::new();
    let mut initial = snapshot();
    initial.artifacts = vec![versioned_artifact(1, "tool-one")];
    state.apply_snapshot(&initial).unwrap();
    let changed = v2::ArtifactChanged {
        artifact: Some(versioned_artifact(2, "tool-two")),
    };
    let update = event(
        10,
        5,
        "artifact-update",
        v2::agent_event::Event::ArtifactChanged(changed),
    );

    state
        .apply_delta_batch(&batch(4, 5, 10, vec![update], "artifact-update"))
        .unwrap();

    let artifact = &state
        .get_session("session-1")
        .unwrap()
        .snapshot
        .artifacts[0];
    assert_eq!(artifact.revision, 2);
    assert_eq!(
        artifact
            .revisions
            .iter()
            .map(|revision| revision
                .provenance
                .as_ref()
                .unwrap()
                .tool_execution_id
                .as_deref()
                .unwrap())
            .collect::<Vec<_>>(),
        ["tool-one", "tool-two"],
    );
}

#[test]
fn terminal_receipt_is_valid_but_cannot_roll_back() {
    let mut state = loaded_state();
    let terminal = receipt_event(
        10,
        5,
        "terminal-command",
        v2::CommandReceiptState::Succeeded,
    );
    apply_one(&mut state, 4, 10, terminal, "terminal").unwrap();
    let before = state.get_session("session-1").unwrap().snapshot.clone();
    let rollback = receipt_event(11, 6, "terminal-command", v2::CommandReceiptState::Running);

    assert!(matches!(
        apply_one(&mut state, 5, 11, rollback, "rollback"),
        Err(AgentWorkbenchError::ReceiptTransition { .. })
    ));
    let session = state.get_session("session-1").unwrap();
    assert_eq!(session.snapshot, before);
    assert_eq!(session.commit_revision, 2);
}

#[test]
fn terminal_receipt_same_state_requires_identical_content() {
    let mut state = AgentWorkbenchState::new();
    let mut initial = snapshot();
    let terminal = receipt("terminal-command", v2::CommandReceiptState::Succeeded);
    initial.command_receipts = vec![terminal.clone()];
    state.apply_snapshot(&initial).unwrap();
    let identical = receipt_changed(10, 5, terminal);
    apply_one(&mut state, 4, 10, identical, "identical-terminal").unwrap();

    let before = state.get_session("session-1").unwrap().snapshot.clone();
    let mut changed = receipt("terminal-command", v2::CommandReceiptState::Succeeded);
    changed.updated_at = Some("2026-07-16T00:00:01Z".into());
    let changed = receipt_changed(11, 6, changed);
    assert!(matches!(
        apply_one(&mut state, 5, 11, changed, "changed-terminal"),
        Err(AgentWorkbenchError::ReceiptTransition { .. })
    ));
    assert_eq!(state.get_session("session-1").unwrap().snapshot, before);
}

#[test]
fn first_visible_receipt_states_are_supported_except_unspecified() {
    let states = [
        v2::CommandReceiptState::Received,
        v2::CommandReceiptState::Accepted,
        v2::CommandReceiptState::Running,
        v2::CommandReceiptState::Succeeded,
    ];
    for receipt_state in states {
        let mut state = AgentWorkbenchState::new();
        state.apply_snapshot(&snapshot()).unwrap();
        let event = receipt_event(10, 5, "first", receipt_state);
        apply_one(&mut state, 4, 10, event, "first-visible").unwrap();
    }

    let mut state = AgentWorkbenchState::new();
    state.apply_snapshot(&snapshot()).unwrap();
    let invalid = receipt_event(10, 5, "invalid", v2::CommandReceiptState::Unspecified);
    assert!(matches!(
        apply_one(&mut state, 4, 10, invalid, "unspecified"),
        Err(AgentWorkbenchError::InvalidPayload { .. })
    ));
}

#[test]
fn same_receipt_state_updates_metadata_but_digest_conflict_fails() {
    let mut state = AgentWorkbenchState::new();
    let mut initial = snapshot();
    let mut received = receipt("metadata", v2::CommandReceiptState::Received);
    received.updated_at = Some("2026-07-16T00:00:00Z".into());
    initial.command_receipts = vec![received];
    state.apply_snapshot(&initial).unwrap();

    let mut updated = receipt("metadata", v2::CommandReceiptState::Received);
    updated.updated_at = Some("2026-07-16T00:00:01Z".into());
    let updated = receipt_changed(10, 5, updated);
    apply_one(&mut state, 4, 10, updated, "metadata").unwrap();
    let current = &state
        .get_session("session-1")
        .unwrap()
        .snapshot
        .command_receipts[0];
    assert_eq!(current.updated_at, Some("2026-07-16T00:00:01Z".into()));

    let mut conflict = receipt("metadata", v2::CommandReceiptState::Received);
    conflict.payload_digest = "different".into();
    let conflict = receipt_changed(11, 6, conflict);
    assert!(matches!(
        apply_one(&mut state, 5, 11, conflict, "digest-conflict"),
        Err(AgentWorkbenchError::DigestConflict { .. })
    ));
}

#[test]
fn event_failure_is_atomic() {
    let mut state = loaded_state();
    let invalid = batch(
        4,
        5,
        10,
        vec![
            timeline_event(10, "new-item", 5, false),
            timeline_event(11, "missing-item", 5, true),
        ],
        "invalid",
    );

    assert!(matches!(
        state.apply_delta_batch(&invalid),
        Err(AgentWorkbenchError::InvalidPayload { .. })
    ));
    let session = state.get_session("session-1").unwrap();
    assert!(session.snapshot.history.is_empty());
    assert_eq!(session.snapshot.revision, 4);
    assert_eq!(session.commit_revision, 1);
}

#[test]
fn event_timeline_content_oneof_is_required() {
    let mut state = loaded_state();
    let invalid = event(
        10,
        5,
        "empty-content",
        v2::agent_event::Event::TimelineItemAppended(v2::TimelineItemAppended {
            content: Some(v2::TimelineItemContent::default()),
        }),
    );
    assert!(matches!(
        state.apply_delta_batch(&batch(4, 5, 10, vec![invalid], "empty-content")),
        Err(AgentWorkbenchError::InvalidPayload { .. })
    ));
}

#[test]
fn unsupported_values_are_deeply_validated_everywhere() {
    for invalid in invalid_unsupported_values() {
        let raw = event(
            10,
            5,
            "raw-unsupported",
            v2::agent_event::Event::Unsupported(invalid.clone()),
        );
        assert_invalid_event(raw, "raw-unsupported");

        let content = v2::TimelineItemContent {
            content: Some(v2::timeline_item_content::Content::Unsupported(
                invalid.clone(),
            )),
        };
        let timeline = event(
            10,
            5,
            "timeline-unsupported",
            v2::agent_event::Event::TimelineItemAppended(v2::TimelineItemAppended {
                content: Some(content.clone()),
            }),
        );
        assert_invalid_event(timeline, "timeline-unsupported");

        let mut state = AgentWorkbenchState::new();
        let mut invalid_snapshot = snapshot();
        invalid_snapshot.history = vec![v2::TimelineItem {
            envelope: Some(envelope(9, "snapshot-unsupported", 4)),
            content: Some(content),
        }];
        assert!(matches!(
            state.apply_snapshot(&invalid_snapshot),
            Err(AgentWorkbenchError::InvalidPayload { .. })
        ));
    }
}

fn permission_requested(sequence: u64) -> v2::AgentEvent {
    let requested = v2::PermissionRequested {
        request: Some(permission_request("permission-1")),
    };
    event(
        sequence,
        5,
        "permission-requested",
        v2::agent_event::Event::PermissionRequested(requested),
    )
}

fn permission_resolved(sequence: u64) -> v2::AgentEvent {
    let resolution = v2::PermissionResolution {
        permission_request_id: "permission-1".into(),
        decision: v2::PermissionDecision::Accept as i32,
        ..Default::default()
    };
    event(
        sequence,
        5,
        "permission-resolved",
        v2::agent_event::Event::PermissionResolved(v2::PermissionResolved {
            resolution: Some(resolution),
        }),
    )
}

fn resource_changed(sequence: u64) -> v2::AgentEvent {
    let changed = v2::ResourceChanged {
        resource: Some(terminal_resource("resource-1")),
    };
    event(
        sequence,
        5,
        "resource",
        v2::agent_event::Event::ResourceChanged(changed),
    )
}

fn terminal_lease_changed(sequence: u64) -> v2::AgentEvent {
    let lease = v2::TerminalLease {
        lease_id: "lease-1".into(),
        holder: "client-1".into(),
        state: v2::TerminalLeaseState::Active as i32,
        expires_at: "2026-07-16T01:00:00Z".into(),
        fencing_epoch: 1,
    };
    let changed = v2::TerminalLeaseChanged {
        resource_id: "resource-1".into(),
        lease: Some(lease),
    };
    event(
        sequence,
        5,
        "lease",
        v2::agent_event::Event::TerminalLeaseChanged(changed),
    )
}

fn artifact_changed(sequence: u64) -> v2::AgentEvent {
    let changed = v2::ArtifactChanged {
        artifact: Some(artifact("artifact-1")),
    };
    event(
        sequence,
        5,
        "artifact",
        v2::agent_event::Event::ArtifactChanged(changed),
    )
}

fn versioned_artifact(revision: u64, tool_execution_id: &str) -> v2::ArtifactDescriptor {
    let mut artifact = artifact("artifact-1");
    artifact.revision = revision;
    artifact.revisions = vec![v2::ArtifactRevision {
        revision,
        provenance: Some(v2::ArtifactProvenance {
            tool_execution_id: Some(tool_execution_id.into()),
            ..Default::default()
        }),
        ..Default::default()
    }];
    artifact
}

fn capabilities_changed(sequence: u64) -> v2::AgentEvent {
    let capabilities = v2::SupportCapabilities {
        protocol_version: "2".into(),
        ..Default::default()
    };
    event(
        sequence,
        5,
        "capabilities",
        v2::agent_event::Event::CapabilitiesChanged(v2::CapabilitiesChanged {
            capabilities: Some(capabilities),
        }),
    )
}

fn failed_status(sequence: u64) -> v2::AgentEvent {
    let error = v2::AgentError {
        code: "agent_failed".into(),
        message: "agent stopped".into(),
        ..Default::default()
    };
    event(
        sequence,
        5,
        "failed",
        v2::agent_event::Event::SessionStatusChanged(v2::SessionStatusChanged {
            status: v2::SessionStatus::Failed as i32,
            error: Some(error),
        }),
    )
}

fn unsupported() -> v2::UnsupportedValue {
    v2::UnsupportedValue {
        identity: Some(v2::ContentIdentity {
            namespace: "runner.event".into(),
            semantic_key: "future".into(),
            schema_version: "3".into(),
            ..Default::default()
        }),
        reason: v2::UnsupportedReason::Unknown as i32,
        payload: Some(v2::StructuredPayload {
            media_type: "application/octet-stream".into(),
            data: vec![0, 1, 255],
        }),
    }
}

fn invalid_unsupported_values() -> Vec<v2::UnsupportedValue> {
    let mut values = Vec::new();
    for field in 0..3 {
        let mut value = unsupported();
        let identity = value.identity.as_mut().unwrap();
        match field {
            0 => identity.namespace.clear(),
            1 => identity.semantic_key.clear(),
            _ => identity.schema_version.clear(),
        }
        values.push(value);
    }
    let mut reason = unsupported();
    reason.reason = v2::UnsupportedReason::Unspecified as i32;
    values.push(reason);
    let mut media_type = unsupported();
    media_type.payload.as_mut().unwrap().media_type.clear();
    values.push(media_type);
    values
}

fn assert_invalid_event(value: v2::AgentEvent, digest: &str) {
    let mut state = loaded_state();
    assert!(matches!(
        state.apply_delta_batch(&batch(4, 5, 10, vec![value], digest)),
        Err(AgentWorkbenchError::InvalidPayload { .. })
    ));
}
