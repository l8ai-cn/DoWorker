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
}
