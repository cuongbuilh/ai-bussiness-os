# Dual operating modes: Task-driven MVP / Goal-driven evolution

## 1. Mô hình sử dụng chung
- **Task-driven (P0)**: work-order từ người / AI Manager / schedule / business event; hợp lệ khi được policy authorize và nằm trong ngân sách. Có `origin_type`, `origin_ref`, `initiated_by`, `accountable_owner`, `trigger_run_id` tùy nguồn. Có thể có task độc lập hoặc dự án.
- **Goal-driven (P2 — Deferred)**: business owner tạo goal có phạm vi, KPIs, deadline, max spend, autonomy tier và reviewer. Agent được ủy quyền tạo nhiều phiên bản plan và initiative; mỗi initiative tạo hoặc gắn các project/task chuẩn thông qua control plane. Owner có thể pause/kill goal và tác vụ phát sinh.
- Chỉ có **một** task engine, budget engine, capability policy, approval engine, reporting schema, cost ledger và evaluation engine. Goal không tạo tool permission riêng; goal chỉ **thu hẹp**, không nâng trần quyền.

## 2. MVP task-driven: scope P0/P1
- P0: tạo task bằng web/API; AI Manager delegation qua work-order đã được phép; executor/sandbox/checkpoints; reviewer; board; ledger.
- P0 (basic): task source metadata và scheduler service phải nhận idempotency key / event ID; đăng ký trigger có thể nằm sau feature flag nếu chưa có UI. Bản triển khai P0 cần được phép tắt hoàn toàn schedule/event mới.
- P1: self-service scheduled/event trigger registration, trigger runs, preview/dry run, allowlisted event types, rate budget, dedup và UI trigger history.
- P1: initiative proposals không xuất phát từ goal: AI Manager phát hiện backlog/SLA và **đề xuất** task hoặc hiring/tuning; không tự biến đề xuất nhạy cảm thành quyết định.
- **Không thuộc MVP:** Autonomous goal decomposition, continuous replanning, strategy loops, long-horizon independent planning không cần work-order hợp lệ.

## 3. Goal-driven (P2): các entity và điều kiện
| Entity | Thuộc tính tối thiểu | Bắt buộc bảo vệ |
|---|---|---|
| `BusinessGoal` | owner, statement, metric targets/baseline, due, spend cap, status, autonomy tier | human approval khi kích hoạt mục tiêu chi tiền/chức năng cao rủi ro |
| `GoalMetric` | method, source, baseline, target, window, confidence | không dùng self-rated progress làm nguồn duy nhất |
| `GoalPlanVersion` | plan DAG, initiative list, budget forecast, model/prompt versions, author, reviewer | snapshot bất biến khi được duyệt |
| `Initiative` | goal, accountable manager, expected impact, budget envelope, status | chỉ tạo subtask qua Work API, có trace link |
| `GoalReview` | observed KPIs, risks, cost-to-date, next actions, reviewer | review định kỳ, no silent authority change |
| `GoalTaskLink` | goal, task, initiative, weight/provenance | một cost ledger fact vẫn chỉ được ghi một lần |

### Goal lifecycle đề xuất
`DRAFT -> PROPOSED -> APPROVED -> ACTIVE -> PAUSED -> ACTIVE -> ACHIEVED/CANCELLED/EXPIRED`.
`REPLAN_REQUIRED -> plan proposal -> approval (nếu thay target/budget/risk/authority) -> plan version activated`.

### Guardrails
1. Strategy planning không có tool credential và không tự authorize work.
2. Goal policy defines per-goal max spend, max task count, daily initiative cap, max delegation depth, reviewer cadence and data scope.
3. Replanning không bỏ qua policy approval. Agent không được đổi goal KPI để làm đẹp số liệu.
4. Pause/cancel goal propagates best-effort cancel to queued/waiting tasks; running external side effects need compensating/reconciliation, not unsafe automatic rollback.
5. Chỉ coi goal hoàn thành khi KPI nguồn đáng tin cậy thỏa điều kiện và authorized reviewer chấp nhận.

## 4. End-to-end flows
### Human assigned task
Web create -> auth/validate -> `task.created.v1` -> queue -> reserve -> execute model/tool -> evaluate -> human ACCEPTED -> ledger/report.

### Scheduled/event-triggered task
Verified trigger -> dedup `(trigger_id, occurrence_key/event_id)` -> preview/work order -> risk/approval gate -> create canonical task -> normal task flow. Suspend trigger on repeated errors/cost anomalies.

### Goal-driven future
Owner defines goal + KPI -> goal plan draft -> reviewer approves risk/budget -> initiative/project/tasks created -> work engine executes -> KPI evaluator collects evidence -> review/approve replan -> goal accepted/archived.

## 5. Release definition
- Task-driven MVP passing the applicable P0 tests (incl. AT-27/28/29/31/34) does not claim Goal-driven shipped; full trigger and HR proposal test suites run at P1.
- Goal-driven only releasable after AT-GOAL-* (approval, authority inheritance, plan version, linkage, stop conditions, spend accounting, KPI acceptance) in CI.
