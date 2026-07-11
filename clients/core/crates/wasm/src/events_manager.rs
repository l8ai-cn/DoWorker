use std::sync::Arc;

use agentsmesh_api_client::ApiClient;
use agentsmesh_events::{
    EventSubscriptionManager, EventSubscriptionManagerOptions, EventType, SubscriptionId,
};
use agentsmesh_transport::runtime::PlatformRuntime;
use wasm_bindgen::prelude::*;

use crate::js_bridge::{make_event_handler, make_state_listener};

/// Events manager exposed to JavaScript via wasm-bindgen.
///
/// Owns an `EventSubscriptionManager` backed by the shared `ApiClient`
/// — Connect server-streaming over HTTP, NOT WebSocket (see R5-11). The
/// `ws_url` constructor parameter is gone; the auth token, base URL,
/// and org slug all come from the `ApiClient` / auth store.
///
/// Backed by `Arc<EventSubscriptionManager>` (not `Rc<RefCell<...>>`) so
/// the same manager can be shared with `AppRuntime` — its `connect` /
/// `disconnect` methods take `&self` and serialize mutation through the
/// internal `Arc<RwLock<Inner>>`.
#[wasm_bindgen]
pub struct WasmEventsManager {
    inner: Arc<EventSubscriptionManager<PlatformRuntime>>,
}

impl WasmEventsManager {
    pub(crate) fn new_internal(client: Arc<ApiClient>) -> Self {
        let manager = EventSubscriptionManager::with_runtime(
            PlatformRuntime,
            client,
            EventSubscriptionManagerOptions::default(),
        );
        Self {
            inner: Arc::new(manager),
        }
    }

    pub(crate) fn new_internal_with_options(
        client: Arc<ApiClient>,
        options: EventSubscriptionManagerOptions,
    ) -> Self {
        let manager = EventSubscriptionManager::with_runtime(PlatformRuntime, client, options);
        Self {
            inner: Arc::new(manager),
        }
    }

    /// Wrap an externally-constructed shared `EventSubscriptionManager`
    /// (the one owned by `AppRuntime`). This is the path
    /// `WasmApiClient::create_events_manager` takes — single manager
    /// instance, single dispatch hook.
    pub(crate) fn from_shared(manager: Arc<EventSubscriptionManager<PlatformRuntime>>) -> Self {
        Self { inner: manager }
    }

    /// Internal accessor for binding facades that need to hand the same
    /// manager to `AppRuntime` for dispatch-hook installation.
    pub(crate) fn manager_arc(&self) -> Arc<EventSubscriptionManager<PlatformRuntime>> {
        Arc::clone(&self.inner)
    }
}

#[wasm_bindgen]
impl WasmEventsManager {
    pub async fn connect(&self) {
        self.inner.connect().await;
    }

    pub async fn disconnect(&self) {
        self.inner.disconnect().await;
    }

    /// Interrupt the reconnect backoff and retry now (network regained / tab
    /// refocused). No-op when already connected or shut down.
    pub async fn nudge(&self) {
        self.inner.nudge().await;
    }

    pub async fn subscribe(
        &self,
        event_type: String,
        callback: js_sys::Function,
    ) -> Result<f64, String> {
        let et = parse_event_type(&event_type)?;
        let handler = make_event_handler(callback);
        let id = self.inner.subscribe(et, handler).await;
        Ok(sub_id_to_f64(id))
    }

    pub async fn subscribe_all(&self, callback: js_sys::Function) -> f64 {
        let handler = make_event_handler(callback);
        let id = self.inner.subscribe_all(handler).await;
        sub_id_to_f64(id)
    }

    pub async fn unsubscribe(&self, id: f64) {
        let sid = f64_to_sub_id(id);
        self.inner.unsubscribe(sid).await;
    }

    pub async fn on_connection_state_change(&self, callback: js_sys::Function) -> f64 {
        let listener = make_state_listener(callback);
        let id = self.inner.on_connection_state_change(listener).await;
        sub_id_to_f64(id)
    }

    pub async fn get_connection_state(&self) -> String {
        self.inner.get_connection_state().await.to_string()
    }

    /// Snapshot of the dispatch tick — increments after every event has
    /// been applied to AppState. React selectors read this to decide whether
    /// to re-derive cached views.
    pub fn tick(&self) -> f64 {
        self.inner.tick() as f64
    }
}

fn sub_id_to_f64(id: SubscriptionId) -> f64 {
    id.as_u64() as f64
}

fn f64_to_sub_id(id: f64) -> SubscriptionId {
    SubscriptionId::from_u64(id as u64)
}

fn parse_event_type(s: &str) -> Result<EventType, String> {
    let parsed: std::result::Result<EventType, _> =
        serde_json::from_value(serde_json::Value::String(s.to_string()));
    parsed.map_err(|e| format!("unknown event type '{s}': {e}"))
}
