use crate::workflow_state::{WorkflowData, WorkflowRunData, WorkflowState, workflow_run_status};

fn make_workflow(slug: &str, name: &str, enabled: bool) -> WorkflowData {
    WorkflowData {
        slug: slug.into(),
        name: name.into(),
        description: None,
        schedule: Some("0 * * * *".into()),
        is_enabled: enabled,
        last_run_at: None,
        created_at: None,
        updated_at: None,
        ..Default::default()
    }
}

fn make_run(id: i64, workflow_slug: &str, status: &str) -> WorkflowRunData {
    WorkflowRunData {
        id,
        workflow_slug: workflow_slug.into(),
        status: status.into(),
        started_at: Some("2026-01-01T00:00:00Z".into()),
        completed_at: None,
        error_message: None,
        ..Default::default()
    }
}

#[test]
fn new_state_is_empty() {
    let state = WorkflowState::new();
    assert!(state.get_workflows().is_empty());
    assert!(state.get_current_workflow().is_none());
    assert!(state.get_runs().is_empty());
}

#[test]
fn set_and_get_workflows() {
    let mut state = WorkflowState::new();
    state.set_workflows(vec![
        make_workflow("l-1", "Hourly", true),
        make_workflow("l-2", "Daily", false),
    ]);
    assert_eq!(state.get_workflows().len(), 2);
    assert_eq!(state.get_workflows()[0].name, "Hourly");
}

#[test]
fn get_workflow_by_slug() {
    let mut state = WorkflowState::new();
    state.set_workflows(vec![
        make_workflow("l-1", "Hourly", true),
        make_workflow("l-2", "Daily", false),
    ]);
    let found = state.get_workflow_by_slug("l-2");
    assert!(found.is_some());
    assert_eq!(found.unwrap().name, "Daily");
    assert!(state.get_workflow_by_slug("l-999").is_none());
}

#[test]
fn set_and_get_current_workflow() {
    let mut state = WorkflowState::new();
    assert!(state.get_current_workflow().is_none());
    state.set_current_workflow(Some(make_workflow("l-1", "Active", true)));
    assert_eq!(state.get_current_workflow().unwrap().slug, "l-1");
    state.set_current_workflow(None);
    assert!(state.get_current_workflow().is_none());
}

#[test]
fn add_run() {
    let mut state = WorkflowState::new();
    state.add_run(make_run(1, "l-1", workflow_run_status::RUNNING));
    state.add_run(make_run(2, "l-1", workflow_run_status::COMPLETED));
    assert_eq!(state.get_runs().len(), 2);
    assert_eq!(state.get_runs()[0].status, workflow_run_status::RUNNING);
    assert_eq!(state.get_runs()[1].status, workflow_run_status::COMPLETED);
}

#[test]
fn update_run_status() {
    let mut state = WorkflowState::new();
    state.add_run(make_run(1, "l-1", workflow_run_status::RUNNING));
    state.update_run_status(1, workflow_run_status::COMPLETED);
    assert_eq!(state.get_runs()[0].status, workflow_run_status::COMPLETED);
}

#[test]
fn update_run_status_nonexistent_is_noop() {
    let mut state = WorkflowState::new();
    state.add_run(make_run(1, "l-1", workflow_run_status::RUNNING));
    state.update_run_status(999, workflow_run_status::FAILED);
    assert_eq!(state.get_runs()[0].status, workflow_run_status::RUNNING);
}

#[test]
fn clear_runs() {
    let mut state = WorkflowState::new();
    state.add_run(make_run(1, "l-1", workflow_run_status::COMPLETED));
    state.add_run(make_run(2, "l-1", workflow_run_status::FAILED));
    assert_eq!(state.get_runs().len(), 2);
    state.clear_runs();
    assert!(state.get_runs().is_empty());
}

#[test]
fn set_workflows_replaces_all() {
    let mut state = WorkflowState::new();
    state.set_workflows(vec![make_workflow("l-1", "First", true)]);
    assert_eq!(state.get_workflows().len(), 1);
    state.set_workflows(vec![
        make_workflow("l-2", "A", true),
        make_workflow("l-3", "B", false),
    ]);
    assert_eq!(state.get_workflows().len(), 2);
    assert!(state.get_workflow_by_slug("l-1").is_none());
}

#[test]
fn default_impl() {
    let state = WorkflowState::default();
    assert!(state.get_workflows().is_empty());
    assert!(state.get_runs().is_empty());
}

#[test]
fn multiple_runs_different_workflows() {
    let mut state = WorkflowState::new();
    state.add_run(make_run(1, "l-1", workflow_run_status::RUNNING));
    state.add_run(make_run(2, "l-2", workflow_run_status::COMPLETED));
    assert_eq!(state.get_runs().len(), 2);
    assert_eq!(state.get_runs()[0].workflow_slug, "l-1");
    assert_eq!(state.get_runs()[1].workflow_slug, "l-2");
}
