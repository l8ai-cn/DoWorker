use agentsmesh_types::proto_agent_workbench_v2 as v2;

use crate::agent_workbench_artifacts::upsert_artifact;
use crate::agent_workbench_receipts::upsert_receipt;
use crate::agent_workbench_state::AgentWorkbenchError;
use crate::agent_workbench_validation::{
    invalid, known, valid_artifact, valid_permission, valid_resource, valid_timeline_content,
};

macro_rules! upsert_valid {
    ($items:expr, $value:expr, $id:ident, $valid:expr, $reason:literal) => {{
        let item = present($value.as_ref())?;
        if !$valid(item) {
            return Err(invalid($reason));
        }
        upsert(&mut $items, item.clone(), |current| current.$id == item.$id);
    }};
}

pub(crate) fn apply_batch(
    snapshot: &mut v2::SessionSnapshot,
    batch: &v2::SessionDeltaBatch,
) -> Result<(), AgentWorkbenchError> {
    batch
        .events
        .iter()
        .try_for_each(|event| apply_event(snapshot, event))
}

fn apply_event(
    snapshot: &mut v2::SessionSnapshot,
    event: &v2::AgentEvent,
) -> Result<(), AgentWorkbenchError> {
    let envelope = event.envelope.as_ref().expect("validated event envelope");
    match event.event.as_ref().expect("validated event payload") {
        v2::agent_event::Event::TimelineItemAppended(value) => {
            let content = timeline_content(value.content.as_ref())?;
            if history_index(snapshot, &envelope.item_id).is_some() {
                return Err(invalid("timeline_item_conflict"));
            }
            snapshot.history.push(timeline_item(envelope, content));
        }
        v2::agent_event::Event::TimelineItemUpdated(value) => {
            let content = timeline_content(value.content.as_ref())?;
            let index = history_index(snapshot, &envelope.item_id)
                .ok_or(invalid("timeline_item_missing"))?;
            snapshot.history[index] = timeline_item(envelope, content);
        }
        v2::agent_event::Event::CommandReceiptChanged(value) => {
            upsert_receipt(
                snapshot,
                present(value.receipt.as_ref())?,
                envelope.revision,
            )?;
        }
        v2::agent_event::Event::PermissionRequested(value) => {
            upsert_valid!(
                snapshot.permission_requests,
                value.request,
                permission_request_id,
                valid_permission,
                "permission_invalid"
            );
        }
        v2::agent_event::Event::PermissionResolved(value) => {
            apply_permission_resolution(snapshot, present(value.resolution.as_ref())?)?;
        }
        v2::agent_event::Event::ResourceChanged(value) => {
            upsert_valid!(
                snapshot.resources,
                value.resource,
                resource_id,
                valid_resource,
                "resource_invalid"
            );
        }
        v2::agent_event::Event::ArtifactChanged(value) => {
            let artifact = present(value.artifact.as_ref())?;
            if !valid_artifact(artifact) {
                return Err(invalid("artifact_invalid"));
            }
            upsert_artifact(&mut snapshot.artifacts, artifact);
        }
        v2::agent_event::Event::TerminalLeaseChanged(value) => {
            apply_terminal_lease(snapshot, value)?;
        }
        v2::agent_event::Event::CapabilitiesChanged(value) => {
            snapshot.capabilities = Some(present(value.capabilities.as_ref())?.clone());
        }
        v2::agent_event::Event::ConfigurationChanged(value) => {
            snapshot.configuration = Some(present(value.configuration.as_ref())?.clone());
        }
        v2::agent_event::Event::SessionStatusChanged(value) => {
            if !known::<v2::SessionStatus>(value.status) {
                return Err(invalid("session_status_invalid"));
            }
            snapshot.status = value.status;
            snapshot.error = value.error.clone();
        }
        v2::agent_event::Event::Unsupported(value) => {
            if history_index(snapshot, &envelope.item_id).is_some() {
                return Err(invalid("timeline_item_conflict"));
            }
            let content = v2::TimelineItemContent {
                content: Some(v2::timeline_item_content::Content::Unsupported(
                    value.clone(),
                )),
            };
            snapshot.history.push(timeline_item(envelope, &content));
        }
    }
    Ok(())
}

fn apply_permission_resolution(
    snapshot: &mut v2::SessionSnapshot,
    resolution: &v2::PermissionResolution,
) -> Result<(), AgentWorkbenchError> {
    if resolution.permission_request_id.is_empty()
        || !known::<v2::PermissionDecision>(resolution.decision)
    {
        return Err(invalid("permission_resolution_invalid"));
    }
    let request = snapshot
        .permission_requests
        .iter_mut()
        .find(|item| item.permission_request_id == resolution.permission_request_id)
        .ok_or(invalid("permission_request_missing"))?;
    request.state = v2::PermissionRequestState::Resolved as i32;
    request.resolution = Some(resolution.clone());
    Ok(())
}

fn apply_terminal_lease(
    snapshot: &mut v2::SessionSnapshot,
    value: &v2::TerminalLeaseChanged,
) -> Result<(), AgentWorkbenchError> {
    let lease = present(value.lease.as_ref())?;
    if value.resource_id.is_empty()
        || lease.lease_id.is_empty()
        || lease.holder.is_empty()
        || lease.expires_at.is_empty()
        || !known::<v2::TerminalLeaseState>(lease.state)
    {
        return Err(invalid("terminal_lease_invalid"));
    }
    let resource = snapshot
        .resources
        .iter_mut()
        .find(|item| item.resource_id == value.resource_id)
        .ok_or(invalid("terminal_resource_missing"))?;
    let Some(v2::session_resource::Resource::Terminal(terminal)) = resource.resource.as_mut()
    else {
        return Err(invalid("terminal_resource_missing"));
    };
    terminal.lease = Some(lease.clone());
    Ok(())
}

fn timeline_content(
    value: Option<&v2::TimelineItemContent>,
) -> Result<&v2::TimelineItemContent, AgentWorkbenchError> {
    present(value).and_then(|content| {
        valid_timeline_content(content)
            .then_some(content)
            .ok_or(invalid("timeline_content_invalid"))
    })
}

fn history_index(snapshot: &v2::SessionSnapshot, item_id: &str) -> Option<usize> {
    snapshot.history.iter().position(|item| {
        item.envelope
            .as_ref()
            .is_some_and(|value| value.item_id == item_id)
    })
}

fn timeline_item(
    envelope: &v2::EventEnvelope,
    content: &v2::TimelineItemContent,
) -> v2::TimelineItem {
    v2::TimelineItem {
        envelope: Some(envelope.clone()),
        content: Some(content.clone()),
    }
}

fn upsert<T>(items: &mut Vec<T>, value: T, same: impl Fn(&T) -> bool) {
    if let Some(index) = items.iter().position(same) {
        items[index] = value;
    } else {
        items.push(value);
    }
}

fn present<T>(value: Option<&T>) -> Result<&T, AgentWorkbenchError> {
    value.ok_or(invalid("event_field_missing"))
}
