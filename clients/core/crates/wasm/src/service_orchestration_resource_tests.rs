use crate::{WasmApiClient, WasmOrchestrationResourceService};

#[test]
fn orchestration_resource_wasm_surface_exposes_creation_and_calls() {
    let _ = WasmApiClient::create_orchestration_resource_service;
    let _ = WasmOrchestrationResourceService::validate_resource_connect;
    let _ = WasmOrchestrationResourceService::plan_resource_connect;
    let _ = WasmOrchestrationResourceService::get_resource_connect;
    let _ = WasmOrchestrationResourceService::list_resources_connect;
    let _ = WasmOrchestrationResourceService::export_resource_connect;
    let _ = WasmOrchestrationResourceService::get_resource_plan_connect;
}
