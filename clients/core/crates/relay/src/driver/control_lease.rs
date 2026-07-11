use agentsmesh_transport::runtime::Runtime;

use super::Driver;
use crate::types::ControlLeaseInfo;

impl<R: Runtime> Driver<R> {
    pub(super) fn set_control_lease(&mut self, lease: ControlLeaseInfo) {
        if self.control_lease == lease {
            return;
        }
        self.control_lease = lease;
        self.write_snapshot();
        self.notify_status();
    }
}
