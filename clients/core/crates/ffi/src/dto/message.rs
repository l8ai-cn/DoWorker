use agentsmesh_types::{
    DeadLetterEntry, DeadLetterListResponse, DirectMessage, DirectMessageListResponse,
    MarkMessagesReadRequest, ReplayDeadLetterResponse, SendDirectMessageRequest,
    UnreadCountResponse,
};

#[derive(Clone, Debug, uniffi::Record)]
pub struct DirectMessageDto {
    pub id: i64,
    pub sender_pod: Option<String>,
    pub receiver_pod: Option<String>,
    pub message_type: Option<String>,
    pub content: Option<String>,
    pub correlation_id: Option<String>,
    pub reply_to_id: Option<i64>,
    pub is_read: Option<bool>,
    pub created_at: Option<String>,
}

impl From<DirectMessage> for DirectMessageDto {
    fn from(m: DirectMessage) -> Self {
        Self {
            id: m.id,
            sender_pod: m.sender_pod,
            receiver_pod: m.receiver_pod,
            message_type: m.message_type,
            content: m.content,
            correlation_id: m.correlation_id,
            reply_to_id: m.reply_to_id,
            is_read: m.is_read,
            created_at: m.created_at,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct SendDirectMessageRequestDto {
    pub receiver_pod: String,
    pub message_type: Option<String>,
    pub content: String,
    pub correlation_id: Option<String>,
    pub reply_to_id: Option<i64>,
}

impl From<SendDirectMessageRequestDto> for SendDirectMessageRequest {
    fn from(d: SendDirectMessageRequestDto) -> Self {
        Self {
            receiver_pod: d.receiver_pod,
            message_type: d.message_type,
            content: d.content,
            correlation_id: d.correlation_id,
            reply_to_id: d.reply_to_id,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct DirectMessageListResponseDto {
    pub messages: Vec<DirectMessageDto>,
    pub total: Option<i64>,
    pub unread_count: Option<i64>,
}

impl From<DirectMessageListResponse> for DirectMessageListResponseDto {
    fn from(r: DirectMessageListResponse) -> Self {
        Self {
            messages: r.messages.into_iter().map(DirectMessageDto::from).collect(),
            total: r.total,
            unread_count: r.unread_count,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct UnreadCountResponseDto {
    pub count: i64,
}

impl From<UnreadCountResponse> for UnreadCountResponseDto {
    fn from(r: UnreadCountResponse) -> Self {
        Self { count: r.count }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct DeadLetterEntryDto {
    pub id: i64,
    pub message: Option<DirectMessageDto>,
    pub error: Option<String>,
    pub created_at: Option<String>,
}

impl From<DeadLetterEntry> for DeadLetterEntryDto {
    fn from(e: DeadLetterEntry) -> Self {
        Self {
            id: e.id,
            message: e.message.map(Into::into),
            error: e.error,
            created_at: e.created_at,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct DeadLetterListResponseDto {
    pub entries: Vec<DeadLetterEntryDto>,
    pub total: Option<i64>,
}

impl From<DeadLetterListResponse> for DeadLetterListResponseDto {
    fn from(r: DeadLetterListResponse) -> Self {
        Self {
            entries: r.entries.into_iter().map(DeadLetterEntryDto::from).collect(),
            total: r.total,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct ReplayDeadLetterResponseDto {
    pub message: Option<String>,
    pub replayed_message: Option<DirectMessageDto>,
}

impl From<ReplayDeadLetterResponse> for ReplayDeadLetterResponseDto {
    fn from(r: ReplayDeadLetterResponse) -> Self {
        Self {
            message: r.message,
            replayed_message: r.replayed_message.map(Into::into),
        }
    }
}

pub(crate) fn mark_messages_read_req(message_ids: Vec<i64>) -> MarkMessagesReadRequest {
    MarkMessagesReadRequest { message_ids }
}
