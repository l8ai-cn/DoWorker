mod command;
mod connection;
mod control_lease;
pub mod dispatch;
mod driver;
pub mod error;
pub mod pool;
pub mod retry;
pub mod types;

pub use error::RelayError;
pub use pool::RelayConnectionPool;
pub use types::{
    AcpCallback, ConnectionHandle, ControlLeaseInfo, ControlLeaseState, DisconnectCallback,
    OutputCallback, RelayStatus, RelayStatusInfo, StatusCallback,
};

#[cfg(test)]
mod control_lease_integration_tests;
#[cfg(test)]
mod control_lease_tests;
#[cfg(test)]
mod dispatch_tests;
#[cfg(test)]
mod integration_tests;
#[cfg(test)]
mod test_support;
