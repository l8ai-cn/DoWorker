use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum EventType {
    #[serde(rename = "pod:created")]
    PodCreated,
    #[serde(rename = "pod:status_changed")]
    PodStatusChanged,
    #[serde(rename = "pod:agent_status_changed")]
    PodAgentStatusChanged,
    #[serde(rename = "pod:terminated")]
    PodTerminated,
    #[serde(rename = "pod:title_changed")]
    PodTitleChanged,
    #[serde(rename = "pod:alias_changed")]
    PodAliasChanged,
    #[serde(rename = "pod:perpetual_changed")]
    PodPerpetualChanged,
    #[serde(rename = "pod:init_progress")]
    PodInitProgress,
    #[serde(rename = "pod:restarting")]
    PodRestarting,

    #[serde(rename = "channel:message")]
    ChannelMessage,
    #[serde(rename = "channel:message_edited")]
    ChannelMessageEdited,
    #[serde(rename = "channel:message_deleted")]
    ChannelMessageDeleted,
    #[serde(rename = "channel:member_added")]
    ChannelMemberAdded,
    #[serde(rename = "channel:member_removed")]
    ChannelMemberRemoved,

    #[serde(rename = "ticket:created")]
    TicketCreated,
    #[serde(rename = "ticket:updated")]
    TicketUpdated,
    #[serde(rename = "ticket:status_changed")]
    TicketStatusChanged,
    #[serde(rename = "ticket:moved")]
    TicketMoved,
    #[serde(rename = "ticket:deleted")]
    TicketDeleted,

    #[serde(rename = "runner:online")]
    RunnerOnline,
    #[serde(rename = "runner:offline")]
    RunnerOffline,
    #[serde(rename = "runner:updated")]
    RunnerUpdated,

    #[serde(rename = "autopilot:status_changed")]
    AutopilotStatusChanged,
    #[serde(rename = "autopilot:iteration")]
    AutopilotIteration,
    #[serde(rename = "autopilot:created")]
    AutopilotCreated,
    #[serde(rename = "autopilot:terminated")]
    AutopilotTerminated,
    #[serde(rename = "autopilot:thinking")]
    AutopilotThinking,

    #[serde(rename = "mr:created")]
    MrCreated,
    #[serde(rename = "mr:updated")]
    MrUpdated,
    #[serde(rename = "mr:merged")]
    MrMerged,
    #[serde(rename = "mr:closed")]
    MrClosed,

    #[serde(rename = "pipeline:updated")]
    PipelineUpdated,

    #[serde(rename = "workflow_run:started")]
    WorkflowRunStarted,
    #[serde(rename = "workflow_run:completed")]
    WorkflowRunCompleted,
    #[serde(rename = "workflow_run:failed")]
    WorkflowRunFailed,
    #[serde(rename = "workflow_run:warning")]
    WorkflowRunWarning,

    // Blockstore events — fan-out of every accepted op so other connected
    // clients can apply it to their cache without polling. See
    // backend/internal/infra/blockstore/op_publisher.go.
    #[serde(rename = "blockstore:op")]
    BlockstoreOp,

    #[serde(rename = "notification")]
    Notification,

    #[serde(rename = "system:maintenance")]
    SystemMaintenance,

    #[serde(rename = "connected")]
    Connected,
    #[serde(rename = "ping")]
    Ping,
    #[serde(rename = "pong")]
    Pong,
}

impl EventType {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::PodCreated => "pod:created",
            Self::PodStatusChanged => "pod:status_changed",
            Self::PodAgentStatusChanged => "pod:agent_status_changed",
            Self::PodTerminated => "pod:terminated",
            Self::PodTitleChanged => "pod:title_changed",
            Self::PodAliasChanged => "pod:alias_changed",
            Self::PodPerpetualChanged => "pod:perpetual_changed",
            Self::PodInitProgress => "pod:init_progress",
            Self::PodRestarting => "pod:restarting",
            Self::ChannelMessage => "channel:message",
            Self::ChannelMessageEdited => "channel:message_edited",
            Self::ChannelMessageDeleted => "channel:message_deleted",
            Self::ChannelMemberAdded => "channel:member_added",
            Self::ChannelMemberRemoved => "channel:member_removed",
            Self::TicketCreated => "ticket:created",
            Self::TicketUpdated => "ticket:updated",
            Self::TicketStatusChanged => "ticket:status_changed",
            Self::TicketMoved => "ticket:moved",
            Self::TicketDeleted => "ticket:deleted",
            Self::RunnerOnline => "runner:online",
            Self::RunnerOffline => "runner:offline",
            Self::RunnerUpdated => "runner:updated",
            Self::AutopilotStatusChanged => "autopilot:status_changed",
            Self::AutopilotIteration => "autopilot:iteration",
            Self::AutopilotCreated => "autopilot:created",
            Self::AutopilotTerminated => "autopilot:terminated",
            Self::AutopilotThinking => "autopilot:thinking",
            Self::MrCreated => "mr:created",
            Self::MrUpdated => "mr:updated",
            Self::MrMerged => "mr:merged",
            Self::MrClosed => "mr:closed",
            Self::PipelineUpdated => "pipeline:updated",
            Self::WorkflowRunStarted => "workflow_run:started",
            Self::WorkflowRunCompleted => "workflow_run:completed",
            Self::WorkflowRunFailed => "workflow_run:failed",
            Self::WorkflowRunWarning => "workflow_run:warning",
            Self::BlockstoreOp => "blockstore:op",
            Self::Notification => "notification",
            Self::SystemMaintenance => "system:maintenance",
            Self::Connected => "connected",
            Self::Ping => "ping",
            Self::Pong => "pong",
        }
    }
}

impl std::fmt::Display for EventType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}
