use std::sync::Arc;

use agentcloud_protocol::MsgType;
use futures::channel::mpsc;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RelayStatus {
    Connecting,
    Connected,
    Disconnected,
    Error,
}

impl std::fmt::Display for RelayStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Connecting => write!(f, "connecting"),
            Self::Connected => write!(f, "connected"),
            Self::Disconnected => write!(f, "disconnected"),
            Self::Error => write!(f, "error"),
        }
    }
}

pub type OutputCallback = Arc<dyn Fn(Vec<u8>) + Send + Sync>;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct RelayStatusInfo {
    pub status: RelayStatus,
    pub runner_disconnected: bool,
    pub control_lease: ControlLeaseInfo,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ControlLeaseState {
    Observer,
    Granted,
    Busy,
    Released,
    Expired,
    Required,
}

impl ControlLeaseState {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Observer => "observer",
            Self::Granted => "granted",
            Self::Busy => "busy",
            Self::Released => "released",
            Self::Expired => "expired",
            Self::Required => "control_required",
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ControlLeaseInfo {
    pub state: ControlLeaseState,
    pub lease_id: Option<String>,
    pub expires_at: Option<i64>,
}

impl ControlLeaseInfo {
    pub(crate) fn from_wire(
        status: &str,
        lease_id: Option<&str>,
        expires_at: Option<i64>,
    ) -> Option<Self> {
        let state = match status {
            "granted" => ControlLeaseState::Granted,
            "busy" => ControlLeaseState::Busy,
            "released" => ControlLeaseState::Released,
            "expired" => ControlLeaseState::Expired,
            "control_required" => ControlLeaseState::Required,
            _ => return None,
        };
        Some(Self {
            state,
            lease_id: lease_id.map(str::to_owned),
            expires_at: expires_at.filter(|value| *value > 0),
        })
    }
}

impl Default for ControlLeaseInfo {
    fn default() -> Self {
        Self {
            state: ControlLeaseState::Observer,
            lease_id: None,
            expires_at: None,
        }
    }
}

/// Driver-owned, pool-readable status mirror. The driver task is the single
/// writer (under its own lock); the pool's `get_status` / `is_runner_disconnected`
/// / `get_pod_size` read it directly instead of round-tripping a command.
#[derive(Debug, Clone)]
pub(crate) struct StatusSnapshot {
    pub status: RelayStatus,
    pub runner_disconnected: bool,
    pub pod_size: Option<(u16, u16)>,
    pub control_lease: ControlLeaseInfo,
}

impl Default for StatusSnapshot {
    fn default() -> Self {
        Self {
            status: RelayStatus::Disconnected,
            runner_disconnected: false,
            pod_size: None,
            control_lease: ControlLeaseInfo::default(),
        }
    }
}

pub type StatusCallback = Arc<dyn Fn(RelayStatusInfo) + Send + Sync>;
pub type AcpCallback = Arc<dyn Fn(MsgType, serde_json::Value) + Send + Sync>;
// Fired once when a pod connection is fully torn down (disconnect_inner) so
// adapters can drop their register-once guard and re-register listeners on the
// next subscribe. Carries the pod_key.
pub type DisconnectCallback = Arc<dyn Fn(String) + Send + Sync>;

pub struct ConnectionHandle {
    pub pod_key: String,
    pub subscription_id: String,
    cmd_tx: mpsc::UnboundedSender<crate::command::Command>,
    unsubscribe_tx: mpsc::UnboundedSender<(String, String)>,
}

impl ConnectionHandle {
    pub(crate) fn new(
        pod_key: String,
        subscription_id: String,
        cmd_tx: mpsc::UnboundedSender<crate::command::Command>,
        unsubscribe_tx: mpsc::UnboundedSender<(String, String)>,
    ) -> Self {
        Self {
            pod_key,
            subscription_id,
            cmd_tx,
            unsubscribe_tx,
        }
    }

    pub fn send(&self, data: Vec<u8>) {
        let _ = self.cmd_tx.unbounded_send(crate::command::Command::Send {
            data: String::from_utf8_lossy(&data).into_owned(),
        });
    }

    pub fn unsubscribe(&self) {
        let _ = self
            .unsubscribe_tx
            .unbounded_send((self.pod_key.clone(), self.subscription_id.clone()));
    }
}
