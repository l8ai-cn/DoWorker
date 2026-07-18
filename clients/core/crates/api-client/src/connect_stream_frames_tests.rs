use agentsmesh_types::proto_events_v1::Event;
use bytes::Bytes;
use futures::stream::{self, Stream, StreamExt};
use prost::Message;

use crate::connect_stream_frames::parse_connect_frames;
use crate::ApiError;

const FLAG_COMPRESSED: u8 = 1;
const FLAG_FINAL: u8 = 2;

#[tokio::test]
async fn single_message_then_clean_end() {
    let message = Event {
        r#type: "pod:status_changed".into(),
        ..Default::default()
    };
    let chunks = vec![frame(0, &message.encode_to_vec()), frame(FLAG_FINAL, b"{}")];
    let stream = parse_connect_frames::<_, Event>(ok_stream(chunks));
    futures::pin_mut!(stream);

    assert_eq!(
        stream.next().await.unwrap().unwrap().r#type,
        "pod:status_changed"
    );
    assert!(stream.next().await.is_none());
}

#[tokio::test]
async fn handles_partial_chunks_across_boundary() {
    let message = Event {
        r#type: "ticket:updated".into(),
        ..Default::default()
    };
    let full = frame(0, &message.encode_to_vec());
    let mid = full.len() / 2;
    let chunks = vec![
        full.slice(..mid),
        full.slice(mid..),
        frame(FLAG_FINAL, b"{}"),
    ];
    let stream = parse_connect_frames::<_, Event>(ok_stream(chunks));
    futures::pin_mut!(stream);

    assert_eq!(
        stream.next().await.unwrap().unwrap().r#type,
        "ticket:updated"
    );
}

#[tokio::test]
async fn surfaces_end_stream_error() {
    let payload = br#"{"error":{"code":"unauthenticated","message":"token expired"}}"#;
    let stream = parse_connect_frames::<_, Event>(ok_stream(vec![frame(FLAG_FINAL, payload)]));
    futures::pin_mut!(stream);

    match stream.next().await.unwrap().unwrap_err() {
        ApiError::Http { code, .. } => assert_eq!(code.as_deref(), Some("unauthenticated")),
        other => panic!("wrong error variant: {other:?}"),
    }
}

#[tokio::test]
async fn rejects_compressed_frames() {
    let stream =
        parse_connect_frames::<_, Event>(ok_stream(vec![frame(FLAG_COMPRESSED, b"\x01\x02\x03")]));
    futures::pin_mut!(stream);

    assert!(matches!(
        stream.next().await.unwrap().unwrap_err(),
        ApiError::Decode(_)
    ));
}

#[tokio::test]
async fn eof_without_final_frame_is_an_error() {
    let message = Event {
        r#type: "ticket:updated".into(),
        ..Default::default()
    };
    let stream =
        parse_connect_frames::<_, Event>(ok_stream(vec![frame(0, &message.encode_to_vec())]));
    futures::pin_mut!(stream);

    assert_eq!(
        stream.next().await.unwrap().unwrap().r#type,
        "ticket:updated"
    );
    let error = stream.next().await.unwrap().unwrap_err();
    assert!(error.to_string().contains("final frame"));
}

#[tokio::test]
async fn eof_with_partial_frame_reports_residual_buffer() {
    let stream = parse_connect_frames::<_, Event>(ok_stream(vec![Bytes::from_static(&[0, 0, 0])]));
    futures::pin_mut!(stream);

    let error = stream.next().await.unwrap().unwrap_err();
    assert!(error.to_string().contains("incomplete frame"));
    assert!(error.to_string().contains("3 buffered bytes"));
}

fn frame(flags: u8, payload: &[u8]) -> Bytes {
    let mut value = Vec::with_capacity(5 + payload.len());
    value.push(flags);
    value.extend_from_slice(&(payload.len() as u32).to_be_bytes());
    value.extend_from_slice(payload);
    Bytes::from(value)
}

fn ok_stream(items: Vec<Bytes>) -> impl Stream<Item = Result<Bytes, ApiError>> + Unpin {
    Box::pin(stream::iter(items.into_iter().map(Ok)))
}
