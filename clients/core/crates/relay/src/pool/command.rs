use agentsmesh_transport::runtime::Runtime;

use crate::command::Command;
use crate::error::RelayError;
use crate::types::RelayStatus;

use super::RelayConnectionPool;

impl<R: Runtime> RelayConnectionPool<R> {
    pub async fn send(&self, pod_key: &str, data: &str) {
        self.send_command(
            pod_key,
            Command::Send {
                data: data.to_string(),
            },
        );
    }

    pub async fn send_resize(&self, pod_key: &str, cols: u16, rows: u16) {
        self.send_command(
            pod_key,
            Command::Resize {
                cols,
                rows,
                force: false,
            },
        );
    }

    pub async fn force_resize(&self, pod_key: &str, cols: u16, rows: u16) {
        self.send_command(
            pod_key,
            Command::Resize {
                cols,
                rows,
                force: true,
            },
        );
    }

    pub async fn send_acp_command(
        &self,
        pod_key: &str,
        command: &serde_json::Value,
    ) -> Result<(), RelayError> {
        let ready = self
            .inner
            .read()
            .pods
            .get(pod_key)
            .map(|handle| handle.snapshot.read().status == RelayStatus::Connected)
            .unwrap_or(false);
        if !ready {
            return Err(RelayError::NotConnected(pod_key.into()));
        }
        if self.send_command(
            pod_key,
            Command::SendAcp {
                command: command.clone(),
            },
        ) {
            Ok(())
        } else {
            Err(RelayError::NotConnected(pod_key.into()))
        }
    }
}
