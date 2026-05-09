use serde::{Deserialize, Serialize};

use crate::{TicketComment, TicketCommit, TicketPriority, TicketRelation, TicketStatus};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateTicketRequest {
    pub title: String,
    pub content: Option<String>,
    pub priority: Option<TicketPriority>,
    pub severity: Option<String>,
    pub estimate: Option<f64>,
    pub repository_id: Option<i64>,
    pub assignee_ids: Option<Vec<i64>>,
    pub labels: Option<Vec<i64>>,
    pub parent_slug: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateTicketRequest {
    pub title: Option<String>,
    pub content: Option<String>,
    pub priority: Option<TicketPriority>,
    pub severity: Option<String>,
    pub estimate: Option<f64>,
    pub repository_id: Option<i64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateTicketStatusRequest {
    pub status: TicketStatus,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateTicketCommentRequest {
    pub content: String,
    pub parent_id: Option<i64>,
    pub mentions: Option<Vec<String>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateTicketCommentRequest {
    pub content: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateTicketRelationRequest {
    pub target_slug: String,
    pub relation_type: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LinkTicketCommitRequest {
    pub commit_sha: String,
    pub commit_message: Option<String>,
    pub commit_url: Option<String>,
    pub author_name: Option<String>,
    pub author_email: Option<String>,
    pub committed_at: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AddAssigneeRequest {
    pub user_id: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AddTicketLabelRequest {
    pub label_id: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateLabelRequest {
    pub name: String,
    pub color: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateLabelRequest {
    pub name: Option<String>,
    pub color: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateTicketPodRequest {
    pub runner_id: Option<i64>,
    pub prompt: Option<String>,
    pub model: Option<String>,
    pub permission_mode: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BatchPodRequest {
    pub ticket_slugs: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TicketCommentListResponse {
    pub comments: Vec<TicketComment>,
    pub total: Option<i64>,
    #[serde(default)]
    pub limit: Option<i64>,
    #[serde(default)]
    pub offset: Option<i64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TicketRelationListResponse {
    pub relations: Vec<TicketRelation>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TicketCommitListResponse {
    pub commits: Vec<TicketCommit>,
}

#[cfg(test)]
mod pagination_tests {
    use super::*;

    #[test]
    fn comment_list_relay_preserves_pagination() {
        let backend = r#"{"comments":[],"total":13,"limit":50,"offset":0}"#;
        let typed: TicketCommentListResponse = serde_json::from_str(backend).unwrap();
        let relayed = serde_json::to_string(&typed).unwrap();
        let parsed: serde_json::Value = serde_json::from_str(&relayed).unwrap();
        assert_eq!(parsed["total"], serde_json::json!(13));
        assert_eq!(parsed["limit"], serde_json::json!(50));
        assert_eq!(parsed["offset"], serde_json::json!(0));
    }
}
