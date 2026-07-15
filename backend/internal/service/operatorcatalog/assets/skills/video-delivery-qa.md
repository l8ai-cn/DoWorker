# Video Delivery QA

Validate the rendered artifact before delivery.

## Required Checks

1. Use `ffprobe` to verify duration, dimensions, display aspect ratio, frame rate, video codec, audio codec, sample rate, channels, and stream count.
2. Compare duration against the approved script or EDL tolerance.
3. Inspect the first two seconds, last two seconds, every cut boundary, and representative middle frames.
4. Check subtitle spelling, timing, line breaks, safe margins, contrast, and overlap with graphics.
5. Listen across every edit for clicks, clipped speech, abrupt ambience, music masking, and channel imbalance.
6. Confirm no placeholder, debug overlay, missing font, offline asset, black frame, or stale revision remains.
7. Verify the delivery filename, container, aspect ratio, and size against the target platform.

## Result

Return a pass or fail report with:

- measured media properties
- failed checks with timestamps
- evidence frame paths
- exact rerender action
- final artifact checksum

Never mark a delivery passed from the render command exit code alone.
