use crate::core::AgentsMeshCore;
use crate::dto::{
    mark_messages_read_req, DeadLetterListResponseDto, DirectMessageDto,
    DirectMessageListResponseDto, ReplayDeadLetterResponseDto, SendDirectMessageRequestDto,
    UnreadCountResponseDto,
};
use crate::error::CoreError;

/// Mesh direct-messaging (pod-to-pod): send/receive structured messages
/// across pods within an org. Distinct from channel messaging — DM is
/// addressed to a `receiver_pod`, no shared topic.
#[uniffi::export(async_runtime = "tokio")]
impl AgentsMeshCore {
    pub async fn send_mesh_message(
        &self,
        req: SendDirectMessageRequestDto,
        sender_pod_key: Option<String>,
    ) -> Result<DirectMessageDto, CoreError> {
        let msg = self
            .api
            .send_mesh_message(&req.into(), sender_pod_key.as_deref())
            .await?;
        Ok(msg.into())
    }

    pub async fn get_mesh_messages(
        &self,
        unread_only: Option<bool>,
        limit: Option<u32>,
        offset: Option<u32>,
    ) -> Result<DirectMessageListResponseDto, CoreError> {
        let resp = self.api.get_mesh_messages(unread_only, limit, offset).await?;
        Ok(resp.into())
    }

    pub async fn get_mesh_unread_count(&self) -> Result<UnreadCountResponseDto, CoreError> {
        let resp = self.api.get_mesh_unread_count().await?;
        Ok(resp.into())
    }

    pub async fn get_mesh_message(&self, id: i64) -> Result<DirectMessageDto, CoreError> {
        let msg = self.api.get_mesh_message(id).await?;
        Ok(msg.into())
    }

    pub async fn mark_mesh_messages_read(&self, message_ids: Vec<i64>) -> Result<i64, CoreError> {
        let resp = self.api
            .mark_mesh_messages_read(&mark_messages_read_req(message_ids))
            .await?;
        Ok(resp.marked_count)
    }

    pub async fn mark_all_mesh_messages_read(&self) -> Result<i64, CoreError> {
        let resp = self.api.mark_all_mesh_messages_read().await?;
        Ok(resp.marked_count)
    }

    pub async fn get_mesh_conversation(
        &self,
        correlation_id: String,
        limit: Option<u32>,
    ) -> Result<DirectMessageListResponseDto, CoreError> {
        let resp = self.api.get_mesh_conversation(&correlation_id, limit).await?;
        Ok(resp.into())
    }

    pub async fn get_mesh_sent_messages(
        &self,
        limit: Option<u32>,
        offset: Option<u32>,
    ) -> Result<DirectMessageListResponseDto, CoreError> {
        let resp = self.api.get_mesh_sent_messages(limit, offset).await?;
        Ok(resp.into())
    }

    pub async fn get_mesh_dead_letters(
        &self,
        limit: Option<u32>,
        offset: Option<u32>,
    ) -> Result<DeadLetterListResponseDto, CoreError> {
        let resp = self.api.get_mesh_dead_letters(limit, offset).await?;
        Ok(resp.into())
    }

    pub async fn replay_mesh_dead_letter(
        &self,
        entry_id: i64,
    ) -> Result<ReplayDeadLetterResponseDto, CoreError> {
        let resp = self.api.replay_mesh_dead_letter(entry_id).await?;
        Ok(resp.into())
    }
}
