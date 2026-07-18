---
name: seedance-expert
description: Generate a Seedance video with the Worker-bound video model and publish a previewable MP4 artifact.
---

# Seedance Expert

Use this skill when the user asks to create a video with Seedance.

## Required environment

- `SEEDANCE_API_KEY`
- `SEEDANCE_BASE_URL`
- `SEEDANCE_MODEL`

Stop with a clear error if any required value is missing. Never print credentials.

## Workflow

1. Turn the user's request into one concise generation prompt. Preserve requested subject, action, camera movement, style, duration, aspect ratio, and exclusions.
2. Submit exactly one generation request to `${SEEDANCE_BASE_URL}/contents/generations/tasks` with bearer authentication. Use `SEEDANCE_MODEL`, one text `content` item, `duration: 5`, `ratio: "16:9"`, `resolution: "720p"`, `generate_audio: true`, and `watermark: false`.
3. Persist the returned task ID. Poll that same task until it succeeds or fails. Do not submit another request while a task is pending.
4. On success, download `content.video_url` to `artifacts/seedance-video.mp4`.
5. Verify the file is non-empty and reports an MP4-compatible media type.
6. Call `workbench.publish_artifact` with a video manifest whose playable representation is `artifacts/seedance-video.mp4`, media type is `video/mp4`, and producer is `{ "namespace": "seedance", "type": "video.generate", "id": "<provider-task-id>" }`.
7. Report success only after publication returns a ready artifact whose playable representation has a positive byte size and SHA-256 digest.
8. If generation succeeds but publication or verification fails, report a partial result. Never describe the task as successful without the verified artifact.

If the provider returns a non-success response, report the status and sanitized response body without retrying billable creation.
The provider task ID is execution metadata supplied by this skill. Platform artifact verification proves the published file's integrity, not independent provider attestation.
