use std::sync::Arc;
use wasm_bindgen::prelude::*;

#[cfg(any(target_arch = "wasm32", test))]
mod agent_workbench_stream_status;
#[cfg(target_arch = "wasm32")]
mod agent_workbench_stream_wasm;
mod api;
mod api_services;
mod auth;
mod auth_proto_convert;
mod events_manager;
mod js_bridge;
mod protocol;
mod relay_manager;
mod service_agent;
mod service_agent_workbench;
mod service_ai_resource;
mod service_apikey;
mod service_auth_connect;
mod service_autopilot;
mod service_billing;
mod service_binding;
mod service_binding_connect;
mod service_blockstore;
mod service_channel;
mod service_channel_connect;
mod service_env_bundle;
mod service_execution_cluster;
mod service_extension;
mod service_file;
mod service_goal_loop;
mod service_grant;
mod service_invitation;
mod service_kb;
mod service_kb_connect;
mod service_mesh;
mod service_mesh_connect;
mod service_notification;
mod service_orchestration_resource;
mod service_org;
mod service_pod;
mod service_pod_worker_creation;
mod service_promocode;
mod service_repository;
mod service_runner;
mod service_sso;
mod service_support_ticket;
mod service_ticket;
mod service_ticket_relations;
mod service_token_usage;
mod service_user;
mod service_user_credential;
mod service_workflow;
mod state_acp;
mod state_agent_workbench;
mod state_app;
mod state_autopilot;
mod state_channel;
mod state_expert;
mod state_loop_builder;
mod state_loopal;
mod state_mesh;
mod state_pod;
mod state_repo;
mod state_runner;
mod state_ticket;
mod state_workflow;
mod state_workflow_proto;
mod ws_transport;

#[cfg(test)]
mod service_orchestration_resource_tests;
#[cfg(test)]
mod service_pod_worker_creation_tests;
#[cfg(test)]
mod state_agent_workbench_tests;

#[cfg(target_arch = "wasm32")]
pub use agent_workbench_stream_wasm::*;
pub use api::*;
pub use auth::*;
pub use events_manager::*;
pub use protocol::*;
pub use relay_manager::*;
pub use service_agent::*;
pub use service_agent_workbench::*;
pub use service_ai_resource::*;
pub use service_apikey::*;
pub use service_auth_connect::*;
pub use service_autopilot::*;
pub use service_billing::*;
pub use service_binding::*;
pub use service_blockstore::*;
pub use service_channel::*;
pub use service_env_bundle::*;
pub use service_execution_cluster::*;
pub use service_extension::*;
pub use service_file::*;
pub use service_goal_loop::*;
pub use service_grant::*;
pub use service_invitation::*;
pub use service_kb::*;
pub use service_mesh::*;
pub use service_notification::*;
pub use service_orchestration_resource::*;
pub use service_org::*;
pub use service_pod::*;
pub use service_promocode::*;
pub use service_repository::*;
pub use service_runner::*;
pub use service_sso::*;
pub use service_support_ticket::*;
pub use service_ticket::*;
pub use service_ticket_relations::*;
pub use service_token_usage::*;
pub use service_user::*;
pub use service_user_credential::*;
pub use service_workflow::*;
pub use state_acp::*;
pub use state_agent_workbench::*;
pub use state_app::*;
pub use state_autopilot::*;
pub use state_channel::*;
pub use state_expert::*;
pub use state_loop_builder::*;
pub use state_loopal::*;
pub use state_mesh::*;
pub use state_pod::*;
pub use state_repo::*;
pub use state_runner::*;
pub use state_ticket::*;
pub use state_workflow::*;
pub use ws_transport::*;

pub(crate) fn new_memory_backend() -> Arc<dyn agentsmesh_persistence::StorageBackend> {
    use agentsmesh_persistence::StorageBackend;
    let b = Arc::new(agentsmesh_persistence::InMemoryBackend::new());
    let _ = b.migrate();
    b
}

#[wasm_bindgen(start)]
pub fn init_panic_hook() {
    console_error_panic_hook::set_once();
}

#[wasm_bindgen]
pub fn version() -> String {
    env!("CARGO_PKG_VERSION").to_string()
}

// Idempotent: repeated calls (React StrictMode double-init) are no-ops.
#[wasm_bindgen]
pub fn init_logger(level: String) -> Result<(), JsValue> {
    agentsmesh_logging::init(agentsmesh_logging::LogConfig::wasm_console(level))
        .map_err(|e| JsValue::from_str(&e.to_string()))?;
    agentsmesh_logging::install_panic_hook();
    Ok(())
}

#[wasm_bindgen]
pub fn log_event(level: String, target: String, msg: String) {
    agentsmesh_logging::log_event(&level, &target, &msg);
}
