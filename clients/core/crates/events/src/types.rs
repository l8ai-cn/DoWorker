use std::sync::Arc;

use serde::{Deserialize, Serialize};

use crate::event_types::EventType;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ConnectionState {
    Disconnected,
    Connecting,
    Connected,
    Reconnecting,
}

impl std::fmt::Display for ConnectionState {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Disconnected => f.write_str("disconnected"),
            Self::Connecting => f.write_str("connecting"),
            Self::Connected => f.write_str("connected"),
            Self::Reconnecting => f.write_str("reconnecting"),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum EventCategory {
    Entity,
    Notification,
    System,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RealtimeEvent {
    #[serde(rename = "type")]
    pub event_type: EventType,
    #[serde(default)]
    pub category: Option<EventCategory>,
    #[serde(default)]
    pub organization_id: i64,
    pub target_user_id: Option<i64>,
    pub target_user_ids: Option<Vec<i64>>,
    pub entity_type: Option<String>,
    pub entity_id: Option<String>,
    #[serde(default)]
    pub data: serde_json::Value,
    #[serde(default)]
    pub timestamp: i64,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct SubscriptionId(pub(crate) u64);

impl SubscriptionId {
    pub fn as_u64(self) -> u64 {
        self.0
    }

    pub fn from_u64(val: u64) -> Self {
        Self(val)
    }
}

pub type EventHandler = Arc<dyn Fn(&RealtimeEvent) + Send + Sync>;
pub type StateListener = Arc<dyn Fn(ConnectionState) + Send + Sync>;

/// Hook invoked synchronously inside `dispatch_event` BEFORE external
/// handlers and BEFORE the tick bump. Implementations (in `state` crate)
/// apply the event to the in-memory `AppState`. Keeping this in the events
/// crate lets `state` impl the trait without inverting the dep direction.
///
/// Contract: implementations MUST NOT call back into
/// `EventSubscriptionManager`. They may acquire their own locks but must
/// drop them before returning. The dispatcher holds no events-side lock
/// while calling the hook.
pub trait EventDispatchHook: Send + Sync {
    fn dispatch(&self, event: &RealtimeEvent);
}

/// Fires once per dispatched realtime event AFTER the AppState mutation
/// + tick increment. Native host bindings can use this to push a "state may
/// have changed, re-read selectors" signal without piping the raw event JSON.
///
/// Wasm platforms poll the tick via getter instead — push only
/// matters where the platform's reactive system needs an explicit kick.
pub trait TickListener: Send + Sync {
    fn on_tick(&self, tick: u64);
}

pub struct EventSubscriptionManagerOptions {
    pub initial_reconnect_delay_ms: u64,
    pub max_reconnect_delay_ms: u64,
    /// No inbound events for this long → treat the stream as stalled and
    /// reconnect (0 = use the loop's built-in default). Replaces the legacy
    /// ping/pong knobs: Connect server-streaming has no app-level ping, the
    /// idle gap is the "still receiving?" signal.
    pub idle_timeout_ms: u64,
    /// Cap a single connect attempt (0 = loop default). A hung handshake has
    /// no OS-level timeout here; without this the loop blocks forever.
    pub connect_timeout_ms: u64,
}

impl Default for EventSubscriptionManagerOptions {
    fn default() -> Self {
        Self {
            initial_reconnect_delay_ms: 1000,
            max_reconnect_delay_ms: 30000,
            idle_timeout_ms: 60000,
            connect_timeout_ms: 15000,
        }
    }
}
