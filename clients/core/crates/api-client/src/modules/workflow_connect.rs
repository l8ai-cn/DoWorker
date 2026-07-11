use crate::ApiClient;
use crate::connect_call::connect_call;
use crate::error::ApiError;
use agentsmesh_types::proto_workflow_v1 as lp;

// =============================================================================
// Connect-RPC (binary wire). See proto-naming-conventions.md §2.5.
// =============================================================================

impl ApiClient {
    pub async fn list_workflows_connect(
        &self,
        req: &lp::ListWorkflowsRequest,
    ) -> Result<lp::ListWorkflowsResponse, ApiError> {
        connect_call(self, "/proto.workflow.v1.WorkflowService/ListWorkflows", req).await
    }

    pub async fn get_workflow_connect(&self, req: &lp::GetWorkflowRequest) -> Result<lp::Workflow, ApiError> {
        connect_call(self, "/proto.workflow.v1.WorkflowService/GetWorkflow", req).await
    }

    pub async fn create_workflow_connect(
        &self,
        req: &lp::CreateWorkflowRequest,
    ) -> Result<lp::Workflow, ApiError> {
        connect_call(self, "/proto.workflow.v1.WorkflowService/CreateWorkflow", req).await
    }

    pub async fn update_workflow_connect(
        &self,
        req: &lp::UpdateWorkflowRequest,
    ) -> Result<lp::Workflow, ApiError> {
        connect_call(self, "/proto.workflow.v1.WorkflowService/UpdateWorkflow", req).await
    }

    pub async fn delete_workflow_connect(
        &self,
        req: &lp::DeleteWorkflowRequest,
    ) -> Result<lp::DeleteWorkflowResponse, ApiError> {
        connect_call(self, "/proto.workflow.v1.WorkflowService/DeleteWorkflow", req).await
    }

    pub async fn enable_workflow_connect(
        &self,
        req: &lp::WorkflowActionRequest,
    ) -> Result<lp::Workflow, ApiError> {
        connect_call(self, "/proto.workflow.v1.WorkflowService/EnableWorkflow", req).await
    }

    pub async fn disable_workflow_connect(
        &self,
        req: &lp::WorkflowActionRequest,
    ) -> Result<lp::Workflow, ApiError> {
        connect_call(self, "/proto.workflow.v1.WorkflowService/DisableWorkflow", req).await
    }

    pub async fn trigger_workflow_connect(
        &self,
        req: &lp::TriggerWorkflowRequest,
    ) -> Result<lp::TriggerWorkflowResponse, ApiError> {
        connect_call(self, "/proto.workflow.v1.WorkflowService/TriggerWorkflow", req).await
    }

    pub async fn list_workflow_runs_connect(
        &self,
        req: &lp::ListWorkflowRunsRequest,
    ) -> Result<lp::ListWorkflowRunsResponse, ApiError> {
        connect_call(self, "/proto.workflow.v1.WorkflowService/ListWorkflowRuns", req).await
    }

    pub async fn cancel_workflow_run_connect(
        &self,
        req: &lp::CancelWorkflowRunRequest,
    ) -> Result<lp::CancelWorkflowRunResponse, ApiError> {
        connect_call(self, "/proto.workflow.v1.WorkflowService/CancelWorkflowRun", req).await
    }
}
