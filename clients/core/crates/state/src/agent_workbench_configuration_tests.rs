use agentcloud_types::proto_agent_workbench_v2 as v2;

use crate::agent_workbench_state::{AgentWorkbenchError, AgentWorkbenchState};
use crate::agent_workbench_test_fixtures::{batch, event, snapshot};

#[test]
fn configuration_changed_replaces_canonical_snapshot() {
    let mut state = AgentWorkbenchState::new();
    let mut initial = snapshot();
    initial.configuration = Some(configuration("model-old", "permission-old"));
    state.apply_snapshot(&initial).unwrap();

    let next = configuration("model-new", "permission-new");
    let changed = event(
        10,
        5,
        "configuration-1",
        v2::agent_event::Event::ConfigurationChanged(v2::ConfigurationChanged {
            configuration: Some(next.clone()),
        }),
    );
    state
        .apply_delta_batch(&batch(4, 5, 10, vec![changed], "configuration"))
        .unwrap();

    let session = state.get_session("session-1").unwrap();
    assert_eq!(session.snapshot.configuration, Some(next));
    assert_eq!(session.snapshot.revision, 5);
    assert_eq!(session.commit_revision, 2);
}

#[test]
fn configuration_changed_requires_configuration_and_is_atomic() {
    let mut state = AgentWorkbenchState::new();
    let mut initial = snapshot();
    initial.configuration = Some(configuration("model-old", "permission-old"));
    state.apply_snapshot(&initial).unwrap();
    let before = state.get_session("session-1").unwrap().snapshot.clone();
    let changed = event(
        10,
        5,
        "configuration-empty",
        v2::agent_event::Event::ConfigurationChanged(v2::ConfigurationChanged {
            configuration: None,
        }),
    );

    assert!(matches!(
        state.apply_delta_batch(&batch(4, 5, 10, vec![changed], "configuration-empty")),
        Err(AgentWorkbenchError::InvalidPayload { .. })
    ));
    let session = state.get_session("session-1").unwrap();
    assert_eq!(session.snapshot, before);
    assert_eq!(session.commit_revision, 1);
}

fn configuration(model: &str, permission_mode: &str) -> v2::SessionConfiguration {
    v2::SessionConfiguration {
        model: Some(model.into()),
        permission_mode: Some(permission_mode.into()),
    }
}
