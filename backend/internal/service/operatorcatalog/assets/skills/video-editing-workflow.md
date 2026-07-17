# Video Editing Workflow

Edit footage through an explicit, reversible decision list.

## Workflow

1. Inventory every source with `ffprobe`; preserve originals.
2. Transcribe spoken material with word-level timestamps when speech drives the cut.
3. Propose the narrative and pacing strategy before changing media.
4. Build an EDL with source, in, out, output order, transition, audio treatment, and reason.
5. Cut on word or action boundaries. Add short audio fades at hard joins to prevent clicks.
6. Normalize dialogue before music. Duck music around speech instead of raising overall loudness.
7. Add overlays before subtitles. Burn subtitles last so graphics cannot cover them.
8. Render a low-resolution preview, inspect every cut boundary, then render the delivery master.

## Correctness Rules

- Never overwrite source media.
- Avoid repeated lossy encoding; extract and concatenate losslessly when codecs permit.
- Keep subtitle timestamps on the output timeline after cuts.
- Use `setpts=PTS-STARTPTS+offset/TB` for time-shifted overlays.
- Verify duration, dimensions, frame rate, codecs, audio channels, and stream presence with `ffprobe`.
- Stop after three failed self-review passes and report the remaining defects.

Persist the EDL, subtitle source, render command, and review notes beside the output.
