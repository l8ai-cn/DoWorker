use agentcloud_types::proto_agent_workbench_v2 as v2;

pub(crate) fn same_canonical_content(
    current: &v2::SessionSnapshot,
    next: &v2::SessionSnapshot,
) -> bool {
    let mut current = current.clone();
    let mut next = next.clone();
    clear_viewer_metadata(&mut current);
    clear_viewer_metadata(&mut next);
    current == next
}

fn clear_viewer_metadata(snapshot: &mut v2::SessionSnapshot) {
    snapshot.digest = None;
    snapshot.grants.clear();
    for artifact in &mut snapshot.artifacts {
        artifact.grants.clear();
    }
}
