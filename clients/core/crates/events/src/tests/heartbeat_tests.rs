use std::time::Duration;

use futures::channel::mpsc;
use futures::stream::StreamExt;

use crate::heartbeat::{HeartbeatEvent, HeartbeatManager};

#[tokio::test]
async fn test_ping_sent_after_interval() {
    let mut hb = HeartbeatManager::new(80, 5000);
    let (tx, mut rx) = mpsc::unbounded();
    hb.start(tx);

    match tokio::time::timeout(Duration::from_millis(300), rx.next()).await {
        Ok(Some(HeartbeatEvent::SendPing)) => {}
        other => panic!("expected SendPing, got timeout={}", other.is_err()),
    }
    hb.stop();
}

#[tokio::test]
async fn test_pong_timeout_fires_when_no_pong() {
    let mut hb = HeartbeatManager::new(200, 50);
    let (tx, mut rx) = mpsc::unbounded();
    hb.start(tx);

    match tokio::time::timeout(Duration::from_millis(500), rx.next()).await {
        Ok(Some(HeartbeatEvent::SendPing)) => {}
        _ => panic!("expected SendPing first"),
    }
    match tokio::time::timeout(Duration::from_millis(300), rx.next()).await {
        Ok(Some(HeartbeatEvent::PongTimeout)) => {}
        other => panic!("expected PongTimeout, got timeout={}", other.is_err()),
    }
}

#[tokio::test]
async fn test_pong_received_prevents_timeout() {
    let mut hb = HeartbeatManager::new(50, 100);
    let (tx, mut rx) = mpsc::unbounded();
    hb.start(tx);

    // consume first SendPing
    let _ = rx.next().await;
    hb.pong_received();

    // next event should be another SendPing, not PongTimeout
    match tokio::time::timeout(Duration::from_millis(200), rx.next()).await {
        Ok(Some(HeartbeatEvent::SendPing)) => {}
        Ok(Some(HeartbeatEvent::PongTimeout)) => panic!("got PongTimeout after pong_received"),
        other => panic!("unexpected: timeout={}", other.is_err()),
    }
    hb.stop();
}

#[tokio::test]
async fn test_multiple_ping_cycles() {
    let mut hb = HeartbeatManager::new(40, 5000);
    let (tx, mut rx) = mpsc::unbounded();
    hb.start(tx);

    for _ in 0..3 {
        match tokio::time::timeout(Duration::from_millis(200), rx.next()).await {
            Ok(Some(HeartbeatEvent::SendPing)) => hb.pong_received(),
            other => panic!("expected SendPing, got timeout={}", other.is_err()),
        }
    }
    hb.stop();
}

#[tokio::test]
async fn test_stop_terminates_loop() {
    let mut hb = HeartbeatManager::new(30, 30);
    let (tx, mut rx) = mpsc::unbounded();
    hb.start(tx);
    hb.stop();

    tokio::time::sleep(Duration::from_millis(100)).await;
    // channel should be closed or empty
    assert!(matches!(rx.try_next(), Err(_) | Ok(None)));
}

#[tokio::test]
async fn test_restart_after_stop() {
    let mut hb = HeartbeatManager::new(50, 5000);
    let (tx1, _rx1) = mpsc::unbounded();
    hb.start(tx1);
    hb.stop();

    let (tx2, mut rx2) = mpsc::unbounded();
    hb.start(tx2);

    match tokio::time::timeout(Duration::from_millis(200), rx2.next()).await {
        Ok(Some(HeartbeatEvent::SendPing)) => {}
        other => panic!("expected SendPing after restart, got timeout={}", other.is_err()),
    }
    hb.stop();
}

#[tokio::test]
async fn test_drop_stops_heartbeat() {
    let (tx, mut rx) = mpsc::unbounded();
    {
        let mut hb = HeartbeatManager::new(30, 5000);
        hb.start(tx);
    } // hb dropped here

    tokio::time::sleep(Duration::from_millis(100)).await;
    // may get one ping but should not infinite-loop
    let count = std::iter::from_fn(|| rx.try_next().ok().flatten()).count();
    assert!(count <= 1);
}
