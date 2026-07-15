# Blocked Execution Decisions

Loop: Worker Onboarding Loop Template

## User Confirmation

- Required before delegation: true
- Prompt: May the decision proxy make only the listed low-risk sequencing and public-metadata retry decisions?
- Confirmation record: `PROGRESS.md`

## Proxy Decision Agent

- Agent: `decision-proxy`
- Authority: `delegated_low_risk`
- Default when uncertain: `ask_user`

Allowed decisions:

- choose next eligible task
- split one task smaller
- retry one public metadata check

Forbidden decisions:

- irreversible action
- production deploy
- credential change
- license acceptance

Decision records must include:

- `blocked_reason`
- `options`
- `selected_option`
- `rationale`
- `evidence_ref`

## Blocked Handling

Blocked signals:

- unknown research conclusion
- license uncertainty
- real credential requirement
- three identical fingerprints
- verifier protection conflict

- Max blocked cycles: 1

Allowed resolution actions:

- record evidence
- request human decision
- process independent non-blocked work

Escalate when:

- the Worker cannot reach terminal verification
- a protected contract requires change
- a security or public API decision is needed

## Supervisor

- Agent: `loop-supervisor`
- Review cadence iterations: 2
- Report path: `monitoring-plan.json`

Drift checks:

- selected Worker remains the worker.json slug
- active task maps to Worker acceptance
- shared contract work has explicit ownership
- no human gate is silently crossed

Intervention actions:

- pause Worker loop
- reopen invalid acceptance
- escalate to human
