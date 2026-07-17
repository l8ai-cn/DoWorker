use std::collections::BTreeMap;

use agentsmesh_types::proto_agent_workbench_v2 as v2;

pub(crate) fn upsert_artifact(
    artifacts: &mut Vec<v2::ArtifactDescriptor>,
    value: &v2::ArtifactDescriptor,
) {
    let Some(index) = artifacts
        .iter()
        .position(|artifact| artifact.artifact_id == value.artifact_id)
    else {
        artifacts.push(value.clone());
        return;
    };

    let mut next = value.clone();
    let mut revisions = BTreeMap::new();
    for revision in artifacts[index]
        .revisions
        .iter()
        .chain(value.revisions.iter())
    {
        revisions.insert(revision.revision, revision.clone());
    }
    next.revisions = revisions.into_values().collect();
    artifacts[index] = next;
}
