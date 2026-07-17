use agentsmesh_types::proto_agent_workbench_v2 as v2;

#[test]
fn service_requests_are_generated_from_v2_proto() {
    let snapshot = v2::GetSessionSnapshotRequest {
        org_slug: "acme".into(),
        session_id: "session-1".into(),
    };
    let deltas = v2::StreamSessionDeltasRequest {
        org_slug: "acme".into(),
        cursor: Some(v2::SessionCursor {
            session_id: "session-1".into(),
            stream_epoch: "epoch-1".into(),
            revision: 4,
            sequence: 9,
        }),
        replay_limit: 128,
    };
    let command = v2::ExecuteCommandRequest {
        org_slug: "acme".into(),
        command: Some(v2::CommandEnvelope::default()),
    };

    assert_eq!(snapshot.session_id, "session-1");
    assert_eq!(deltas.replay_limit, 128);
    assert!(command.command.is_some());
}
