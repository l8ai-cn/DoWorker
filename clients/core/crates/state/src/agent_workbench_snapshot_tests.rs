use agentcloud_types::proto_agent_workbench_v2 as v2;

use crate::agent_workbench_state::{AgentWorkbenchError, AgentWorkbenchState};
use crate::agent_workbench_test_fixtures::snapshot;

#[test]
fn same_cursor_refreshes_session_grants() {
    let mut state = AgentWorkbenchState::new();
    state.apply_snapshot(&snapshot()).unwrap();
    let mut authorized = snapshot();
    authorized.grants = vec![session_grant("grant-1", "agent.prompt.send")];

    state.apply_snapshot(&authorized).unwrap();

    let session = state.get_session("session-1").unwrap();
    assert_eq!(session.snapshot.grants, authorized.grants);
    assert_eq!(session.commit_revision, 2);
}

#[test]
fn same_cursor_refreshes_artifact_grants() {
    let mut state = AgentWorkbenchState::new();
    let mut initial = snapshot();
    initial.artifacts = vec![artifact()];
    state.apply_snapshot(&initial).unwrap();
    let mut authorized = initial.clone();
    authorized.artifacts[0].grants =
        vec![artifact_grant("artifact-grant-1", "artifact.content.read")];

    state.apply_snapshot(&authorized).unwrap();

    let session = state.get_session("session-1").unwrap();
    assert_eq!(
        session.snapshot.artifacts[0].grants,
        authorized.artifacts[0].grants,
    );
    assert_eq!(session.commit_revision, 2);
}

#[test]
fn same_cursor_still_rejects_canonical_content_changes() {
    let mut state = AgentWorkbenchState::new();
    state.apply_snapshot(&snapshot()).unwrap();
    let mut changed = snapshot();
    changed.status = v2::SessionStatus::Idle as i32;

    assert!(matches!(
        state.apply_snapshot(&changed),
        Err(AgentWorkbenchError::InvalidPayload {
            reason: "snapshot_cursor_conflict"
        })
    ));
}

fn session_grant(id: &str, action: &str) -> v2::AuthorizationGrant {
    v2::AuthorizationGrant {
        grant_id: id.into(),
        issuer: "backend".into(),
        subject: "user-1".into(),
        session_id: "session-1".into(),
        actions: vec![action.into()],
        issued_at: "2026-07-16T00:00:00Z".into(),
        ..Default::default()
    }
}

fn artifact() -> v2::ArtifactDescriptor {
    v2::ArtifactDescriptor {
        artifact_id: "artifact-1".into(),
        revision: 1,
        filename: "result.txt".into(),
        media_type: "text/plain".into(),
        status: v2::ArtifactStatus::Ready as i32,
        ..Default::default()
    }
}

fn artifact_grant(id: &str, action: &str) -> v2::ArtifactGrant {
    v2::ArtifactGrant {
        grant_id: id.into(),
        issuer: Some("backend".into()),
        subject: Some("user-1".into()),
        actions: vec![action.into()],
        issued_at: Some("2026-07-16T00:00:00Z".into()),
        ..Default::default()
    }
}
