use std::sync::Arc;

use agentsmesh_transport::runtime::Runtime;

use crate::types::{AcpCallback, ControlLeaseInfo, RelayStatus, RelayStatusInfo, StatusCallback};

use super::RelayConnectionPool;

impl<R: Runtime> RelayConnectionPool<R> {
    pub async fn on_status_change(&self, pod_key: &str, listener: StatusCallback) {
        let listener_id = self.next_listener_id("legacy-status");
        self.set_status_listener(pod_key, &listener_id, listener)
            .await;
    }

    pub async fn set_status_listener(
        &self,
        pod_key: &str,
        listener_id: &str,
        listener: StatusCallback,
    ) {
        {
            let mut router = self.inner.write();
            router
                .status_listeners
                .entry(pod_key.to_string())
                .or_default()
                .insert(listener_id.to_string(), Arc::clone(&listener));
        }
        listener(self.status_info(pod_key));
    }

    pub fn remove_status_listener(&self, pod_key: &str, listener_id: &str) {
        let mut router = self.inner.write();
        if let Some(listeners) = router.status_listeners.get_mut(pod_key) {
            listeners.remove(listener_id);
            if listeners.is_empty() {
                router.status_listeners.remove(pod_key);
            }
        }
    }

    pub async fn on_acp_message(&self, pod_key: &str, listener: AcpCallback) {
        let listener_id = self.next_listener_id("legacy-acp");
        self.set_acp_listener(pod_key, &listener_id, listener).await;
    }

    pub async fn set_acp_listener(&self, pod_key: &str, listener_id: &str, listener: AcpCallback) {
        self.inner
            .write()
            .acp_listeners
            .entry(pod_key.to_string())
            .or_default()
            .insert(listener_id.to_string(), listener);
    }

    pub fn remove_acp_listener(&self, pod_key: &str, listener_id: &str) {
        let mut router = self.inner.write();
        if let Some(listeners) = router.acp_listeners.get_mut(pod_key) {
            listeners.remove(listener_id);
            if listeners.is_empty() {
                router.acp_listeners.remove(pod_key);
            }
        }
    }

    fn next_listener_id(&self, prefix: &str) -> String {
        let mut router = self.inner.write();
        router.next_listener_id += 1;
        format!("{prefix}-{}", router.next_listener_id)
    }

    fn status_info(&self, pod_key: &str) -> RelayStatusInfo {
        self.inner
            .read()
            .pods
            .get(pod_key)
            .map(|handle| {
                let snapshot = handle.snapshot.read();
                RelayStatusInfo {
                    status: snapshot.status,
                    runner_disconnected: snapshot.runner_disconnected,
                    control_lease: snapshot.control_lease.clone(),
                }
            })
            .unwrap_or(RelayStatusInfo {
                status: RelayStatus::Disconnected,
                runner_disconnected: false,
                control_lease: ControlLeaseInfo::default(),
            })
    }
}
