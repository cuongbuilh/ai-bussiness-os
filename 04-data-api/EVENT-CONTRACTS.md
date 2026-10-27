# Event contracts and reliability

## Envelope
```json
{
  "eventId": "UUID",
  "eventType": "task.assigned.v1",
  "occurredAt": "2026-10-27T05:00:00Z",
  "aggregateId": "TASK_UUID",
  "aggregateVersion": 12,
  "organizationId": "ORG_UUID",
  "actor": {"type": "HUMAN", "employeeId": "EMP_UUID"},
  "correlationId": "TRACE_OR_REQUEST_ID",
  "causationId": "COMMAND_ID",
  "payload": {"taskId": "TASK_UUID", "ownerEmployeeId": "EMP_UUID", "projectId": null}
}
```

## Proposed domain events
- `employee.created.v1`, `employee.manager.changed.v1`, `employee.suspended.v1`
- `capability.grant.changed.v1`, `mcp.server.updated.v1`, `skill.version.activated.v1`
- `project.created.v1`, `project.status.changed.v1`, `milestone.accepted.v1`
- `task.created.v1`, `task.assigned.v1`, `task.status.changed.v1`, `task.accepted.v1`, `task.blocked.v1`
- `execution.started.v1`, `execution.checkpointed.v1`, `execution.completed.v1`, `execution.failed.v1`
- `tool.invocation.requested.v1`, `tool.invocation.allowed.v1`, `tool.invocation.denied.v1`
- `budget.reserved.v1`, `budget.settled.v1`, `budget.threshold.reached.v1`
- `approval.requested.v1`, `approval.decided.v1`, `model.change.proposed.v1`, `model.assignment.activated.v1`

## Additional V1.1 events
- `hr.change.proposed.v1`, `hr.change.approved.v1`, `hr.change.applied.v1` (separate approval vs apply)
- `task.proposed.v1`, `task.created.from-trigger.v1`, `trigger.occurrence.received.v1`, `trigger.occurrence.deduplicated.v1`
- `memory.candidate.created.v1`, `memory.verified.v1`, `memory.revoked.v1`, `memory.access.denied.v1`
- `budget.threshold.80.reached.v1`, `budget.topup.requested.v1`, `budget.topup.applied.v1`
- `optimization.proposed.v1`, `optimization.approved.v1`, `optimization.rolledback.v1`
- Future only: `goal.proposed.v1`, `goal.plan.activated.v1`, `initiative.created.v1`, `goal.paused.v1`.

**Idempotency:** A trigger occurrence is uniquely keyed by `(trigger_id, occurrence_key)` even if received events have different broker delivery IDs. A budget-threshold message is unique by `(budget_account_id, window, threshold)`; dispatch could repeat but downstream notification is idempotent.

**Sensitive data:** Proposal payloads reference reviewable sanitized diffs/artifacts, not raw skill scripts, API secrets or confidential memory contents. Event recipients re-evaluate scope instead of trusting event body.

## Broker/runtime rules
1. Outbox inserted in same database transaction as the business change.
2. Publisher asynchronously sends outbox events with message identifier; may re-send on crash.
3. Consumer deduplicates by `(event_id, handler)`; side-effect commands also require their own idempotency key.
4. Per-aggregate ordering requires aggregate version check and delayed retry on future version.
5. Poison messages go to dead-letter queue, page owner; no silently dropped work.
6. UI SSE stream may replay by Last-Event-ID and re-fetch board snapshot; never assume every SSE event received.
7. Sensitive payload references are stored as artifact IDs, not raw API keys, prompts or external tool secrets.
8. Scheduled expiration of approvals/reservations emits auditable state-change events.

## Example web stream
```
id: evt_0201
event: task.status.changed
data: {"taskId":"T-42","from":"READY","to":"IN_PROGRESS","version":13}

id: evt_0202
event: usage.settled
data: {"taskId":"T-42","amountUsd":"0.00230000","tokens":900,"ledgerId":"..."}
```
The frontend should treat events as invalidation hints and query authoritative current state where precision matters.
