use agentcloud_protocol::{decode_message, MsgType};

use crate::control_lease::{encode_acquire, encode_release, encode_renew};
use crate::dispatch::{dispatch_message, DispatchAction};
use crate::pool::RelayConnectionPool;
use crate::types::{ControlLeaseInfo, ControlLeaseState, StatusSnapshot};

#[test]
fn dispatches_granted_control_lease() {
    let payload = br#"{
        "type":"control_lease",
        "status":"granted",
        "lease_id":"lease-1",
        "expires_at":1234
    }"#;
    let action = dispatch_message(MsgType::Control, payload, &[]);
    assert_eq!(
        action,
        DispatchAction::ControlLease(ControlLeaseInfo {
            state: ControlLeaseState::Granted,
            lease_id: Some("lease-1".into()),
            expires_at: Some(1234),
        })
    );
}

#[test]
fn dispatches_non_owner_control_states_without_credentials() {
    for (wire, state) in [
        ("busy", ControlLeaseState::Busy),
        ("released", ControlLeaseState::Released),
        ("expired", ControlLeaseState::Expired),
        ("control_required", ControlLeaseState::Required),
    ] {
        let payload = serde_json::json!({
            "type": "control_lease",
            "status": wire,
        });
        let action = dispatch_message(MsgType::Control, payload.to_string().as_bytes(), &[]);
        assert_eq!(
            action,
            DispatchAction::ControlLease(ControlLeaseInfo {
                state,
                lease_id: None,
                expires_at: None,
            })
        );
    }
}

#[test]
fn encodes_explicit_control_lease_commands() {
    let cases = [
        (
            encode_acquire("mobile").unwrap(),
            serde_json::json!({
                "type": "control_lease",
                "action": "acquire",
                "client_label": "mobile",
            }),
        ),
        (
            encode_renew("lease-1").unwrap(),
            serde_json::json!({
                "type": "control_lease",
                "action": "renew",
                "lease_id": "lease-1",
            }),
        ),
        (
            encode_release("lease-1").unwrap(),
            serde_json::json!({
                "type": "control_lease",
                "action": "release",
                "lease_id": "lease-1",
            }),
        ),
    ];
    for (frame, expected) in cases {
        let (msg_type, payload) = decode_message(&frame).unwrap();
        assert_eq!(msg_type, MsgType::Control);
        assert_eq!(
            serde_json::from_slice::<serde_json::Value>(payload).unwrap(),
            expected
        );
    }
}

#[test]
fn control_lease_defaults_to_observer() {
    assert_eq!(
        StatusSnapshot::default().control_lease,
        ControlLeaseInfo::default()
    );
    assert_eq!(
        ControlLeaseInfo::default().state,
        ControlLeaseState::Observer
    );
}

#[test]
fn control_commands_fail_when_link_is_not_ready() {
    let (pool, _rx) = RelayConnectionPool::new();
    futures::executor::block_on(async {
        assert!(pool.acquire_control("missing", "mobile").await.is_err());
        assert!(pool.renew_control("missing", "lease").await.is_err());
        assert!(pool.release_control("missing", "lease").await.is_err());
    });
}
