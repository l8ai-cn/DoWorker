use crate::connect_call::connect_call;
use crate::error::ApiError;
use crate::ApiClient;
use orchestration_resource_proto::proto::orchestration_resource::v1 as resource;

impl ApiClient {
    pub async fn validate_resource_connect(
        &self,
        request: &resource::ValidateResourceRequest,
    ) -> Result<resource::ValidateResourceResponse, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/ValidateResource",
            request,
        )
        .await
    }

    pub async fn plan_resource_connect(
        &self,
        request: &resource::PlanResourceRequest,
    ) -> Result<resource::PlanResourceResponse, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/PlanResource",
            request,
        )
        .await
    }

    pub async fn get_resource_connect(
        &self,
        request: &resource::GetResourceRequest,
    ) -> Result<resource::Resource, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/GetResource",
            request,
        )
        .await
    }

    pub async fn list_resources_connect(
        &self,
        request: &resource::ListResourcesRequest,
    ) -> Result<resource::ListResourcesResponse, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/ListResources",
            request,
        )
        .await
    }

    pub async fn export_resource_connect(
        &self,
        request: &resource::ExportResourceRequest,
    ) -> Result<resource::ExportResourceResponse, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/ExportResource",
            request,
        )
        .await
    }

    pub async fn get_resource_plan_connect(
        &self,
        request: &resource::GetResourcePlanRequest,
    ) -> Result<resource::ResourcePlan, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/GetResourcePlan",
            request,
        )
        .await
    }

    pub async fn apply_binding_resource_plan_connect(
        &self,
        request: &resource::ApplyBindingResourcePlanRequest,
    ) -> Result<resource::Resource, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyBindingResourcePlan",
            request,
        )
        .await
    }

    pub async fn apply_worker_template_plan_connect(
        &self,
        request: &resource::ApplyWorkerTemplatePlanRequest,
    ) -> Result<resource::ApplyWorkerTemplatePlanResponse, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyWorkerTemplatePlan",
            request,
        )
        .await
    }

    pub async fn create_worker_from_plan_connect(
        &self,
        request: &resource::CreateWorkerFromPlanRequest,
    ) -> Result<resource::CreateWorkerFromPlanResponse, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/CreateWorkerFromPlan",
            request,
        )
        .await
    }

    pub async fn create_goal_loop_from_plan_connect(
        &self,
        request: &resource::CreateGoalLoopFromPlanRequest,
    ) -> Result<resource::CreateGoalLoopFromPlanResponse, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/CreateGoalLoopFromPlan",
            request,
        )
        .await
    }

    pub async fn apply_prompt_plan_connect(
        &self,
        request: &resource::ApplyPromptPlanRequest,
    ) -> Result<resource::Resource, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyPromptPlan",
            request,
        )
        .await
    }

    pub async fn apply_expert_plan_connect(
        &self,
        request: &resource::ApplyExpertPlanRequest,
    ) -> Result<resource::ApplyExpertPlanResponse, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyExpertPlan",
            request,
        )
        .await
    }

    pub async fn apply_workflow_plan_connect(
        &self,
        request: &resource::ApplyWorkflowPlanRequest,
    ) -> Result<resource::ApplyWorkflowPlanResponse, ApiError> {
        connect_call(
            self,
            "/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyWorkflowPlan",
            request,
        )
        .await
    }
}
