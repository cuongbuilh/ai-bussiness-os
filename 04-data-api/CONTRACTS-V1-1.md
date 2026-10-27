# V1.1 contracts delta — task sources, HR, memory, budget, optimization, goals

Các interfaces bên dưới là **starter design** bổ sung cho `openapi-starter.yaml` và `schema-starter.sql`, chưa đủ security/migrations/DTOs production. AuthZ luôn kiểm tra server-side.

## 1. Operational API contracts
| Method | Path | Purpose / outcome |
|---|---|---|
| POST | `/hr/change-proposals` | AI/human tạo proposal `HIRE_AI, ASSIGN_POSITION, CHANGE_MANAGER, REORGANIZE` → PENDING_APPROVAL |
| GET | `/hr/change-proposals` | Scope-filtered proposals cho approver/manager |
| POST | `/triggers` | Tạo schedule/event definition với owner, cap, mode, dedup policy; sensitivity gate |
| POST | `/triggers/{id}/dry-run` | Preview task/work-order and budgets, no side effects |
| GET | `/triggers/{id}/runs` | Cron/event occurrence, status, dedup and associated task IDs |
| POST | `/task-proposals` | Đề xuất task ngoài current delegated work; materialize only after approval if necessary |
| GET | `/memory/items?scopeType&scopeId` | Authorized filtered memory entries; default exclude nonverified |
| POST | `/memory/items` | Candidate memory write, version/provenance/classification |
| POST | `/memory/items/{id}/verify` | Reviewer verification; do not trust actor's self-rating alone |
| POST | `/budgets/{id}/top-up-requests` | Ask human for budget increase; does not change effective cap |
| GET | `/budgets/{id}/usage` | settled + reserved + percentage + warning thresholds + trace |
| POST | `/optimization/proposals` | PROMPT/SKILL/MODEL/PROVIDER changes with evidence & rollback version |
| GET | `/optimization/proposals` | Filter by employee, model, position, status, score window |
| GET | `/employees/{id}/memory-access` | Effective namespaces, rules, reason codes; never reveal restricted content |

API writes require idempotency key when creating external events / applying approvals; resource version checks require `If-Match`. Safe error surface: `409 CONFLICT`, `403 FORBIDDEN`, `422 POLICY_BLOCK`, `429 RATE_LIMIT`, `402/409 BUDGET_EXHAUSTED` (choose one domain error contract; avoid relying on HTTP 402 alone). Standardize `application/problem+json` with RFC 9457 for errors as proposed.

## 2. Future Goal-driven interfaces (NOT MVP)
- `POST /goals`, `GET /goals`, `GET /goals/{id}`.
- `POST /goals/{id}/plan-proposals`, `POST /goals/{id}/reviews`, `POST /goals/{id}/pause`, `POST /goals/{id}/resume`.
- `POST /initiatives`, `GET /goals/{id}/initiatives`, `GET /goals/{id}/tasks`.
- `GET /goals/{id}/kpis` returns versioned KPI/evidence and last observation; `POST /goals/{id}/accept` human gate.

## 3. Workflow examples
### Hiring
`employee recommendation -> hr.change-proposal.created -> approval.required -> human APPROVE -> revalidate budget/permissions/org tree -> employee created DRAFT -> activation`.

### Triggered task
`trusted event/scheduler tick -> trigger occurrence registry INSERT UNIQUE -> evaluate enable/cap/risk -> proposal or canonical task creation -> task.created -> normal execution`.

### Budget top-up
`budget warning 80% -> AI manager request -> pending approval -> human approve -> budget policy revision -> event + ledger remains intact`.

### Improvement
`task eval -> model/prompt/skill suggestion -> frozen benchmark dataset, policy compare -> approver decision -> versioned model/skill/prompt deployment -> canary/rollback`.

## 4. Event additions
`hr.change.proposed.v1`, `hr.change.approved.v1`, `hr.change.applied.v1`, `trigger.occurrence.received.v1`, `trigger.occurrence.deduplicated.v1`, `task.proposed.v1`, `task.created.from-trigger.v1`, `memory.candidate.created.v1`, `memory.verified.v1`, `memory.access.denied.v1`, `budget.threshold.80.reached.v1`, `budget.topup.requested.v1`, `budget.topup.applied.v1`, `optimization.proposed.v1`, `optimization.approved.v1`, `optimization.rolledback.v1`.

Future: `goal.proposed.v1`, `goal.approved.v1`, `goal.plan.activated.v1`, `initiative.created.v1`, `goal.reviewed.v1`, `goal.paused.v1`.
