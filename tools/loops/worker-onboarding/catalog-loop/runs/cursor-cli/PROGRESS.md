# Progress

Loop: Worker Onboarding Loop Template

## Current Status

- Status: blocked on a valid non-production Cursor credential bundle.
- Active loop node: `worker-goal`
- Active atomic task: `verify-worker-flow`
- Last verifier result: the browser blocked plan generation before any resource or Worker was created.
- Last no-progress fingerprint: `cursor-cli|missing-credential-reference|definition-v2`

## Verified

- Cursor uses executable `agent`, adapter `cursor-acp`, and ACP mode `agent acp`.
- Runner evidence shows ACP initialization completes, while `session/new` fails with
  `Authentication required` when no `CURSOR_API_KEY` is supplied.
- The AgentFile now declares `CURSOR_API_KEY` required and declares
  `CAPABILITY streaming false`, matching the runtime capability response.
- Backend validates required secret references. The Web form shows the required
  credential field, explains the missing reference, and disables plan generation.
- The attempted missing-credential browser path created neither an orchestration
  resource nor a Pod.
- Definition loading, projection sync, targeted Go tests, and the full Web test
  suite passed.

## Blocker

Terminal ACP verification requires a valid encrypted Cursor credential bundle.
No credential value was read, logged, or changed during this run.

## Next Cycle

1. Select an existing non-production Cursor credential bundle in the Worker form.
2. Generate and apply the Worker plan, then create the Worker.
3. Send a browser prompt and require a completed ACP session and agent reply.
4. Record terminal evidence before checking any acceptance item.
