use agentcloud_transport::{WebSocketConnection, WsReceiver, WsSender};

use crate::error::RelayError;

/// Dial the relay and hand back the split sender/receiver. Transport already
/// runs its own read/write pump (inbound frames surface via `WsReceiver::recv`),
/// so the driver selects directly on the receiver — relay keeps no extra read loop.
pub(crate) async fn connect(
    relay_url: &str,
    token: &str,
) -> Result<(WsSender, WsReceiver), RelayError> {
    let url = format!(
        "{relay_url}/browser/relay?token={}",
        urlencoding::encode(token)
    );
    let conn = WebSocketConnection::connect(&url)
        .await
        .map_err(|e| RelayError::Connection(e.to_string()))?;
    Ok(conn.into_split())
}
