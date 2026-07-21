use std::sync::Arc;

use agentcloud_api_client::ApiClient;
use agentcloud_services::TicketService;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmTicketService(pub(crate) TicketService);

#[wasm_bindgen]
impl WasmTicketService {
    pub(crate) fn new(client: Arc<ApiClient>) -> Self {
        Self(TicketService::new(client))
    }

    // -------- Ticket→Pod lookup (MeshService domain). Fetch-only now: the
    // result is mirrored into runtime.state via WasmTicketState.set_ticket_pods
    // by the caller (useTicketPods), so the cache is the shared SSOT.

    pub async fn get_ticket_pods(
        &self,
        slug: &str,
        active_only: Option<bool>,
    ) -> Result<String, String> {
        self.0.get_ticket_pods(slug, active_only).await
    }

    // -------- Connect-RPC (binary wire) --------
    //
    // Each `*_connect` method takes prost-encoded bytes (Uint8Array on the JS
    // side) and returns prost-encoded bytes — TS callers encode via
    // @bufbuild/protobuf .toBinary() and decode via .fromBinary().

    pub async fn list_tickets_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_tickets_connect(request_bytes).await
    }

    pub async fn get_ticket_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_ticket_connect(request_bytes).await
    }

    pub async fn create_ticket_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.create_ticket_connect(request_bytes).await
    }

    pub async fn update_ticket_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.update_ticket_connect(request_bytes).await
    }

    pub async fn delete_ticket_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.delete_ticket_connect(request_bytes).await
    }

    pub async fn update_ticket_status_connect(
        &self,
        request_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0.update_ticket_status_connect(request_bytes).await
    }

    pub async fn get_active_tickets_connect(
        &self,
        request_bytes: &[u8],
    ) -> Result<Vec<u8>, String> {
        self.0.get_active_tickets_connect(request_bytes).await
    }

    pub async fn get_board_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_board_connect(request_bytes).await
    }

    pub async fn get_sub_tickets_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_sub_tickets_connect(request_bytes).await
    }

    pub async fn add_assignee_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.add_assignee_connect(request_bytes).await
    }

    pub async fn remove_assignee_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.remove_assignee_connect(request_bytes).await
    }

    pub async fn list_labels_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_labels_connect(request_bytes).await
    }

    pub async fn create_label_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.create_label_connect(request_bytes).await
    }

    pub async fn update_label_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.update_label_connect(request_bytes).await
    }

    pub async fn delete_label_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.delete_label_connect(request_bytes).await
    }

    pub async fn add_label_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.add_label_connect(request_bytes).await
    }

    pub async fn remove_label_connect(&self, request_bytes: &[u8]) -> Result<Vec<u8>, String> {
        self.0.remove_label_connect(request_bytes).await
    }
}
