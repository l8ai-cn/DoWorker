use crate::agent_workbench_stream_status::AgentWorkbenchStreamStatus;

#[test]
fn stream_terminal_status_distinguishes_client_remote_and_error() {
    let mut client_closed = AgentWorkbenchStreamStatus::default();
    assert_eq!(client_closed.code(), "open");
    assert!(client_closed.mark_client_closed());
    assert_eq!(client_closed.code(), "client_closed");
    assert_eq!(client_closed.error(), None);
    assert!(!client_closed.mark_remote_closed());

    let mut remote_closed = AgentWorkbenchStreamStatus::default();
    assert!(remote_closed.mark_remote_closed());
    assert_eq!(remote_closed.code(), "remote_closed");
    assert_eq!(remote_closed.error(), None);

    let mut failed = AgentWorkbenchStreamStatus::default();
    assert!(failed.mark_failed("missing final frame".into()));
    assert_eq!(failed.code(), "failed");
    assert_eq!(failed.error(), Some("missing final frame"));
}
