use std::sync::atomic::{AtomicU32, AtomicU64, Ordering};
use std::sync::Arc;

use agentcloud_api_client::{ApiClient, AuthTokenStore};

use crate::{
    ConnectionState, EventHandler, EventSubscriptionManager, EventSubscriptionManagerOptions,
    EventType, RealtimeEvent, StateListener, SubscriptionId,
};
use crate::types::{EventDispatchHook, TickListener};

struct StubAuth;
impl AuthTokenStore for StubAuth {
    fn get_token(&self) -> Option<String> {
        Some("test-token".into())
    }
    fn get_refresh_token(&self) -> Option<String> {
        None
    }
    fn set_tokens(&self, _t: String, _r: String, _e: Option<i64>) {}
    fn clear_tokens(&self) {}
    fn get_current_org_slug(&self) -> Option<String> {
        Some("test-org".into())
    }
}

fn make_client() -> Arc<ApiClient> {
    Arc::new(ApiClient::new(
        "http://localhost:9999".into(),
        Arc::new(StubAuth),
    ))
}

fn make_event(event_type: EventType) -> RealtimeEvent {
    serde_json::from_value(serde_json::json!({
        "type": event_type.as_str(),
        "category": "entity",
        "organization_id": 1,
        "data": {},
        "timestamp": 1000
    }))
    .unwrap()
}

#[tokio::test]
async fn test_initial_state_is_disconnected() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    assert_eq!(mgr.get_connection_state().await, ConnectionState::Disconnected);
}

#[tokio::test]
async fn test_subscribe_and_unsubscribe() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let counter = Arc::new(AtomicU32::new(0));
    let c = Arc::clone(&counter);
    let handler: EventHandler = Arc::new(move |_| {
        c.fetch_add(1, Ordering::SeqCst);
    });
    let id = mgr.subscribe(EventType::PodCreated, handler).await;
    mgr.unsubscribe(id).await;
}

#[tokio::test]
async fn test_subscribe_all_and_unsubscribe() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let counter = Arc::new(AtomicU32::new(0));
    let c = Arc::clone(&counter);
    let handler: EventHandler = Arc::new(move |_| {
        c.fetch_add(1, Ordering::SeqCst);
    });
    let id = mgr.subscribe_all(handler).await;
    mgr.unsubscribe(id).await;
}

#[tokio::test]
async fn test_connection_state_listener_gets_initial() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let received = Arc::new(std::sync::Mutex::new(Vec::new()));
    let r = Arc::clone(&received);
    let listener: StateListener = Arc::new(move |state| {
        r.lock().unwrap().push(state);
    });
    let _id = mgr.on_connection_state_change(listener).await;
    tokio::time::sleep(std::time::Duration::from_millis(10)).await;
    let states = received.lock().unwrap();
    assert_eq!(states.len(), 1);
    assert_eq!(states[0], ConnectionState::Disconnected);
}

#[tokio::test]
async fn test_multiple_subscriptions_different_types() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let h1: EventHandler = Arc::new(|_| {});
    let h2: EventHandler = Arc::new(|_| {});
    let id1 = mgr.subscribe(EventType::PodCreated, h1).await;
    let id2 = mgr.subscribe(EventType::RunnerOnline, h2).await;
    assert_ne!(id1, id2);
    mgr.unsubscribe(id1).await;
    mgr.unsubscribe(id2).await;
}

#[tokio::test]
async fn test_unsubscribe_nonexistent_id_is_noop() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    mgr.unsubscribe(SubscriptionId::from_u64(99999)).await;
}

#[tokio::test]
async fn test_options_default() {
    let opts = EventSubscriptionManagerOptions::default();
    assert_eq!(opts.initial_reconnect_delay_ms, 1000);
    assert_eq!(opts.max_reconnect_delay_ms, 30000);
    assert_eq!(opts.idle_timeout_ms, 60000);
    assert_eq!(opts.connect_timeout_ms, 15000);
}

#[tokio::test]
async fn test_custom_options() {
    let opts = EventSubscriptionManagerOptions {
        initial_reconnect_delay_ms: 500,
        max_reconnect_delay_ms: 15000,
        idle_timeout_ms: 5000,
        connect_timeout_ms: 2000,
    };
    let mgr = EventSubscriptionManager::new(make_client(), opts);
    assert_eq!(mgr.get_connection_state().await, ConnectionState::Disconnected);
}

#[test]
fn test_realtime_event_deserialization_variations() {
    let json = r#"{
        "type": "notification",
        "category": "notification",
        "organization_id": 5,
        "target_user_id": 42,
        "data": {"title": "Hello"},
        "timestamp": 2000
    }"#;
    let event: RealtimeEvent = serde_json::from_str(json).unwrap();
    assert_eq!(event.event_type, EventType::Notification);
    assert_eq!(event.target_user_id, Some(42));
    assert!(event.entity_type.is_none());

    let json = r#"{
        "type": "system:maintenance",
        "category": "system",
        "organization_id": 0,
        "data": {"message": "downtime"},
        "timestamp": 3000
    }"#;
    let event: RealtimeEvent = serde_json::from_str(json).unwrap();
    assert_eq!(event.event_type, EventType::SystemMaintenance);
}

#[tokio::test]
async fn test_disconnect_without_connect() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    mgr.disconnect().await;
    assert_eq!(mgr.get_connection_state().await, ConnectionState::Disconnected);
}

#[test]
fn test_make_event_pod_created() {
    let event = make_event(EventType::PodCreated);
    assert_eq!(event.event_type, EventType::PodCreated);
    assert_eq!(event.organization_id, 1);
    assert_eq!(event.timestamp, 1000);
}

#[tokio::test]
async fn test_subscribe_receives_dispatched_event() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let counter = Arc::new(AtomicU32::new(0));
    let c = counter.clone();
    let handler: EventHandler = Arc::new(move |evt| {
        assert_eq!(evt.event_type, EventType::PodCreated);
        c.fetch_add(1, Ordering::SeqCst);
    });
    let _id = mgr.subscribe(EventType::PodCreated, handler).await;
    let event = make_event(EventType::PodCreated);
    crate::subscription_manager::dispatch_event(&mgr.inner, &event);
    assert_eq!(counter.load(Ordering::SeqCst), 1);
}

#[tokio::test]
async fn test_subscribe_all_receives_any_event() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let counter = Arc::new(AtomicU32::new(0));
    let c = counter.clone();
    let handler: EventHandler = Arc::new(move |_| {
        c.fetch_add(1, Ordering::SeqCst);
    });
    let _id = mgr.subscribe_all(handler).await;
    let evt1 = make_event(EventType::PodCreated);
    let evt2 = make_event(EventType::RunnerOnline);
    crate::subscription_manager::dispatch_event(&mgr.inner, &evt1);
    crate::subscription_manager::dispatch_event(&mgr.inner, &evt2);
    assert_eq!(counter.load(Ordering::SeqCst), 2);
}

#[tokio::test]
async fn test_unsubscribe_stops_delivery() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let counter = Arc::new(AtomicU32::new(0));
    let c = counter.clone();
    let handler: EventHandler = Arc::new(move |_| {
        c.fetch_add(1, Ordering::SeqCst);
    });
    let id = mgr.subscribe(EventType::PodCreated, handler).await;
    mgr.unsubscribe(id).await;
    let event = make_event(EventType::PodCreated);
    crate::subscription_manager::dispatch_event(&mgr.inner, &event);
    assert_eq!(counter.load(Ordering::SeqCst), 0);
}

// ---- Phase 1: Rust-SSOT dispatch hook + tick ----

struct CountingHook {
    counter: Arc<AtomicU32>,
}

impl EventDispatchHook for CountingHook {
    fn dispatch(&self, _event: &RealtimeEvent) {
        self.counter.fetch_add(1, Ordering::SeqCst);
    }
}

struct PanickingHook;

impl EventDispatchHook for PanickingHook {
    fn dispatch(&self, _event: &RealtimeEvent) {
        panic!("boom");
    }
}

#[tokio::test]
async fn dispatch_hook_runs_before_external_handlers() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let order = Arc::new(std::sync::Mutex::new(Vec::<&'static str>::new()));

    // Hook records "hook" first
    let o = order.clone();
    struct OrderingHook(Arc<std::sync::Mutex<Vec<&'static str>>>);
    impl EventDispatchHook for OrderingHook {
        fn dispatch(&self, _event: &RealtimeEvent) {
            self.0.lock().unwrap().push("hook");
        }
    }
    mgr.set_dispatch_hook(Arc::new(OrderingHook(o)));

    // External handler records "handler" second
    let o2 = order.clone();
    let handler: EventHandler = Arc::new(move |_| {
        o2.lock().unwrap().push("handler");
    });
    let _id = mgr.subscribe_all(handler).await;

    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));

    let seq = order.lock().unwrap().clone();
    assert_eq!(seq, vec!["hook", "handler"]);
}

#[tokio::test]
async fn dispatch_increments_tick() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    assert_eq!(mgr.tick(), 0);

    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));
    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::RunnerOnline));
    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodTerminated));

    assert_eq!(mgr.tick(), 3);
}

#[tokio::test]
async fn dispatch_hook_panic_does_not_abort_loop() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    mgr.set_dispatch_hook(Arc::new(PanickingHook));

    let counter = Arc::new(AtomicU32::new(0));
    let c = counter.clone();
    let handler: EventHandler = Arc::new(move |_| {
        c.fetch_add(1, Ordering::SeqCst);
    });
    let _id = mgr.subscribe_all(handler).await;

    // catch_unwind inside dispatch_event must swallow the panic — both
    // tick increment and external handler must still run.
    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));

    assert_eq!(mgr.tick(), 1);
    assert_eq!(counter.load(Ordering::SeqCst), 1);
}

#[tokio::test]
async fn dispatch_hook_can_be_replaced() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let counter_a = Arc::new(AtomicU32::new(0));
    let counter_b = Arc::new(AtomicU32::new(0));

    mgr.set_dispatch_hook(Arc::new(CountingHook { counter: counter_a.clone() }));
    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));
    assert_eq!(counter_a.load(Ordering::SeqCst), 1);

    mgr.set_dispatch_hook(Arc::new(CountingHook { counter: counter_b.clone() }));
    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));
    assert_eq!(counter_a.load(Ordering::SeqCst), 1, "old hook must not fire");
    assert_eq!(counter_b.load(Ordering::SeqCst), 1);
}

#[tokio::test]
async fn dispatch_with_no_hook_still_runs_external_handlers() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let counter = Arc::new(AtomicU32::new(0));
    let c = counter.clone();
    let handler: EventHandler = Arc::new(move |_| {
        c.fetch_add(1, Ordering::SeqCst);
    });
    let _id = mgr.subscribe_all(handler).await;

    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));

    assert_eq!(mgr.tick(), 1);
    assert_eq!(counter.load(Ordering::SeqCst), 1);
}

#[tokio::test]
async fn dispatch_hook_holding_own_lock_does_not_deadlock_external_read() {
    // Reentrancy guard: hook implementation holds a separate lock and
    // mutates state; external handler concurrently reads via a wrapper
    // closure. dispatch_event must hold no events-side lock while
    // calling either, so no lock-order inversion is possible.
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let shared = Arc::new(parking_lot::RwLock::new(0u32));

    struct WriteHook {
        shared: Arc<parking_lot::RwLock<u32>>,
    }
    impl EventDispatchHook for WriteHook {
        fn dispatch(&self, _event: &RealtimeEvent) {
            *self.shared.write() += 1;
        }
    }
    mgr.set_dispatch_hook(Arc::new(WriteHook { shared: shared.clone() }));

    let s = shared.clone();
    let handler: EventHandler = Arc::new(move |_| {
        // Acquiring read lock during handler MUST be safe — the hook
        // already released its write lock by now.
        let _v = *s.read();
    });
    let _id = mgr.subscribe_all(handler).await;

    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));
    assert_eq!(*shared.read(), 1);
}

// ---- Phase 5: TickListener push to FFI ----

struct RecordingTickListener {
    ticks: Arc<parking_lot::Mutex<Vec<u64>>>,
}
impl TickListener for RecordingTickListener {
    fn on_tick(&self, tick: u64) {
        self.ticks.lock().push(tick);
    }
}

struct PanickingTickListener;
impl TickListener for PanickingTickListener {
    fn on_tick(&self, _tick: u64) {
        panic!("listener boom");
    }
}

#[tokio::test]
async fn tick_listener_fires_after_each_dispatch() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let ticks = Arc::new(parking_lot::Mutex::new(Vec::new()));
    mgr.set_tick_listener(Arc::new(RecordingTickListener { ticks: ticks.clone() }));

    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));
    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::RunnerOnline));
    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodTerminated));

    let recorded = ticks.lock().clone();
    assert_eq!(recorded, vec![1, 2, 3]);
    assert_eq!(mgr.tick(), 3);
}

#[tokio::test]
async fn tick_listener_panic_does_not_abort_dispatch() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    mgr.set_tick_listener(Arc::new(PanickingTickListener));

    // Add external handler to confirm dispatch flow continues past panic.
    let counter = Arc::new(AtomicU32::new(0));
    let c = counter.clone();
    let handler: EventHandler = Arc::new(move |_| { c.fetch_add(1, Ordering::SeqCst); });
    let _id = mgr.subscribe_all(handler).await;

    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));

    // Tick still incremented + external handler still fired.
    assert_eq!(mgr.tick(), 1);
    assert_eq!(counter.load(Ordering::SeqCst), 1);
}

#[tokio::test]
async fn clear_tick_listener_stops_callbacks() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let ticks = Arc::new(parking_lot::Mutex::new(Vec::new()));
    mgr.set_tick_listener(Arc::new(RecordingTickListener { ticks: ticks.clone() }));

    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));
    assert_eq!(ticks.lock().len(), 1);

    mgr.clear_tick_listener();
    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));
    assert_eq!(ticks.lock().len(), 1, "post-clear dispatch must not fire listener");
}

#[tokio::test]
async fn tick_listener_replaces_prior_registration() {
    let mgr = EventSubscriptionManager::with_default_options(make_client());
    let a = Arc::new(parking_lot::Mutex::new(Vec::new()));
    let b = Arc::new(parking_lot::Mutex::new(Vec::new()));

    mgr.set_tick_listener(Arc::new(RecordingTickListener { ticks: a.clone() }));
    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));
    assert_eq!(a.lock().len(), 1);

    mgr.set_tick_listener(Arc::new(RecordingTickListener { ticks: b.clone() }));
    crate::subscription_manager::dispatch_event(&mgr.inner, &make_event(EventType::PodCreated));
    assert_eq!(a.lock().len(), 1, "old listener must not fire after replace");
    assert_eq!(b.lock().len(), 1);
}

// Suppress unused import lint on AtomicU64 (used by other tests indirectly
// once we add ordering-related assertions in follow-ups).
#[allow(dead_code)]
fn _atomic_u64_anchor() -> AtomicU64 { AtomicU64::new(0) }
