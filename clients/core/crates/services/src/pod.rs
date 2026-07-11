use std::sync::Arc;

use agentsmesh_api_client::ApiClient;
use agentsmesh_types::proto_pod_v1 as pod_proto;
use prost::Message;

// Networking-only service for the pod domain. The pod cache lives in the
// shared `AppState.pods` (dispatch-hook SSOT), reached via the wasm
// `app_pod*` surface — this service speaks only the Connect-RPC wire.
pub struct PodService {
    client: Arc<ApiClient>,
}

impl PodService {
    pub fn new(client: Arc<ApiClient>) -> Self {
        Self { client }
    }

    // -------- Connect-RPC (binary wire) --------

    pub async fn list_pods_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = pod_proto::ListPodsRequest::decode(request_bytes)
            .map_err(|e| format!("decode list_pods request: {e}"))?;
        tracing::debug!(target: "pod", org_slug = %req.org_slug, "list pods");
        let resp = self.client.list_pods_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn get_pod_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = pod_proto::GetPodRequest::decode(request_bytes)
            .map_err(|e| format!("decode get_pod request: {e}"))?;
        tracing::debug!(target: "pod", org_slug = %req.org_slug, pod_key = %req.pod_key, "get pod");
        let resp = self.client.get_pod_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn create_pod_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = pod_proto::CreatePodRequest::decode(request_bytes)
            .map_err(|e| format!("decode create_pod request: {e}"))?;
        tracing::info!(target: "pod", org_slug = %req.org_slug, agent_slug = %req.agent_slug, runner_id = ?req.runner_id, "create pod");
        let resp = self.client.create_pod_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn list_worker_create_options_connect(
        &self,
        request_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        let req = pod_proto::ListWorkerCreateOptionsRequest::decode(request_bytes)
            .map_err(|e| format!("decode list_worker_create_options request: {e}"))?;
        tracing::debug!(target: "pod", org_slug = %req.org_slug, "list worker create options");
        let resp = self.client
            .list_worker_create_options_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn preflight_worker_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = pod_proto::PreflightWorkerRequest::decode(request_bytes)
            .map_err(|e| format!("decode preflight_worker request: {e}"))?;
        tracing::debug!(target: "pod", org_slug = %req.org_slug, "preflight worker");
        let resp = self.client
            .preflight_worker_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn fill_worker_draft_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = pod_proto::FillWorkerDraftRequest::decode(request_bytes)
            .map_err(|e| format!("decode fill_worker_draft request: {e}"))?;
        tracing::debug!(target: "pod", org_slug = %req.org_slug, prompt_len = req.prompt.len(), "fill worker draft");
        let resp = self.client
            .fill_worker_draft_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn terminate_pod_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = pod_proto::TerminatePodRequest::decode(request_bytes)
            .map_err(|e| format!("decode terminate_pod request: {e}"))?;
        tracing::info!(target: "pod", org_slug = %req.org_slug, pod_key = %req.pod_key, "terminate pod");
        let resp = self.client.terminate_pod_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn update_pod_alias_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = pod_proto::UpdatePodAliasRequest::decode(request_bytes)
            .map_err(|e| format!("decode update_pod_alias request: {e}"))?;
        tracing::info!(target: "pod", org_slug = %req.org_slug, pod_key = %req.pod_key, "update pod alias");
        let resp = self.client.update_pod_alias_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn update_pod_perpetual_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = pod_proto::UpdatePodPerpetualRequest::decode(request_bytes)
            .map_err(|e| format!("decode update_pod_perpetual request: {e}"))?;
        tracing::info!(target: "pod", org_slug = %req.org_slug, pod_key = %req.pod_key, perpetual = req.perpetual, "update pod perpetual");
        let resp = self.client.update_pod_perpetual_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn update_pod_preview_config_connect(
        &self,
        request_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        let req = pod_proto::UpdatePodPreviewConfigRequest::decode(request_bytes)
            .map_err(|e| format!("decode update_pod_preview_config request: {e}"))?;
        let resp = self.client
            .update_pod_preview_config_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn get_mobile_access_descriptor_connect(
        &self,
        request_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        let req = pod_proto::GetMobileAccessDescriptorRequest::decode(request_bytes)
            .map_err(|e| format!("decode get_mobile_access_descriptor request: {e}"))?;
        let resp = self.client
            .get_mobile_access_descriptor_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn get_pod_connection_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = pod_proto::GetPodConnectionRequest::decode(request_bytes)
            .map_err(|e| format!("decode get_pod_connection request: {e}"))?;
        tracing::debug!(target: "pod", org_slug = %req.org_slug, pod_key = %req.pod_key, "get pod connection");
        let resp = self.client.get_pod_connection_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn send_pod_prompt_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = pod_proto::SendPodPromptRequest::decode(request_bytes)
            .map_err(|e| format!("decode send_pod_prompt request: {e}"))?;
        tracing::info!(target: "pod", org_slug = %req.org_slug, pod_key = %req.pod_key, prompt_len = req.prompt.len(), "send pod prompt");
        let resp = self.client.send_pod_prompt_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }

    pub async fn list_pods_by_ticket_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        let req = pod_proto::ListPodsByTicketRequest::decode(request_bytes)
            .map_err(|e| format!("decode list_pods_by_ticket request: {e}"))?;
        tracing::debug!(target: "pod", org_slug = %req.org_slug, ticket_id = req.ticket_id, "list pods by ticket");
        let resp = self.client.list_pods_by_ticket_connect(&req).await.map_err(crate::wire)?;
        Ok(resp.encode_to_vec())
    }
}
