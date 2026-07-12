use agentsmesh_transport::runtime::PlatformRuntime;

use super::RelayConnectionPool;

impl Default for RelayConnectionPool<PlatformRuntime> {
    fn default() -> Self {
        Self::new().0
    }
}
