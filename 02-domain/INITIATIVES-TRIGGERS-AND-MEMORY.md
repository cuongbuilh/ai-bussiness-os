# Proactive work, HR/change proposals, memory policies

## A. HR / hiring proposal workflow
**BR-HR-001** AI Manager được tạo đề xuất tuyển AI, bổ nhiệm, đổi vị trí/manager/cấu trúc org: reason, capacity-gap evidence, candidate position, proposed skills/MCP/model, anticipated recurring/max cost, budget source and ROI hypotheses.
**BR-HR-002** Đề xuất không tạo employee ACTIVE hoặc position grant có hiệu lực trước human approval và các policy gate. Cấm requester/beneficiary AI tự quyết; không được dùng proxy-human làm người duyệt thay.
**BR-HR-003** Approval tiêu thụ/giữ chỗ capacity budget theo policy và kiểm tra manager tree, org quotas ngay lúc apply (TOCTOU), idempotent apply once. AI không thể tự tạo/đổi người duyệt.
`DRAFT -> SUBMITTED -> REVIEWING -> APPROVED|REJECTED|EXPIRED -> APPLIED` (APPLIED chỉ từ APPROVED).

## B. Trigger-driven proactive task
**BR-TRIG-001** Nguồn task: `HUMAN`, `MANAGER`, `SCHEDULE`, `EVENT`, sau này `GOAL`; mỗi task có origin_type + provenance + owner + creator + criteria + budget policy.
**BR-TRIG-002** Trigger phải có owner nhân sự/người, scope, allowlisted event or cron, max tasks/window, max estimated spend/window, expiration, mode `PROPOSE_ONLY|AUTO_CREATE_IN_POLICY`, enabled flag và quarantine behavior.
**BR-TRIG-003** Schedule/event dedup `(trigger_id, occurrence_key)`; một occurrence không tạo task lặp khi broker retry, race hoặc reload.
**BR-TRIG-004** Agent được phép tự **đề xuất** task ngoài scope ban đầu; chỉ được auto-create task khi trigger/delegation policy cho phép, có hard budget và không cần sensitive approval. Work không có approval đi theo review queue.
**BR-TRIG-005** Không để event/tool output cấp thêm quyền hoặc tự tạo trigger rộng hơn; các thay đổi trigger mức nhạy cảm do người có thẩm quyền duyệt.

## C. Collaboration & ownership
**BR-COL-001** Trước READY task có chính xác 1 accountable owner; 0..N collaborators/reviewers/actual executors, mỗi contributor có giới hạn quyền và ngân sách task riêng.
**BR-COL-002** Manager lập plan/subtask phải giữ parent/root task trace, max depth, WIP/time/token/cost guard; subtask và execution không được coi là 2 công việc độc lập khi aggregate parent progress.
**BR-COL-003** Delegation hay role collaboration không tạo quyền truy cập toàn bộ project/org/memory; tính quyền theo mỗi work item.
**BR-COL-004** Cost per task = một tập ledger charge không trùng; phân bổ team/project cost-center theo immutable primary attribution, analytics collaborator chỉ là slice chứ không phải cộng gộp.

## D. MCP/skills permission ceiling
**BR-CAP-001** Effective tool allowance = org policy ∩ position ceiling ∩ agent grants ∩ task scope ∩ data/network/policy ∩ valid approval ∩ model runtime support (applicability, not authority).
**BR-CAP-002** Skill binding = versioned skill plus required tool prerequisites; skill instructions không cấp auth, credentials, tăng limits hoặc thay đổi model.
**BR-CAP-003** Position role policy đổi => active grants phải re-evaluate at tool-call time, không cho phép đã cache quyền cũ tiếp tục hành động.
**BR-CAP-004** Scoped exception elevation chỉ có hiệu lực khi human approval hợp lệ, with expiry, subject/target/operation bind and audit; default deny.

## E. Persistent memory
**BR-MEM-001** Memory có `scope_type` từ `EMPLOYEE | TEAM | DEPARTMENT | PROJECT | ORGANIZATION` và `scope_id`, classification, provenance, author, content_ref, TTL/retention, revision, status.
**BR-MEM-002** Quyền đọc/ghi được kiểm tra theo user/agent, membership/capability/task, classification và purpose. Không mặc định suy ra một employee được đọc mọi memory của manager hoặc phòng khác.
**BR-MEM-003** Tạo memory từ kết quả task chưa biến thành 'fact' tin cậy: `CANDIDATE -> VERIFIED/REJECTED -> ACTIVE`, hỗ trợ nguồn/citation và invalidation.
**BR-MEM-004** Thay model hoặc chuyển position/manager giữ identity/historical memory, nhưng effective read scopes re-evaluate và có thể revoke; không đưa long-term secret/PII vào prompt mặc định.
**BR-MEM-005** Delete/redaction/retention nghiệp vụ, source revocation và vector index synchronization phải cùng thực hiện có audit. Policy này áp dụng với cả embeddings/cache.

## F. Performance-driven improvement
**BR-OPT-001** Agent/Manager có thể tạo proposal `PROMPT | SKILL | MODEL | PROVIDER | TOOL_POLICY` kèm baseline, evidence, dataset, ảnh hưởng role/task, projected cost/risk và rollback version.
**BR-OPT-002** Thay đổi đáng kể (system instructions, shared skill activation, primary model, tool privileges, security policy) phải human review/approval; minor per-task draft không tăng quyền có thể dùng trong scope.
**BR-OPT-003** Compare bằng fixed evaluator/dataset hoặc canary cohort; account tool/retry/latency/safety, sample size; không tự chỉnh chuẩn KPI hoặc xóa negative traces.
**BR-OPT-004** Rollback/disable on regression; version pin đủ prompt, model, skill, MCP schema và evaluator trên mỗi execution.

## G. Budget threshold
**BR-BUD-006** 80% warning khi `(settled + active reserved)/hard_cap >= 0.8` trên scope/window applicable; unique threshold event per budget window, các lần gọi sau vẫn được phép nếu chưa chạm hard stop.
**BR-BUD-007** Hard limit deny/reschedule; AI Manager chỉ `REQUEST_TOP_UP`, không tự chỉnh quota, trực tiếp hoặc qua change proposal/child agent.
**BR-BUD-008** Top-up apply idempotently after authorized human approval, performs parent allocation check and creates new budget policy revision; preserve ledger history.

## H. Approval and authority matrix
| Action | AI employee | AI Manager | Human Manager | Founder/Admin |
|---|---|---|---|---|
| Create in-scope task | If delegated | If delegated | Within scope | Yes |
| Propose hiring/transfer | Proposal | Proposal | Proposal / approval if delegated | Approve/apply |
| Execute R0/R1 tools | On valid grants | On valid grants | N/A | Policy admin |
| Execute R3/R4 side effects | Request approval | Request approval | Approve if explicitly authorized | Approve if policy allows |
| Budget increase | Request | Request | Approve only with finance authority | Approve/apply |
| Change primary model | Request | Request | Approve only if policy allows | Approve/apply |
| Change tool ceiling | Request | Request | Restricted approval | Approve/apply |

**Rule:** Tách quyền `propose`, `approve`, `apply`. Một approval không thay cho authorization tại thời điểm apply.
