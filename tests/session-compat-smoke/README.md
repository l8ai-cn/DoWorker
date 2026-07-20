# Session Compatibility Smoke Scripts

API / browser smoke suites for the session compatibility stack.

Run via:

```bash
bash deploy/dev/session_compat_smoke.sh
```

Scripts write screenshots and JSON reports under repo-root `output/` (gitignored scratch).

Each suite creates and deletes only its own session fixtures through `DELETE
/v1/sessions/{id}`. The harness never terminates every live pod in `dev-org`,
because that shared organization can contain other active development work.
