use agentsmesh_transport::runtime::Runtime;

use crate::types::RelayStatus;

use super::RelayConnectionPool;

impl<R: Runtime> RelayConnectionPool<R> {
    pub async fn get_status(&self, pod_key: &str) -> RelayStatus {
        self.inner
            .read()
            .pods
            .get(pod_key)
            .map(|handle| handle.snapshot.read().status)
            .unwrap_or(RelayStatus::Disconnected)
    }

    pub async fn is_runner_disconnected(&self, pod_key: &str) -> bool {
        self.inner
            .read()
            .pods
            .get(pod_key)
            .map(|handle| handle.snapshot.read().runner_disconnected)
            .unwrap_or(false)
    }

    pub async fn get_pod_size(&self, pod_key: &str) -> Option<(u16, u16)> {
        self.inner
            .read()
            .pods
            .get(pod_key)
            .and_then(|handle| handle.snapshot.read().pod_size)
    }
}
