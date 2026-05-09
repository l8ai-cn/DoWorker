use std::sync::Arc;

use agentsmesh_api_client::ApiClient;
use agentsmesh_types::*;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmMessageService {
    client: Arc<ApiClient>,
}

#[wasm_bindgen]
impl WasmMessageService {
    pub(crate) fn new(client: Arc<ApiClient>) -> Self {
        Self { client }
    }

    pub async fn send_message(
        &self, json: &str, pod_key: Option<String>,
    ) -> Result<String, String> {
        let req: SendDirectMessageRequest = serde_json::from_str(json).map_err(agentsmesh_services::wire)?;
        let resp = self.client
            .send_mesh_message(&req, pod_key.as_deref())
            .await.map_err(agentsmesh_services::wire)?;
        serde_json::to_string(&resp).map_err(agentsmesh_services::wire)
    }

    pub async fn get_messages(
        &self, unread_only: Option<bool>, limit: Option<u32>, offset: Option<u32>,
    ) -> Result<String, String> {
        let resp = self.client
            .get_mesh_messages(unread_only, limit, offset)
            .await.map_err(agentsmesh_services::wire)?;
        serde_json::to_string(&resp).map_err(agentsmesh_services::wire)
    }

    pub async fn get_unread_count(&self) -> Result<String, String> {
        let resp = self.client.get_mesh_unread_count().await.map_err(agentsmesh_services::wire)?;
        serde_json::to_string(&resp).map_err(agentsmesh_services::wire)
    }

    pub async fn get_message(&self, id: i64) -> Result<String, String> {
        let resp = self.client.get_mesh_message(id).await.map_err(agentsmesh_services::wire)?;
        serde_json::to_string(&resp).map_err(agentsmesh_services::wire)
    }

    pub async fn mark_read(&self, json: &str) -> Result<String, String> {
        let req: MarkMessagesReadRequest = serde_json::from_str(json).map_err(agentsmesh_services::wire)?;
        let resp = self.client.mark_mesh_messages_read(&req).await.map_err(agentsmesh_services::wire)?;
        serde_json::to_string(&resp).map_err(agentsmesh_services::wire)
    }

    pub async fn mark_all_read(&self) -> Result<String, String> {
        let resp = self.client.mark_all_mesh_messages_read().await.map_err(agentsmesh_services::wire)?;
        serde_json::to_string(&resp).map_err(agentsmesh_services::wire)
    }

    pub async fn get_conversation(
        &self, correlation_id: &str, limit: Option<u32>,
    ) -> Result<String, String> {
        let resp = self.client
            .get_mesh_conversation(correlation_id, limit)
            .await.map_err(agentsmesh_services::wire)?;
        serde_json::to_string(&resp).map_err(agentsmesh_services::wire)
    }

    pub async fn get_sent_messages(
        &self, limit: Option<u32>, offset: Option<u32>,
    ) -> Result<String, String> {
        let resp = self.client
            .get_mesh_sent_messages(limit, offset)
            .await.map_err(agentsmesh_services::wire)?;
        serde_json::to_string(&resp).map_err(agentsmesh_services::wire)
    }

    pub async fn get_dead_letters(
        &self, limit: Option<u32>, offset: Option<u32>,
    ) -> Result<String, String> {
        let resp = self.client
            .get_mesh_dead_letters(limit, offset)
            .await.map_err(agentsmesh_services::wire)?;
        serde_json::to_string(&resp).map_err(agentsmesh_services::wire)
    }

    pub async fn replay_dead_letter(&self, entry_id: i64) -> Result<String, String> {
        let resp = self.client.replay_mesh_dead_letter(entry_id).await.map_err(agentsmesh_services::wire)?;
        serde_json::to_string(&resp).map_err(agentsmesh_services::wire)
    }
}
