pub mod acp_session;
pub mod acp_types;
mod agent_workbench_artifacts;
mod agent_workbench_receipts;
mod agent_workbench_reducer;
mod agent_workbench_snapshot;
pub mod agent_workbench_state;
mod agent_workbench_validation;
pub mod app_runtime;
pub mod app_state;
pub mod auth_types;
pub mod autopilot_state;
pub mod autopilot_wire_convert;
pub mod blockstore_apply;
pub mod blockstore_state;
pub mod blockstore_types;
pub mod channel_state;
pub mod channel_types;
pub mod credential_types;
pub mod event_dispatch;
pub mod expert_state;
pub mod expert_types;
pub mod loop_builder_state;
pub mod loopal_dispatch;
pub mod loopal_session;
pub mod loopal_types;
pub mod mesh_state;
pub mod notification_specs;
mod persist_helpers;
pub mod pod_query_snapshots;
pub mod pod_state;
pub mod repo_state;
pub mod runner_state;
pub mod ticket_state;
pub mod workflow_state;
pub mod workflow_types;

#[cfg(test)]
mod agent_workbench_test_fixtures {
    use agentsmesh_types::proto_agent_workbench_v2 as v2;

    pub(crate) fn snapshot() -> v2::SessionSnapshot {
        v2::SessionSnapshot {
            session_id: "session-1".into(),
            stream_epoch: "epoch-1".into(),
            revision: 4,
            latest_sequence: 9,
            status: v2::SessionStatus::Running as i32,
            ..Default::default()
        }
    }

    pub(crate) fn batch(
        base_revision: u64,
        revision: u64,
        first_sequence: u64,
        events: Vec<v2::AgentEvent>,
        digest: &str,
    ) -> v2::SessionDeltaBatch {
        v2::SessionDeltaBatch {
            session_id: "session-1".into(),
            stream_epoch: "epoch-1".into(),
            base_revision,
            revision,
            first_sequence,
            last_sequence: first_sequence + events.len() as u64 - 1,
            events,
            digest: digest.into(),
        }
    }

    pub(crate) fn envelope(sequence: u64, item_id: &str, revision: u64) -> v2::EventEnvelope {
        v2::EventEnvelope {
            session_id: "session-1".into(),
            stream_epoch: "epoch-1".into(),
            revision,
            sequence,
            item_id: item_id.into(),
            created_at: "2026-07-16T00:00:00Z".into(),
            ..Default::default()
        }
    }

    pub(crate) fn event(
        sequence: u64,
        revision: u64,
        item_id: &str,
        value: v2::agent_event::Event,
    ) -> v2::AgentEvent {
        v2::AgentEvent {
            envelope: Some(envelope(sequence, item_id, revision)),
            event: Some(value),
        }
    }

    pub(crate) fn system_content() -> v2::TimelineItemContent {
        v2::TimelineItemContent {
            content: Some(v2::timeline_item_content::Content::System(
                v2::SystemTimelineItem::default(),
            )),
        }
    }

    pub(crate) fn timeline_item(item_id: &str, sequence: u64) -> v2::TimelineItem {
        v2::TimelineItem {
            envelope: Some(envelope(sequence, item_id, 4)),
            content: Some(system_content()),
        }
    }

    pub(crate) fn timeline_event(
        sequence: u64,
        item_id: &str,
        revision: u64,
        update: bool,
    ) -> v2::AgentEvent {
        let content = Some(system_content());
        let value = if update {
            v2::agent_event::Event::TimelineItemUpdated(v2::TimelineItemUpdated { content })
        } else {
            v2::agent_event::Event::TimelineItemAppended(v2::TimelineItemAppended { content })
        };
        event(sequence, revision, item_id, value)
    }

    pub(crate) fn receipt(command_id: &str, state: v2::CommandReceiptState) -> v2::CommandReceipt {
        v2::CommandReceipt {
            session_id: "session-1".into(),
            command_id: command_id.into(),
            state: state as i32,
            payload_digest: format!("digest-{command_id}"),
            ..Default::default()
        }
    }

    pub(crate) fn receipt_event(
        sequence: u64,
        revision: u64,
        command_id: &str,
        state: v2::CommandReceiptState,
    ) -> v2::AgentEvent {
        let changed = v2::CommandReceiptChanged {
            receipt: Some(receipt(command_id, state)),
        };
        event(
            sequence,
            revision,
            &format!("receipt-{command_id}"),
            v2::agent_event::Event::CommandReceiptChanged(changed),
        )
    }
}

#[cfg(test)]
mod acp_session_tests;
#[cfg(test)]
mod agent_workbench_configuration_tests;
#[cfg(test)]
mod agent_workbench_event_tests;
#[cfg(test)]
mod agent_workbench_snapshot_tests;
#[cfg(test)]
mod agent_workbench_state_tests;
#[cfg(test)]
mod app_runtime_tests;
#[cfg(test)]
mod autopilot_state_tests;
#[cfg(test)]
mod blockstore_state_tests;
#[cfg(test)]
mod channel_state_tests;
#[cfg(test)]
mod expert_state_tests;
#[cfg(test)]
mod loop_builder_state_tests;
#[cfg(test)]
mod loopal_dispatch_tests;
#[cfg(test)]
mod loopal_fold_contract_tests;
#[cfg(test)]
mod mesh_state_tests;
#[cfg(test)]
mod pod_state_tests;
#[cfg(test)]
mod pod_query_snapshots_tests;
#[cfg(test)]
mod repo_state_tests;
#[cfg(test)]
mod runner_state_tests;
#[cfg(test)]
mod ticket_state_tests;
#[cfg(test)]
mod workflow_state_tests;
