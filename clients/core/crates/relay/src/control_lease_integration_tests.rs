use std::sync::atomic::Ordering::SeqCst;
use std::time::Duration;

use agentsmesh_protocol::{decode_message, MsgType};

use crate::pool::RelayConnectionPool;
use crate::test_support::{
    make_output_cb, start_mock_relay, wait_ready, wait_transport, wait_until,
};
use crate::types::ControlLeaseState;

#[tokio::test]
async fn control_lease_commands_update_mirror_and_clear_on_reconnect() {
    let mock_relay = start_mock_relay().await;
    let (pool, _unsubscribe_rx) = RelayConnectionPool::new();
    let (output, _received) = make_output_cb();
    pool.subscribe("pod-1", "sub-1", &mock_relay.url, "token", output)
        .await;
    assert!(wait_transport(&mock_relay).await);
    mock_relay.push(
        MsgType::Snapshot,
        serde_json::json!({"serialized_content":"","cols":80,"rows":24})
            .to_string()
            .as_bytes(),
    );
    assert!(wait_ready(&pool, "pod-1").await);

    pool.acquire_control("pod-1", "mobile").await.unwrap();
    let acquire = receive_control(&mock_relay).await;
    assert_eq!(acquire["action"], "acquire");
    assert_eq!(acquire["client_label"], "mobile");

    mock_relay.push(
        MsgType::Control,
        br#"{"type":"control_lease","status":"granted","lease_id":"lease-1","expires_at":9999999999999}"#,
    );
    assert!(
        wait_for_control_state(&pool, "pod-1", ControlLeaseState::Granted).await,
        "granted state did not reach the pool mirror"
    );
    let granted = pool.get_control_lease("pod-1").await;
    assert_eq!(granted.lease_id.as_deref(), Some("lease-1"));

    pool.renew_control("pod-1", "lease-1").await.unwrap();
    assert_eq!(receive_control(&mock_relay).await["action"], "renew");
    pool.release_control("pod-1", "lease-1").await.unwrap();
    assert_eq!(receive_control(&mock_relay).await["action"], "release");
    mock_relay.push(
        MsgType::Control,
        br#"{"type":"control_lease","status":"released"}"#,
    );
    assert!(wait_for_control_state(&pool, "pod-1", ControlLeaseState::Released).await);

    pool.acquire_control("pod-1", "mobile").await.unwrap();
    let _ = receive_control(&mock_relay).await;
    mock_relay.push(
        MsgType::Control,
        br#"{"type":"control_lease","status":"granted","lease_id":"lease-2","expires_at":9999999999999}"#,
    );
    assert!(wait_for_control_state(&pool, "pod-1", ControlLeaseState::Granted).await);

    mock_relay.drop_signal.send(()).unwrap();
    assert!(
        wait_for_control_state(&pool, "pod-1", ControlLeaseState::Observer).await,
        "transport loss retained stale control ownership"
    );
    assert!(
        wait_until(
            || mock_relay.conn_count.load(SeqCst) >= 2,
            Duration::from_secs(20),
        )
        .await,
        "relay did not reconnect"
    );
}

async fn receive_control(mock_relay: &crate::test_support::MockRelay) -> serde_json::Value {
    let deadline = std::time::Instant::now() + Duration::from_secs(3);
    loop {
        let remaining = deadline.saturating_duration_since(std::time::Instant::now());
        let frame = mock_relay
            .recv(remaining)
            .await
            .expect("no control command");
        let Ok((MsgType::Control, payload)) = decode_message(&frame) else {
            continue;
        };
        return serde_json::from_slice(payload).expect("invalid control JSON");
    }
}

async fn wait_for_control_state(
    pool: &RelayConnectionPool,
    pod_key: &str,
    expected: ControlLeaseState,
) -> bool {
    let deadline = std::time::Instant::now() + Duration::from_secs(3);
    while std::time::Instant::now() < deadline {
        if pool.get_control_lease(pod_key).await.state == expected {
            return true;
        }
        tokio::time::sleep(Duration::from_millis(10)).await;
    }
    false
}
