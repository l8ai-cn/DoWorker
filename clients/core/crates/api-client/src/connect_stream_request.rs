use prost::Message;

use crate::ApiError;

pub(crate) fn frame_connect_stream_request(request: &impl Message) -> Result<Vec<u8>, ApiError> {
    let payload = request.encode_to_vec();
    let length = u32::try_from(payload.len()).map_err(|_| {
        ApiError::Decode(format!(
            "connect stream request too large: {}",
            payload.len()
        ))
    })?;
    let mut frame = Vec::with_capacity(5 + payload.len());
    frame.push(0);
    frame.extend_from_slice(&length.to_be_bytes());
    frame.extend(payload);
    Ok(frame)
}
