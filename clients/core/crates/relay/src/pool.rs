use agentsmesh_transport::runtime::{PlatformRuntime, Runtime};
use futures::channel::mpsc;
use parking_lot::RwLock;
use std::collections::HashMap;
use std::sync::Arc;

use crate::command::Command;
use crate::driver::Driver;
use crate::types::{
    ConnectionHandle, ControlLeaseInfo, DisconnectCallback, OutputCallback, RelayStatus,
    StatusSnapshot,
};

mod command;
mod control_lease;
mod default;
mod listener;
mod status;

/// Thin routing table over per-pod driver actors. Holds only each pod's command
/// sender + status mirror, plus pool-scoped listeners (which may be registered
/// before a driver exists). All connection state lives inside the drivers.
#[derive(Clone)]
pub struct RelayConnectionPool<R: Runtime = PlatformRuntime> {
    inner: Arc<RwLock<PoolRouter>>,
    runtime: R,
    unsubscribe_tx: mpsc::UnboundedSender<(String, String)>,
}

pub(crate) struct PoolRouter {
    pub pods: HashMap<String, PodHandle>,
    pub status_listeners: HashMap<String, HashMap<String, crate::types::StatusCallback>>,
    pub acp_listeners: HashMap<String, HashMap<String, crate::types::AcpCallback>>,
    pub on_pod_disconnected: Option<DisconnectCallback>,
    next_listener_id: u64,
}

pub(crate) struct PodHandle {
    cmd_tx: mpsc::UnboundedSender<Command>,
    snapshot: Arc<RwLock<StatusSnapshot>>,
}

impl RelayConnectionPool<PlatformRuntime> {
    pub fn new() -> (Self, mpsc::UnboundedReceiver<(String, String)>) {
        Self::with_runtime(PlatformRuntime)
    }
}

impl<R: Runtime> RelayConnectionPool<R> {
    pub fn with_runtime(runtime: R) -> (Self, mpsc::UnboundedReceiver<(String, String)>) {
        let (tx, rx) = mpsc::unbounded();
        let inner = PoolRouter {
            pods: HashMap::new(),
            status_listeners: HashMap::new(),
            acp_listeners: HashMap::new(),
            on_pod_disconnected: None,
            next_listener_id: 0,
        };
        (
            Self {
                inner: Arc::new(RwLock::new(inner)),
                runtime,
                unsubscribe_tx: tx,
            },
            rx,
        )
    }

    pub fn set_on_pod_disconnected(&self, callback: DisconnectCallback) {
        self.inner.write().on_pod_disconnected = Some(callback);
    }

    pub async fn subscribe(
        &self,
        pod_key: &str,
        subscription_id: &str,
        relay_url: &str,
        relay_token: &str,
        callback: OutputCallback,
    ) -> ConnectionHandle {
        tracing::info!(target: "relay", pod_key, %subscription_id, "subscribe");
        let cmd_tx = {
            let mut router = self.inner.write();
            if let Some(handle) = router.pods.get(pod_key) {
                let tx = handle.cmd_tx.clone();
                let _ = tx.unbounded_send(Command::AddSubscriber {
                    sub_id: subscription_id.to_string(),
                    cb: callback,
                });
                tx
            } else {
                let (cmd_tx, cmd_rx) = mpsc::unbounded();
                // Mirror starts at Connecting (matching the driver's initial
                // state), so get_status during the first connect window doesn't
                // read the StatusSnapshot::default() Disconnected.
                let snapshot = Arc::new(RwLock::new(StatusSnapshot {
                    status: RelayStatus::Connecting,
                    runner_disconnected: false,
                    pod_size: None,
                    control_lease: ControlLeaseInfo::default(),
                }));
                router.pods.insert(
                    pod_key.to_string(),
                    PodHandle {
                        cmd_tx: cmd_tx.clone(),
                        snapshot: Arc::clone(&snapshot),
                    },
                );
                tracing::debug!(target: "relay", pod_key, "no live driver — spawning");
                Driver::spawn(
                    self.runtime.clone(),
                    Arc::clone(&self.inner),
                    pod_key.to_string(),
                    relay_url.to_string(),
                    relay_token.to_string(),
                    snapshot,
                    cmd_rx,
                    (subscription_id.to_string(), callback),
                );
                cmd_tx
            }
        };
        ConnectionHandle::new(
            pod_key.to_string(),
            subscription_id.to_string(),
            cmd_tx,
            self.unsubscribe_tx.clone(),
        )
    }

    pub async fn unsubscribe(&self, pod_key: &str, subscription_id: &str) {
        self.send_command(
            pod_key,
            Command::RemoveSubscriber {
                sub_id: subscription_id.to_string(),
            },
        );
    }

    pub async fn disconnect(&self, pod_key: &str) {
        tracing::info!(target: "relay", pod_key, "disconnect");
        self.send_command(pod_key, Command::Disconnect);
    }

    pub async fn disconnect_all(&self) {
        let txs: Vec<_> = self
            .inner
            .read()
            .pods
            .values()
            .map(|h| h.cmd_tx.clone())
            .collect();
        for tx in txs {
            let _ = tx.unbounded_send(Command::Disconnect);
        }
    }

    /// Forward a command to a live driver; false if the pod has no driver.
    fn send_command(&self, pod_key: &str, cmd: Command) -> bool {
        match self.inner.read().pods.get(pod_key) {
            Some(h) => h.cmd_tx.unbounded_send(cmd).is_ok(),
            None => false,
        }
    }
}
