use crate::{WasmApiClient, WasmOrchestrationResourceService};

#[test]
fn orchestration_resource_wasm_surface_exposes_creation_and_calls() {
    let _ = WasmApiClient::create_orchestration_resource_service;
    let _ = WasmOrchestrationResourceService::validate_resource_connect;
    let _ = WasmOrchestrationResourceService::plan_resource_connect;
    let _ = WasmOrchestrationResourceService::get_resource_connect;
    let _ = WasmOrchestrationResourceService::get_resource_capabilities_connect;
    let _ = WasmOrchestrationResourceService::list_resources_connect;
    let _ = WasmOrchestrationResourceService::export_resource_connect;
    let _ = WasmOrchestrationResourceService::get_resource_plan_connect;
    let _ = WasmOrchestrationResourceService::apply_binding_resource_plan_connect;
    let _ = WasmOrchestrationResourceService::apply_worker_template_plan_connect;
    let _ = WasmOrchestrationResourceService::create_worker_from_plan_connect;
    let _ = WasmOrchestrationResourceService::create_goal_loop_from_plan_connect;
    let _ = WasmOrchestrationResourceService::apply_prompt_plan_connect;
    let _ = WasmOrchestrationResourceService::apply_expert_plan_connect;
    let _ = WasmOrchestrationResourceService::apply_workflow_plan_connect;
}
