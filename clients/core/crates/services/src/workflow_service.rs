use std::sync::Arc;

use agentsmesh_api_client::ApiClient;
use agentsmesh_types::proto_workflow_v1 as lp;
use prost::Message;

pub struct WorkflowService {
    client: Arc<ApiClient>,
}

impl WorkflowService {
    pub fn new(client: Arc<ApiClient>) -> Self {
        Self { client }
    }

    // -------- Connect-RPC (binary wire) --------
    //
    // The workflow + run cache is the AppState SSOT (runtime.state.workflows), fed by
    // the WorkflowRun* dispatch arms + the wasm workflow-state surface; this
    // service is networking-only.

    pub async fn list_workflows_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::ListWorkflowsRequest::decode(request_bytes)
            .map_err(|e| format!("decode list_workflows request: {e}"))?;
        tracing::debug!(target: "workflow", org_slug = %req.org_slug, "list workflows");
        let resp = self.client.list_workflows_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn get_workflow_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::GetWorkflowRequest::decode(request_bytes)
            .map_err(|e| format!("decode get_workflow request: {e}"))?;
        tracing::debug!(target: "workflow", org_slug = %req.org_slug, workflow_slug = %req.workflow_slug, "get workflow");
        let resp = self.client.get_workflow_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn create_workflow_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::CreateWorkflowRequest::decode(request_bytes)
            .map_err(|e| format!("decode create_workflow request: {e}"))?;
        tracing::info!(target: "workflow", org_slug = %req.org_slug, slug = %req.slug, agent_slug = %req.agent_slug, "create workflow");
        let resp = self.client.create_workflow_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn update_workflow_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::UpdateWorkflowRequest::decode(request_bytes)
            .map_err(|e| format!("decode update_workflow request: {e}"))?;
        tracing::info!(target: "workflow", org_slug = %req.org_slug, workflow_slug = %req.workflow_slug, "update workflow");
        let resp = self.client.update_workflow_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn delete_workflow_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::DeleteWorkflowRequest::decode(request_bytes)
            .map_err(|e| format!("decode delete_workflow request: {e}"))?;
        tracing::info!(target: "workflow", org_slug = %req.org_slug, workflow_slug = %req.workflow_slug, "delete workflow");
        let resp = self.client.delete_workflow_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn workflow_action_connect(
        &self, action: &str, request_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        let req = lp::WorkflowActionRequest::decode(request_bytes)
            .map_err(|e| format!("decode workflow_action request: {e}"))?;
        tracing::info!(target: "workflow", org_slug = %req.org_slug, workflow_slug = %req.workflow_slug, action, "workflow action");
        let resp = match action {
            "enable" => self.client.enable_workflow_connect(&req).await,
            "disable" => self.client.disable_workflow_connect(&req).await,
            other => return Err(format!("unknown workflow action: {other}")),
        }.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn trigger_workflow_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::TriggerWorkflowRequest::decode(request_bytes)
            .map_err(|e| format!("decode trigger_workflow request: {e}"))?;
        tracing::info!(target: "workflow", org_slug = %req.org_slug, workflow_slug = %req.workflow_slug, "trigger workflow");
        let resp = self.client.trigger_workflow_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn list_workflow_runs_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::ListWorkflowRunsRequest::decode(request_bytes)
            .map_err(|e| format!("decode list_workflow_runs request: {e}"))?;
        tracing::debug!(target: "workflow", org_slug = %req.org_slug, workflow_slug = %req.workflow_slug, "list workflow runs");
        let resp = self.client.list_workflow_runs_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn cancel_workflow_run_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::CancelWorkflowRunRequest::decode(request_bytes)
            .map_err(|e| format!("decode cancel_workflow_run request: {e}"))?;
        tracing::info!(target: "workflow", org_slug = %req.org_slug, workflow_slug = %req.workflow_slug, run_id = req.run_id, "cancel workflow run");
        let resp = self.client.cancel_workflow_run_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }
}
