use std::sync::atomic::{AtomicU32, Ordering};
use std::sync::Arc;

use crate::connection_loop::ManagerOpts;
use crate::subscription_manager::{dispatch_event, set_state, Inner};
use crate::types::{
    ConnectionState, EventHandler, EventSubscriptionManagerOptions, PingMessage, RealtimeEvent,
};
use crate::EventType;
use parking_lot::RwLock;
use std::collections::HashMap;

fn make_inner() -> Arc<RwLock<Inner>> {
    Arc::new(RwLock::new(Inner {
        handlers: HashMap::new(),
        global_handlers: HashMap::new(),
        state_listeners: HashMap::new(),
        connection_state: ConnectionState::Disconnected,
    }))
}

// ── ManagerOpts::from_options ──

#[test]
fn manager_opts_from_options() {
    let opts = EventSubscriptionManagerOptions {
        max_reconnect_attempts: 5,
        initial_reconnect_delay_ms: 200,
        max_reconnect_delay_ms: 8000,
        ping_interval_ms: 15000,
        pong_timeout_ms: 3000,
    };
    let mo = ManagerOpts::from_options(&opts);
    assert_eq!(mo.max_reconnect_attempts, 5);
    assert_eq!(mo.initial_reconnect_delay_ms, 200);
    assert_eq!(mo.max_reconnect_delay_ms, 8000);
    assert_eq!(mo.ping_interval_ms, 15000);
    assert_eq!(mo.pong_timeout_ms, 3000);
}

#[test]
fn manager_opts_from_default_options() {
    let opts = EventSubscriptionManagerOptions::default();
    let mo = ManagerOpts::from_options(&opts);
    assert_eq!(mo.max_reconnect_attempts, 10);
    assert_eq!(mo.initial_reconnect_delay_ms, 1000);
    assert_eq!(mo.max_reconnect_delay_ms, 30000);
    assert_eq!(mo.ping_interval_ms, 30000);
    assert_eq!(mo.pong_timeout_ms, 10000);
}

// ── PingMessage serialization ──

#[test]
fn ping_message_serializes_correctly() {
    let ping = PingMessage::new(1234567890);
    let json = serde_json::to_string(&ping).unwrap();
    assert!(json.contains("\"type\":\"ping\""));
    assert!(json.contains("\"timestamp\":1234567890"));
}

#[test]
fn ping_message_zero_timestamp() {
    let ping = PingMessage::new(0);
    let json = serde_json::to_string(&ping).unwrap();
    assert!(json.contains("\"timestamp\":0"));
}

// ── set_state ──

#[tokio::test]
async fn set_state_transitions_and_notifies() {
    let inner = make_inner();
    let received = Arc::new(std::sync::Mutex::new(Vec::new()));
    let r = received.clone();
    let listener = Arc::new(move |s: ConnectionState| {
        r.lock().unwrap().push(s);
    });

    {
        let mut guard = inner.write();
        guard
            .state_listeners
            .insert(crate::types::SubscriptionId(1), listener);
    }

    set_state(&inner, ConnectionState::Connecting);
    let states = received.lock().unwrap();
    assert_eq!(states.len(), 1);
    assert_eq!(states[0], ConnectionState::Connecting);
}

#[tokio::test]
async fn set_state_same_state_is_noop() {
    let inner = make_inner();
    let counter = Arc::new(AtomicU32::new(0));
    let c = counter.clone();
    let listener = Arc::new(move |_: ConnectionState| {
        c.fetch_add(1, Ordering::SeqCst);
    });

    {
        let mut guard = inner.write();
        guard
            .state_listeners
            .insert(crate::types::SubscriptionId(1), listener);
    }

    // inner starts as Disconnected; setting to Disconnected again should not notify
    set_state(&inner, ConnectionState::Disconnected);
    assert_eq!(counter.load(Ordering::SeqCst), 0);
}

// ── dispatch_event ──

#[tokio::test]
async fn dispatch_event_calls_typed_handler() {
    let inner = make_inner();
    let counter = Arc::new(AtomicU32::new(0));
    let c = counter.clone();
    let handler: EventHandler = Arc::new(move |_| {
        c.fetch_add(1, Ordering::SeqCst);
    });

    {
        let mut guard = inner.write();
        guard
            .handlers
            .entry(EventType::PodCreated)
            .or_default()
            .insert(crate::types::SubscriptionId(1), handler);
    }

    let event: RealtimeEvent = serde_json::from_value(serde_json::json!({
        "type": "pod:created",
        "category": "entity",
        "organization_id": 1,
        "data": {},
        "timestamp": 1000
    }))
    .unwrap();

    dispatch_event(&inner, &event);
    assert_eq!(counter.load(Ordering::SeqCst), 1);
}

#[tokio::test]
async fn dispatch_event_calls_global_handler() {
    let inner = make_inner();
    let counter = Arc::new(AtomicU32::new(0));
    let c = counter.clone();
    let handler: EventHandler = Arc::new(move |_| {
        c.fetch_add(1, Ordering::SeqCst);
    });

    {
        let mut guard = inner.write();
        guard
            .global_handlers
            .insert(crate::types::SubscriptionId(1), handler);
    }

    let event: RealtimeEvent = serde_json::from_value(serde_json::json!({
        "type": "runner:online",
        "category": "entity",
        "organization_id": 1,
        "data": {},
        "timestamp": 2000
    }))
    .unwrap();

    dispatch_event(&inner, &event);
    assert_eq!(counter.load(Ordering::SeqCst), 1);
}

#[tokio::test]
async fn dispatch_event_calls_both_typed_and_global() {
    let inner = make_inner();
    let typed_count = Arc::new(AtomicU32::new(0));
    let global_count = Arc::new(AtomicU32::new(0));

    let tc = typed_count.clone();
    let typed_handler: EventHandler = Arc::new(move |_| {
        tc.fetch_add(1, Ordering::SeqCst);
    });

    let gc = global_count.clone();
    let global_handler: EventHandler = Arc::new(move |_| {
        gc.fetch_add(1, Ordering::SeqCst);
    });

    {
        let mut guard = inner.write();
        guard
            .handlers
            .entry(EventType::TicketCreated)
            .or_default()
            .insert(crate::types::SubscriptionId(1), typed_handler);
        guard
            .global_handlers
            .insert(crate::types::SubscriptionId(2), global_handler);
    }

    let event: RealtimeEvent = serde_json::from_value(serde_json::json!({
        "type": "ticket:created",
        "category": "entity",
        "organization_id": 1,
        "data": {},
        "timestamp": 3000
    }))
    .unwrap();

    dispatch_event(&inner, &event);
    assert_eq!(typed_count.load(Ordering::SeqCst), 1);
    assert_eq!(global_count.load(Ordering::SeqCst), 1);
}

#[tokio::test]
async fn dispatch_event_no_matching_handler_is_noop() {
    let inner = make_inner();
    let event: RealtimeEvent = serde_json::from_value(serde_json::json!({
        "type": "pod:terminated",
        "category": "entity",
        "organization_id": 1,
        "data": {},
        "timestamp": 4000
    }))
    .unwrap();

    // no handlers registered — should not panic
    dispatch_event(&inner, &event);
}
