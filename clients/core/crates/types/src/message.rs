use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DirectMessage {
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

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SendDirectMessageRequest {
    pub receiver_pod: String,
    pub message_type: Option<String>,
    pub content: String,
    pub correlation_id: Option<String>,
    pub reply_to_id: Option<i64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MarkMessagesReadRequest {
    pub message_ids: Vec<i64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DirectMessageListResponse {
    pub messages: Vec<DirectMessage>,
    pub total: Option<i64>,
    #[serde(default)]
    pub unread_count: Option<i64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UnreadCountResponse {
    pub count: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeadLetterEntry {
    pub id: i64,
    pub message: Option<DirectMessage>,
    pub error: Option<String>,
    pub created_at: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeadLetterListResponse {
    pub entries: Vec<DeadLetterEntry>,
    #[serde(default)]
    pub total: Option<i64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReplayDeadLetterResponse {
    #[serde(default)]
    pub message: Option<String>,
    pub replayed_message: Option<DirectMessage>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MarkReadResponse {
    pub marked_count: i64,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn direct_message_list_relay_preserves_unread_count() {
        let backend = r#"{
            "messages": [{"id":1}],
            "total": 1,
            "unread_count": 7
        }"#;
        let typed: DirectMessageListResponse = serde_json::from_str(backend).unwrap();
        let relayed = serde_json::to_string(&typed).unwrap();
        let parsed: serde_json::Value = serde_json::from_str(&relayed).unwrap();
        assert_eq!(parsed["unread_count"], serde_json::json!(7));
        assert_eq!(parsed["total"], serde_json::json!(1));
    }

    #[test]
    fn dead_letter_list_relay_preserves_total() {
        let backend = r#"{"entries":[],"total":3}"#;
        let typed: DeadLetterListResponse = serde_json::from_str(backend).unwrap();
        let relayed = serde_json::to_string(&typed).unwrap();
        let parsed: serde_json::Value = serde_json::from_str(&relayed).unwrap();
        assert_eq!(parsed["total"], serde_json::json!(3));
    }

    #[test]
    fn replay_dead_letter_relay_preserves_replayed_message() {
        let backend = r#"{
            "message": "Replayed successfully",
            "replayed_message": {"id": 88, "sender_pod": "p1"}
        }"#;
        let typed: ReplayDeadLetterResponse = serde_json::from_str(backend).unwrap();
        let relayed = serde_json::to_string(&typed).unwrap();
        let parsed: serde_json::Value = serde_json::from_str(&relayed).unwrap();
        assert_eq!(parsed["message"], serde_json::json!("Replayed successfully"));
        assert_eq!(parsed["replayed_message"]["id"], serde_json::json!(88));
    }
}
