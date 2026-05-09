use crate::ApiClient;
use crate::error::ApiError;
use agentsmesh_types::*;

impl ApiClient {
    pub async fn list_support_tickets(
        &self,
        status: Option<&str>,
        page: Option<u32>,
        page_size: Option<u32>,
    ) -> Result<SupportTicketListResponse, ApiError> {
        let mut path = "/api/v1/support-tickets".to_string();
        let mut params = Vec::new();
        if let Some(s) = status {
            params.push(format!("status={s}"));
        }
        if let Some(p) = page {
            params.push(format!("page={p}"));
        }
        if let Some(ps) = page_size {
            params.push(format!("page_size={ps}"));
        }
        if !params.is_empty() {
            path = format!("{path}?{}", params.join("&"));
        }
        self.get(&path).await
    }

    pub async fn get_support_ticket_detail(
        &self,
        id: i64,
    ) -> Result<SupportTicketDetailResponse, ApiError> {
        self.get(&format!("/api/v1/support-tickets/{id}")).await
    }

    pub async fn get_support_ticket_attachment_url(
        &self,
        attachment_id: i64,
    ) -> Result<AttachmentUrlResponse, ApiError> {
        self.get_resource(
            &format!("/api/v1/support-tickets/attachments/{attachment_id}/url"),
            "url",
        ).await
    }
}
