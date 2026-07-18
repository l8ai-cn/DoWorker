use std::sync::Arc;

use agentsmesh_api_client::ApiClient;
use orchestration_resource_proto::proto::orchestration_resource::v1 as resource;
use prost::Message;

pub struct OrchestrationResourceService {
    client: Arc<ApiClient>,
}

macro_rules! wire_rpc {
    ($name:ident, $request:ty, $client_method:ident) => {
        pub async fn $name(&self, bytes: &[u8]) -> Result<Vec<u8>, String> {
            let request = <$request>::decode(bytes)
                .map_err(|error| format!("decode {} request: {error}", stringify!($name)))?;
            let response = self
                .client
                .$client_method(&request)
                .await
                .map_err(crate::wire)?;
            Ok(response.encode_to_vec())
        }
    };
}

impl OrchestrationResourceService {
    pub fn new(client: Arc<ApiClient>) -> Self {
        Self { client }
    }

    wire_rpc!(
        validate_resource_connect,
        resource::ValidateResourceRequest,
        validate_resource_connect
    );
    wire_rpc!(
        plan_resource_connect,
        resource::PlanResourceRequest,
        plan_resource_connect
    );
    wire_rpc!(
        get_resource_connect,
        resource::GetResourceRequest,
        get_resource_connect
    );
    wire_rpc!(
        get_resource_capabilities_connect,
        resource::GetResourceCapabilitiesRequest,
        get_resource_capabilities_connect
    );
    wire_rpc!(
        list_resources_connect,
        resource::ListResourcesRequest,
        list_resources_connect
    );
    wire_rpc!(
        export_resource_connect,
        resource::ExportResourceRequest,
        export_resource_connect
    );
    wire_rpc!(
        get_resource_plan_connect,
        resource::GetResourcePlanRequest,
        get_resource_plan_connect
    );
    wire_rpc!(
        apply_binding_resource_plan_connect,
        resource::ApplyBindingResourcePlanRequest,
        apply_binding_resource_plan_connect
    );
    wire_rpc!(
        apply_worker_template_plan_connect,
        resource::ApplyWorkerTemplatePlanRequest,
        apply_worker_template_plan_connect
    );
    wire_rpc!(
        create_worker_from_plan_connect,
        resource::CreateWorkerFromPlanRequest,
        create_worker_from_plan_connect
    );
    wire_rpc!(
        create_goal_loop_from_plan_connect,
        resource::CreateGoalLoopFromPlanRequest,
        create_goal_loop_from_plan_connect
    );
    wire_rpc!(
        apply_prompt_plan_connect,
        resource::ApplyPromptPlanRequest,
        apply_prompt_plan_connect
    );
    wire_rpc!(
        apply_expert_plan_connect,
        resource::ApplyExpertPlanRequest,
        apply_expert_plan_connect
    );
    wire_rpc!(
        apply_workflow_plan_connect,
        resource::ApplyWorkflowPlanRequest,
        apply_workflow_plan_connect
    );
}
