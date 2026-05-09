use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SupportTicket {
    pub id: i64,
    pub title: String,
    pub category: Option<String>,
    pub content: Option<String>,
    pub priority: Option<String>,
    pub status: Option<String>,
    pub created_at: Option<String>,
    pub updated_at: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SupportTicketMessage {
    pub id: i64,
    pub ticket_id: i64,
    pub content: Option<String>,
    pub sender_type: Option<String>,
    pub attachments: Option<Vec<String>>,
    pub created_at: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SupportTicketListResponse {
    /// Backend returns `data` (gin paginated convention); the field is kept
    /// public as `tickets` for ergonomics in Rust callers, but the wire name
    /// stays `data` so both frontend types and the Go handler agree.
    #[serde(rename = "data")]
    pub tickets: Vec<SupportTicket>,
    #[serde(default)]
    pub total: Option<i64>,
    #[serde(default)]
    pub page: Option<i32>,
    #[serde(default, rename = "page_size")]
    pub page_size: Option<i32>,
    #[serde(default, rename = "total_pages")]
    pub total_pages: Option<i32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SupportTicketDetailResponse {
    pub ticket: SupportTicket,
    #[serde(default)]
    pub messages: Vec<SupportTicketMessage>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AttachmentUrlResponse {
    pub url: String,
}
