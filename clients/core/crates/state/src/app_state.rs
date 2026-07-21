use std::sync::Arc;

use agentcloud_events::types::RealtimeEvent;
use agentcloud_persistence::StorageBackend;

use crate::acp_session::AcpSessionManager;
use crate::agent_workbench_state::AgentWorkbenchState;
pub use crate::app_runtime::{AppRuntime, AppStateDispatchHook};
use crate::autopilot_state::AutopilotState;
use crate::channel_state::ChannelState;
use crate::event_dispatch;
use crate::expert_state::ExpertState;
use crate::loop_builder_state::LoopBuilderState;
use crate::loopal_session::LoopalSessionManager;
use crate::mesh_state::MeshState;
pub use crate::notification_specs::{NotificationSpec, ToastSpec};
use crate::pod_query_snapshots::PodQuerySnapshots;
use crate::pod_state::PodState;
use crate::repo_state::RepoState;
use crate::runner_state::RunnerState;
use crate::ticket_state::TicketState;
use crate::workflow_state::WorkflowState;

pub struct AppState {
    pub pods: PodState,
    pub pod_query_snapshots: PodQuerySnapshots,
    pub channels: ChannelState,
    pub runners: RunnerState,
    pub tickets: TicketState,
    pub workflows: WorkflowState,
    pub mesh: MeshState,
    pub autopilot: AutopilotState,
    pub workbench: AgentWorkbenchState,
    pub acp: AcpSessionManager,
    pub loopal: LoopalSessionManager,
    pub loop_builder: LoopBuilderState,
    pub repo: RepoState,
    pub experts: ExpertState,

    /// Toast notifications queued by dispatch (workflow_run:warning,
    /// system:maintenance, etc). Drained per-tick by platform consumers.
    pub pending_toasts: Vec<ToastSpec>,
    /// Browser/OS-level notifications queued by `notification` events.
    pub pending_browser_notifications: Vec<NotificationSpec>,
    /// Ticket slugs whose details must be refetched (MR/Pipeline events
    /// only carry slug+podId — the actual MR/pipeline data must be
    /// pulled via Connect-RPC). Platforms drain per-tick.
    pub pending_refetch_ticket_slugs: Vec<String>,
    /// Pod keys whose details must be refetched (same rationale as above
    /// for MR/Pipeline events that carry only podId).
    pub pending_refetch_pod_keys: Vec<String>,
    /// Set to true when the realtime connection transitions back to
    /// `Connected` after a disconnect; platforms drain per-tick and run
    /// a global refetch (pods + tickets + channels) to catch up on
    /// events missed during the gap.
    pub pending_post_reconnect_refetch: bool,
    /// Autopilot controllers list needs refetch (created event carries
    /// partial data; full list pull picks up missing fields).
    pub pending_refetch_autopilot: bool,
}

impl AppState {
    pub fn new() -> Self {
        Self {
            pods: PodState::new(),
            pod_query_snapshots: PodQuerySnapshots::default(),
            channels: ChannelState::new(),
            runners: RunnerState::new(),
            tickets: TicketState::new(),
            workflows: WorkflowState::new(),
            mesh: MeshState::default(),
            autopilot: AutopilotState::default(),
            workbench: AgentWorkbenchState::new(),
            acp: AcpSessionManager::new(),
            loopal: LoopalSessionManager::new(),
            loop_builder: LoopBuilderState::new(),
            repo: RepoState::new(),
            experts: ExpertState::new(),
            pending_toasts: Vec::new(),
            pending_browser_notifications: Vec::new(),
            pending_refetch_ticket_slugs: Vec::new(),
            pending_refetch_pod_keys: Vec::new(),
            pending_post_reconnect_refetch: false,
            pending_refetch_autopilot: false,
        }
    }

    pub fn with_storage(backend: Arc<dyn StorageBackend>) -> Self {
        Self {
            pods: PodState::with_storage(backend.clone()),
            pod_query_snapshots: PodQuerySnapshots::default(),
            channels: ChannelState::with_storage(backend.clone()),
            runners: RunnerState::with_storage(backend.clone()),
            tickets: TicketState::with_storage(backend.clone()),
            workflows: WorkflowState::with_storage(backend.clone()),
            mesh: MeshState::default(),
            autopilot: AutopilotState::default(),
            workbench: AgentWorkbenchState::new(),
            acp: AcpSessionManager::new(),
            loopal: LoopalSessionManager::new(),
            loop_builder: LoopBuilderState::new(),
            repo: RepoState::with_storage(backend),
            experts: ExpertState::new(),
            pending_toasts: Vec::new(),
            pending_browser_notifications: Vec::new(),
            pending_refetch_ticket_slugs: Vec::new(),
            pending_refetch_pod_keys: Vec::new(),
            pending_post_reconnect_refetch: false,
            pending_refetch_autopilot: false,
        }
    }

    pub fn dispatch(&mut self, event: &RealtimeEvent) {
        event_dispatch::dispatch(self, event);
    }

    /// Atomic take-and-clear for pending toasts. Platform consumer drains
    /// this and emits via sonner / UNNotificationCenter etc. Items are
    /// not re-enqueued on consumer failure — log and drop.
    pub fn take_pending_toasts(&mut self) -> Vec<ToastSpec> {
        std::mem::take(&mut self.pending_toasts)
    }

    pub fn take_pending_browser_notifications(&mut self) -> Vec<NotificationSpec> {
        std::mem::take(&mut self.pending_browser_notifications)
    }

    pub fn take_pending_refetch_ticket_slugs(&mut self) -> Vec<String> {
        std::mem::take(&mut self.pending_refetch_ticket_slugs)
    }

    pub fn take_pending_refetch_pod_keys(&mut self) -> Vec<String> {
        std::mem::take(&mut self.pending_refetch_pod_keys)
    }

    pub fn take_pending_post_reconnect_refetch(&mut self) -> bool {
        std::mem::replace(&mut self.pending_post_reconnect_refetch, false)
    }

    pub fn take_pending_refetch_autopilot(&mut self) -> bool {
        std::mem::replace(&mut self.pending_refetch_autopilot, false)
    }

    /// Clear all org-scoped state on org switch. Keeps user-scoped
    /// settings (acp sessions, repo cache that's per-user) intact.
    /// Preferable to rebuilding the whole AppState because it preserves
    /// the live EventSubscriptionManager connection and its callbacks.
    pub fn reset_for_org_switch(&mut self) {
        self.pods = PodState::new();
        self.pod_query_snapshots = PodQuerySnapshots::default();
        self.channels = ChannelState::new();
        self.runners = RunnerState::new();
        self.tickets = TicketState::new();
        self.workflows = WorkflowState::new();
        self.mesh = MeshState::default();
        self.autopilot = AutopilotState::default();
        self.workbench = AgentWorkbenchState::new();
        self.experts = ExpertState::new();
        self.loop_builder.reset();
        self.pending_toasts.clear();
        self.pending_browser_notifications.clear();
        self.pending_refetch_ticket_slugs.clear();
        self.pending_refetch_pod_keys.clear();
        self.pending_post_reconnect_refetch = false;
        self.pending_refetch_autopilot = false;
    }
}

impl Default for AppState {
    fn default() -> Self {
        Self::new()
    }
}
