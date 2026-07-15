# Remotion Video Production

Build deterministic React-based video compositions and render them with Remotion.

## Workflow

1. Define composition width, height, frame rate, duration, input schema, and delivery format.
2. Keep timing data separate from visual components. Derive frames from seconds through the composition frame rate.
3. Preload fonts and media. Fail early when required assets are absent.
4. Build scenes as module-level components with stable dimensions and explicit start and duration frames.
5. Use `Sequence`, interpolation, and spring motion with bounded input and output ranges.
6. Keep text inside safe areas for vertical, square, and horizontal variants.
7. Render a representative frame and a short preview before the full composition.
8. Render the master, then verify streams and dimensions with `ffprobe`.

## Quality Rules

- No network fetches during final rendering.
- No viewport-dependent layout.
- No unbounded random values; seed any procedural variation.
- No overlapping captions, controls, or brand marks.
- Respect reduced-motion requirements for preview interfaces.
- Keep a render manifest containing composition ID, props, frame rate, duration, and command.
