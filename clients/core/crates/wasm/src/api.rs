use std::sync::Arc;

use agentsmesh_api_client::{ApiClient, AuthTokenStore};
use agentsmesh_events::{EventSubscriptionManager, EventSubscriptionManagerOptions};
use agentsmesh_state::app_state::AppRuntime;
use agentsmesh_transport::runtime::PlatformRuntime;
use wasm_bindgen::prelude::*;

mod service_factories;

#[wasm_bindgen]
pub struct WasmApiClient {
    client: Arc<ApiClient>,
    base_url: String,
    /// Singleton AppRuntime for this client. Created once at construction
    /// time and shared by all services + the events manager. The events
    /// manager's dispatch hook is wired into `AppRuntime.state.dispatch`
    /// at this point, so any event delivered through the connection_workflow
    /// updates the same `AppState` that services + selectors read from.
    runtime: Arc<AppRuntime>,
}

#[wasm_bindgen]
impl WasmApiClient {
    #[wasm_bindgen(constructor)]
    pub fn new(base_url: String, auth: &crate::auth::WasmAuthManager) -> Self {
        let store: Arc<dyn AuthTokenStore> = auth.token_store_arc();
        let client = Arc::new(ApiClient::new(base_url.clone(), store));
        let events = Arc::new(EventSubscriptionManager::with_runtime(
            PlatformRuntime,
            client.clone(),
            EventSubscriptionManagerOptions::default(),
        ));
        let runtime = AppRuntime::new(events);
        Self {
            client,
            base_url,
            runtime,
        }
    }

    #[wasm_bindgen(getter)]
    pub fn base_url(&self) -> String {
        self.base_url.clone()
    }

    // ── AppRuntime state views ──
    // These return per-domain view structs over the SINGLE shared
    // `AppState`. All views observe the same writes — including writes
    // from `EventSubscriptionManager.dispatch_event` (via the dispatch
    // hook installed in `AppRuntime::new`). Stable wasm-bindgen API,
    // safe to call from JS many times.

    pub fn get_pod_state(&self) -> crate::state_pod::WasmPodState {
        crate::state_pod::WasmPodState::from_runtime(self.runtime.state.clone())
    }

    pub fn get_channel_state(&self) -> crate::state_channel::WasmChannelState {
        crate::state_channel::WasmChannelState::from_runtime(self.runtime.state.clone())
    }

    pub fn get_ticket_state(&self) -> crate::state_ticket::WasmTicketState {
        crate::state_ticket::WasmTicketState::from_runtime(self.runtime.state.clone())
    }

    pub fn get_runner_state(&self) -> crate::state_runner::WasmRunnerState {
        crate::state_runner::WasmRunnerState::from_runtime(self.runtime.state.clone())
    }

    pub fn get_workflow_state(&self) -> crate::state_workflow::WasmWorkflowState {
        crate::state_workflow::WasmWorkflowState::from_runtime(self.runtime.state.clone())
    }

    pub fn get_mesh_state(&self) -> crate::state_mesh::WasmMeshState {
        crate::state_mesh::WasmMeshState::from_runtime(self.runtime.state.clone())
    }

    pub fn get_autopilot_state(&self) -> crate::state_autopilot::WasmAutopilotState {
        crate::state_autopilot::WasmAutopilotState::from_runtime(self.runtime.state.clone())
    }

    pub fn get_repo_state(&self) -> crate::state_repo::WasmRepoState {
        crate::state_repo::WasmRepoState::from_runtime(self.runtime.state.clone())
    }

    pub fn get_expert_state(&self) -> crate::state_expert::WasmExpertState {
        crate::state_expert::WasmExpertState::from_runtime(self.runtime.state.clone())
    }

    pub fn get_acp_manager(&self) -> crate::state_acp::WasmAcpSessionManager {
        crate::state_acp::WasmAcpSessionManager::from_runtime(self.runtime.state.clone())
    }

    pub fn get_loopal_manager(&self) -> crate::state_loopal::WasmLoopalManager {
        crate::state_loopal::WasmLoopalManager::from_runtime(self.runtime.state.clone())
    }

    pub fn get_loop_builder_state(&self) -> crate::state_loop_builder::WasmLoopBuilderState {
        crate::state_loop_builder::WasmLoopBuilderState::from_runtime(self.runtime.state.clone())
    }

    // ── Pending side-effect drains ──
    // Rust SSOT dispatch queues side-effects (toast, browser notification,
    // refetch keys) into AppState.pending_*. JS drains these per tick;
    // each drain is atomic on the Rust side (take + clear).

    pub fn take_pending_toasts(&self) -> String {
        let toasts = self.runtime.state.write().take_pending_toasts();
        serde_json::to_string(&toasts).unwrap_or_else(|_| "[]".to_string())
    }

    pub fn take_pending_browser_notifications(&self) -> String {
        let notifs = self
            .runtime
            .state
            .write()
            .take_pending_browser_notifications();
        serde_json::to_string(&notifs).unwrap_or_else(|_| "[]".to_string())
    }

    pub fn take_pending_refetch_ticket_slugs(&self) -> String {
        let slugs = self
            .runtime
            .state
            .write()
            .take_pending_refetch_ticket_slugs();
        serde_json::to_string(&slugs).unwrap_or_else(|_| "[]".to_string())
    }

    pub fn take_pending_refetch_pod_keys(&self) -> String {
        let keys = self.runtime.state.write().take_pending_refetch_pod_keys();
        serde_json::to_string(&keys).unwrap_or_else(|_| "[]".to_string())
    }

    /// Tick counter — increments after every event dispatched to AppState.
    /// React selectors use this as the snapshot for `useSyncExternalStore`.
    pub fn tick(&self) -> f64 {
        self.runtime.events.tick() as f64
    }

    /// Clear all org-scoped state. Used on org switch; preserves the
    /// realtime connection + dispatch hook registration.
    pub fn reset_for_org_switch(&self) {
        self.runtime.state.write().reset_for_org_switch();
    }
}
