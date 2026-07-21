use std::sync::Arc;

use agentcloud_api_client::ApiClient;
use agentcloud_services::ChannelService;
use wasm_bindgen::prelude::*;

// Networking-only wasm handle for the channel domain. The channel cache lives
// in the shared `AppState.channels` (reached via `WasmChannelState`); this
// service exposes only the `*Connect` surface (service_channel_connect.rs).
#[wasm_bindgen]
pub struct WasmChannelService(pub(crate) ChannelService);

impl WasmChannelService {
    pub(crate) fn new(client: Arc<ApiClient>) -> Self {
        Self(ChannelService::new(client))
    }
}
