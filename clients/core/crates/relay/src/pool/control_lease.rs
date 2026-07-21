use agentcloud_transport::runtime::Runtime;

use super::RelayConnectionPool;
use crate::command::Command;
use crate::error::RelayError;
use crate::types::{ControlLeaseInfo, RelayStatus};

impl<R: Runtime> RelayConnectionPool<R> {
    pub async fn acquire_control(
        &self,
        pod_key: &str,
        client_label: &str,
    ) -> Result<(), RelayError> {
        self.send_ready_command(
            pod_key,
            Command::AcquireControl {
                client_label: client_label.to_string(),
            },
        )
    }

    pub async fn renew_control(&self, pod_key: &str, lease_id: &str) -> Result<(), RelayError> {
        self.send_ready_command(
            pod_key,
            Command::RenewControl {
                lease_id: lease_id.to_string(),
            },
        )
    }

    pub async fn release_control(&self, pod_key: &str, lease_id: &str) -> Result<(), RelayError> {
        self.send_ready_command(
            pod_key,
            Command::ReleaseControl {
                lease_id: lease_id.to_string(),
            },
        )
    }

    pub async fn get_control_lease(&self, pod_key: &str) -> ControlLeaseInfo {
        self.inner
            .read()
            .pods
            .get(pod_key)
            .map(|handle| handle.snapshot.read().control_lease.clone())
            .unwrap_or_default()
    }

    fn send_ready_command(&self, pod_key: &str, command: Command) -> Result<(), RelayError> {
        let ready = self
            .inner
            .read()
            .pods
            .get(pod_key)
            .map(|handle| handle.snapshot.read().status == RelayStatus::Connected)
            .unwrap_or(false);
        if !ready || !self.send_command(pod_key, command) {
            return Err(RelayError::NotConnected(pod_key.into()));
        }
        Ok(())
    }
}
