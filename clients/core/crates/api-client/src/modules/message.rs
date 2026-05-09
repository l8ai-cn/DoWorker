use std::collections::HashMap;

use crate::ApiClient;
use crate::error::ApiError;
use crate::request::RequestOptions;
use agentsmesh_types::*;
use reqwest::Method;

impl ApiClient {
    pub async fn send_mesh_message(
        &self,
        data: &SendDirectMessageRequest,
        pod_key: Option<&str>,
    ) -> Result<DirectMessage, ApiError> {
        let mut opts = RequestOptions {
            body: Some(serde_json::to_value(data)?),
            ..Default::default()
        };
        if let Some(key) = pod_key {
            let mut headers = HashMap::new();
            headers.insert("X-Pod-Key".to_string(), key.to_string());
            opts.headers = Some(headers);
        }
        self.request(Method::POST, &self.org_path("/messages"), opts)
            .await
    }

    pub async fn get_mesh_messages(
        &self,
        unread_only: Option<bool>,
        limit: Option<u32>,
        offset: Option<u32>,
    ) -> Result<DirectMessageListResponse, ApiError> {
        let mut path = self.org_path("/messages");
        let mut params = Vec::new();
        if let Some(u) = unread_only {
            params.push(format!("unread_only={u}"));
        }
        if let Some(l) = limit {
            params.push(format!("limit={l}"));
        }
        if let Some(o) = offset {
            params.push(format!("offset={o}"));
        }
        if !params.is_empty() {
            path = format!("{path}?{}", params.join("&"));
        }
        self.get(&path).await
    }

    pub async fn get_mesh_unread_count(&self) -> Result<UnreadCountResponse, ApiError> {
        self.get(&self.org_path("/messages/unread-count")).await
    }

    pub async fn get_mesh_message(&self, id: i64) -> Result<DirectMessage, ApiError> {
        self.get_resource(&self.org_path(&format!("/messages/{id}")), "message").await
    }

    pub async fn mark_mesh_messages_read(
        &self,
        data: &MarkMessagesReadRequest,
    ) -> Result<MarkReadResponse, ApiError> {
        self.post(&self.org_path("/messages/mark-read"), data)
            .await
    }

    pub async fn mark_all_mesh_messages_read(&self) -> Result<MarkReadResponse, ApiError> {
        self.post(
            &self.org_path("/messages/mark-all-read"),
            &serde_json::json!({}),
        )
        .await
    }

    pub async fn get_mesh_conversation(
        &self,
        correlation_id: &str,
        limit: Option<u32>,
    ) -> Result<DirectMessageListResponse, ApiError> {
        let mut path = self.org_path(&format!("/messages/conversation/{correlation_id}"));
        if let Some(l) = limit {
            path = format!("{path}?limit={l}");
        }
        self.get(&path).await
    }

    pub async fn get_mesh_sent_messages(
        &self,
        limit: Option<u32>,
        offset: Option<u32>,
    ) -> Result<DirectMessageListResponse, ApiError> {
        let mut path = self.org_path("/messages/sent");
        let mut params = Vec::new();
        if let Some(l) = limit {
            params.push(format!("limit={l}"));
        }
        if let Some(o) = offset {
            params.push(format!("offset={o}"));
        }
        if !params.is_empty() {
            path = format!("{path}?{}", params.join("&"));
        }
        self.get(&path).await
    }

    pub async fn get_mesh_dead_letters(
        &self,
        limit: Option<u32>,
        offset: Option<u32>,
    ) -> Result<DeadLetterListResponse, ApiError> {
        let mut path = self.org_path("/messages/dlq");
        let mut params = Vec::new();
        if let Some(l) = limit {
            params.push(format!("limit={l}"));
        }
        if let Some(o) = offset {
            params.push(format!("offset={o}"));
        }
        if !params.is_empty() {
            path = format!("{path}?{}", params.join("&"));
        }
        self.get(&path).await
    }

    pub async fn replay_mesh_dead_letter(
        &self,
        entry_id: i64,
    ) -> Result<ReplayDeadLetterResponse, ApiError> {
        self.post(
            &self.org_path(&format!("/messages/dlq/{entry_id}/replay")),
            &serde_json::json!({}),
        )
        .await
    }
}
