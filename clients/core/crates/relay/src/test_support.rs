use std::sync::atomic::{AtomicUsize, Ordering::SeqCst};
use std::sync::{Arc, Mutex};
use std::time::Duration;

use agentcloud_protocol::{encode_message, MsgType};
use futures_util::stream::SplitSink;
use futures_util::{SinkExt, StreamExt};
use tokio::net::TcpStream;
use tokio::sync::mpsc::{unbounded_channel, UnboundedReceiver, UnboundedSender};
use tokio::sync::Mutex as TokioMutex;
use tokio_tungstenite::tungstenite::Message;
use tokio_tungstenite::WebSocketStream;

use crate::pool::RelayConnectionPool;
use crate::types::{OutputCallback, RelayStatus};

type ServerSink = SplitSink<WebSocketStream<TcpStream>, Message>;

pub(crate) struct MockRelay {
    pub(crate) url: String,
    to_client: UnboundedSender<Vec<u8>>,
    from_client: TokioMutex<UnboundedReceiver<Vec<u8>>>,
    pub(crate) drop_signal: UnboundedSender<()>,
    pub(crate) conn_count: Arc<AtomicUsize>,
}

pub(crate) async fn start_mock_relay() -> MockRelay {
    let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let url = format!("ws://{}", listener.local_addr().unwrap());
    let (to_client, mut to_client_rx) = unbounded_channel::<Vec<u8>>();
    let (from_client_tx, from_client_rx) = unbounded_channel::<Vec<u8>>();
    let (drop_signal, mut drop_rx) = unbounded_channel::<()>();
    let conn_count = Arc::new(AtomicUsize::new(0));
    let active: Arc<TokioMutex<Option<ServerSink>>> = Arc::new(TokioMutex::new(None));

    {
        let active = active.clone();
        tokio::spawn(async move {
            while let Some(frame) = to_client_rx.recv().await {
                if let Some(sink) = active.lock().await.as_mut() {
                    let _ = sink.send(Message::Binary(frame.into())).await;
                }
            }
        });
    }
    {
        let active = active.clone();
        tokio::spawn(async move {
            while drop_rx.recv().await.is_some() {
                if let Some(sink) = active.lock().await.as_mut() {
                    let _ = sink.send(Message::Close(None)).await;
                }
            }
        });
    }
    {
        let active = active.clone();
        let connection_count = conn_count.clone();
        tokio::spawn(async move {
            loop {
                let Ok((stream, _)) = listener.accept().await else {
                    break;
                };
                let ws = match tokio_tungstenite::accept_async(stream).await {
                    Ok(ws) => ws,
                    Err(_) => continue,
                };
                let (write, mut read) = ws.split();
                *active.lock().await = Some(write);
                connection_count.fetch_add(1, SeqCst);
                let from_client = from_client_tx.clone();
                tokio::spawn(async move {
                    while let Some(Ok(message)) = read.next().await {
                        if let Message::Binary(data) = message {
                            let _ = from_client.send(data.to_vec());
                        }
                    }
                });
            }
        });
    }

    MockRelay {
        url,
        to_client,
        from_client: TokioMutex::new(from_client_rx),
        drop_signal,
        conn_count,
    }
}

impl MockRelay {
    pub(crate) fn push(&self, msg_type: MsgType, payload: &[u8]) {
        self.to_client
            .send(encode_message(msg_type, payload))
            .unwrap();
    }

    pub(crate) async fn recv(&self, timeout: Duration) -> Option<Vec<u8>> {
        let mut receiver = self.from_client.lock().await;
        tokio::time::timeout(timeout, receiver.recv())
            .await
            .ok()
            .flatten()
    }
}

pub(crate) fn make_output_cb() -> (OutputCallback, Arc<Mutex<Vec<Vec<u8>>>>) {
    let received = Arc::new(Mutex::new(Vec::new()));
    let output = received.clone();
    let callback: OutputCallback = Arc::new(move |data| output.lock().unwrap().push(data));
    (callback, received)
}

pub(crate) async fn wait_until<F: Fn() -> bool>(condition: F, timeout: Duration) -> bool {
    let start = std::time::Instant::now();
    while !condition() {
        if start.elapsed() > timeout {
            return false;
        }
        tokio::time::sleep(Duration::from_millis(10)).await;
    }
    true
}

pub(crate) async fn wait_transport(mock_relay: &MockRelay) -> bool {
    wait_until(
        || mock_relay.conn_count.load(SeqCst) >= 1,
        Duration::from_secs(3),
    )
    .await
}

pub(crate) async fn wait_ready(pool: &RelayConnectionPool, pod_key: &str) -> bool {
    let start = std::time::Instant::now();
    while pool.get_status(pod_key).await != RelayStatus::Connected {
        if start.elapsed() > Duration::from_secs(3) {
            return false;
        }
        tokio::time::sleep(Duration::from_millis(10)).await;
    }
    true
}

pub(crate) fn buf_has(received: &Arc<Mutex<Vec<Vec<u8>>>>, expected: &[u8]) -> bool {
    received
        .lock()
        .unwrap()
        .iter()
        .any(|frame| frame == expected)
}
