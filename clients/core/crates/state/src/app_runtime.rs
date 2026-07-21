use std::sync::Arc;

use agentcloud_events::subscription_manager::EventSubscriptionManager;
use agentcloud_events::types::{EventDispatchHook, RealtimeEvent};
use parking_lot::RwLock;

use crate::app_state::AppState;

pub struct AppRuntime {
    pub state: Arc<RwLock<AppState>>,
    pub events: Arc<EventSubscriptionManager>,
}

impl AppRuntime {
    pub fn new(events: Arc<EventSubscriptionManager>) -> Arc<Self> {
        Self::with_state(events, AppState::new())
    }

    pub fn with_state(events: Arc<EventSubscriptionManager>, state: AppState) -> Arc<Self> {
        let state = Arc::new(RwLock::new(state));
        let hook: Arc<dyn EventDispatchHook> =
            Arc::new(AppStateDispatchHook::new(Arc::clone(&state)));
        events.set_dispatch_hook(hook);
        Arc::new(Self { state, events })
    }

    pub fn tick(&self) -> u64 {
        self.events.tick()
    }
}

pub struct AppStateDispatchHook {
    state: Arc<RwLock<AppState>>,
}

impl AppStateDispatchHook {
    pub fn new(state: Arc<RwLock<AppState>>) -> Self {
        Self { state }
    }
}

impl EventDispatchHook for AppStateDispatchHook {
    fn dispatch(&self, event: &RealtimeEvent) {
        self.state.write().dispatch(event);
    }
}
