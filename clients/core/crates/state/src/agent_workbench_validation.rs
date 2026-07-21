use std::collections::HashSet;

use agentcloud_types::proto_agent_workbench_v2 as v2;

use crate::agent_workbench_receipts::valid_receipt;
use crate::agent_workbench_state::{AgentWorkbenchError, AgentWorkbenchError::InvalidPayload};

pub(crate) fn invalid(reason: &'static str) -> AgentWorkbenchError {
    InvalidPayload { reason }
}

pub(crate) fn validate_snapshot(snapshot: &v2::SessionSnapshot) -> Result<(), AgentWorkbenchError> {
    if snapshot.session_id.is_empty()
        || snapshot.stream_epoch.is_empty()
        || !known::<v2::SessionStatus>(snapshot.status)
    {
        return Err(invalid("snapshot_identity"));
    }
    if snapshot.revision > snapshot.latest_sequence
        || (snapshot.revision == 0) != (snapshot.latest_sequence == 0)
    {
        return Err(invalid("snapshot_position"));
    }
    validate_history(snapshot)?;
    validate_unique(
        &snapshot.command_receipts,
        |value| value.command_id.as_str(),
        |value| valid_receipt(value, &snapshot.session_id, snapshot.revision),
        "snapshot_receipt",
    )?;
    validate_unique(
        &snapshot.grants,
        |value| value.grant_id.as_str(),
        |value| value.session_id == snapshot.session_id,
        "snapshot_grant",
    )?;
    validate_unique(
        &snapshot.permission_requests,
        |value| value.permission_request_id.as_str(),
        valid_permission,
        "snapshot_permission",
    )?;
    validate_unique(
        &snapshot.resources,
        |value| value.resource_id.as_str(),
        valid_resource,
        "snapshot_resource",
    )?;
    validate_unique(
        &snapshot.artifacts,
        |value| value.artifact_id.as_str(),
        valid_artifact,
        "snapshot_artifact",
    )
}

pub(crate) fn validate_batch(batch: &v2::SessionDeltaBatch) -> Result<(), AgentWorkbenchError> {
    let count = u64::try_from(batch.events.len()).map_err(|_| invalid("delta_event_count"))?;
    let range = batch
        .last_sequence
        .checked_sub(batch.first_sequence)
        .and_then(|value| value.checked_add(1));
    if batch.session_id.is_empty()
        || batch.stream_epoch.is_empty()
        || batch.digest.is_empty()
        || batch.events.is_empty()
        || batch.base_revision.checked_add(1) != Some(batch.revision)
        || range != Some(count)
    {
        return Err(invalid("delta_batch"));
    }
    for (offset, event) in batch.events.iter().enumerate() {
        let envelope = event.envelope.as_ref().ok_or(invalid("delta_envelope"))?;
        let payload = event.event.as_ref().ok_or(invalid("delta_envelope"))?;
        let sequence = batch
            .first_sequence
            .checked_add(offset as u64)
            .ok_or(invalid("delta_sequence"))?;
        if envelope.session_id != batch.session_id
            || envelope.stream_epoch != batch.stream_epoch
            || envelope.revision != batch.revision
            || envelope.sequence != sequence
            || envelope.item_id.is_empty()
            || envelope.created_at.is_empty()
        {
            return Err(invalid("delta_envelope"));
        }
        validate_event_payload(payload)?;
    }
    Ok(())
}

fn validate_history(snapshot: &v2::SessionSnapshot) -> Result<(), AgentWorkbenchError> {
    let mut item_ids = HashSet::new();
    let mut sequences = HashSet::new();
    for item in &snapshot.history {
        let envelope = item.envelope.as_ref().ok_or(invalid("snapshot_history"))?;
        if !item.content.as_ref().is_some_and(valid_timeline_content)
            || envelope.item_id.is_empty()
            || envelope.created_at.is_empty()
            || envelope.session_id != snapshot.session_id
            || envelope.stream_epoch != snapshot.stream_epoch
            || envelope.revision > snapshot.revision
            || envelope.sequence > snapshot.latest_sequence
            || !item_ids.insert(envelope.item_id.as_str())
            || !sequences.insert(envelope.sequence)
        {
            return Err(invalid("snapshot_history"));
        }
    }
    Ok(())
}

fn validate_event_payload(event: &v2::agent_event::Event) -> Result<(), AgentWorkbenchError> {
    let valid = match event {
        v2::agent_event::Event::ConfigurationChanged(value) => value.configuration.is_some(),
        v2::agent_event::Event::Unsupported(value) => valid_unsupported(value),
        v2::agent_event::Event::TimelineItemAppended(value) => {
            value.content.as_ref().is_some_and(valid_timeline_content)
        }
        v2::agent_event::Event::TimelineItemUpdated(value) => {
            value.content.as_ref().is_some_and(valid_timeline_content)
        }
        _ => true,
    };
    valid.then_some(()).ok_or(invalid("event_payload_invalid"))
}

pub(crate) fn valid_timeline_content(value: &v2::TimelineItemContent) -> bool {
    match value.content.as_ref() {
        Some(v2::timeline_item_content::Content::Unsupported(value)) => valid_unsupported(value),
        Some(_) => true,
        None => false,
    }
}

fn valid_unsupported(value: &v2::UnsupportedValue) -> bool {
    value.identity.as_ref().is_some_and(|identity| {
        !identity.namespace.is_empty()
            && !identity.semantic_key.is_empty()
            && !identity.schema_version.is_empty()
    }) && known::<v2::UnsupportedReason>(value.reason)
        && value
            .payload
            .as_ref()
            .is_some_and(|payload| !payload.media_type.is_empty())
}

fn validate_unique<'a, T>(
    values: &'a [T],
    id: impl Fn(&'a T) -> &'a str,
    valid: impl Fn(&T) -> bool,
    reason: &'static str,
) -> Result<(), AgentWorkbenchError> {
    let mut ids = HashSet::new();
    if values.iter().all(|value| {
        let key = id(value);
        !key.is_empty() && valid(value) && ids.insert(key)
    }) {
        Ok(())
    } else {
        Err(invalid(reason))
    }
}

pub(crate) fn valid_permission(value: &v2::PermissionRequest) -> bool {
    let Ok(state) = v2::PermissionRequestState::try_from(value.state) else {
        return false;
    };
    let request_valid = value.request.as_ref().is_none_or(|request| match request {
        v2::permission_request::Request::Unsupported(value) => valid_unsupported(value),
        _ => true,
    });
    let resolution_valid = value.resolution.as_ref().is_none_or(|resolution| {
        resolution.permission_request_id == value.permission_request_id
            && known::<v2::PermissionDecision>(resolution.decision)
    });
    state != v2::PermissionRequestState::Unspecified
        && (state != v2::PermissionRequestState::Pending || value.request.is_some())
        && (state != v2::PermissionRequestState::Resolved || value.resolution.is_some())
        && request_valid
        && resolution_valid
}

pub(crate) fn valid_resource(value: &v2::SessionResource) -> bool {
    known::<v2::SessionResourceStatus>(value.status)
        && value
            .resource
            .as_ref()
            .is_some_and(|resource| match resource {
                v2::session_resource::Resource::Unsupported(value) => valid_unsupported(value),
                _ => true,
            })
}

pub(crate) fn valid_artifact(value: &v2::ArtifactDescriptor) -> bool {
    value.revision > 0
        && !value.filename.is_empty()
        && !value.media_type.is_empty()
        && known::<v2::ArtifactStatus>(value.status)
}

pub(crate) fn known<T: TryFrom<i32>>(value: i32) -> bool {
    value != 0 && T::try_from(value).is_ok()
}
