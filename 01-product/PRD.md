# PRD — AI Business OS

## 1. Problem, vision, measurable outcomes
**Problem:** Nhân sự AI hiện phân tán trong nhiều tools, không có khái niệm người quản lý, role/job title, ngân sách cá nhân và dự án, chất lượng công việc, hoặc audit từ business objective đến model execution.

**Vision:** Tạo một công ty AI/hybrid trên web: Founder khởi tạo organization, tạo position/role, thuê hoặc bổ nhiệm agent, cấp MCP/skills/model/budget, giao dự án/task; agent tự xử lý trong quyền và báo cáo; manager thấy portfolio/board; platform đo hiệu quả và đề xuất tối ưu model.

**North-star KPI:** Số **accepted work items** có bằng chứng hoàn thành và tổng chi phí đầy đủ (cost per accepted work item); KPI đi kèm: acceptance %, human rework minutes, SLO hit rate, budget overrun count=0, unsafe tool call allowed=0.

## 2. Personas
- **Founder / Organization Admin:** Tạo org, phê duyệt role/policy/model changes, xem toàn bộ dashboard, cấu hình provider keys.
- **Human Manager:** Quản lý reportees, project portfolio, phân công/đánh giá, chỉ duyệt trong phạm vi được ủy quyền.
- **AI Manager:** Plan, delegate, aggregate, escalate; bị chặn bởi quyền, quota, depth, no self-approval.
- **AI Employee:** Học skills, dùng MCP/tools được phép, hoàn thành task, phản hồi review, không tự cấp quyền.
- **Project Owner / PM:** Quản lý scope, milestone, backlog, dependencies, budget và nghiệm thu.
- **Reviewer / Compliance:** Audit, kiểm tra outputs, review sensitive actions/model changes.

## 3. Functional requirements (FR)
| ID | Requirement | MVP |
|---|---|---|
| FR-ORG-001 | Tạo phòng ban, team, position, role/title và thay đổi sơ đồ tổ chức | P0 |
| FR-ORG-002 | Employee human/AI; 0..1 direct manager; không vòng lặp; lịch sử điều chuyển | P0 |
| FR-ORG-003 | Employee lifecycle ACTIVE/SUSPENDED/RETIRED, profile giữ dữ liệu | P0 |
| FR-TOOL-001 | Đăng ký MCP server, tools, resources, prompts; phiên bản capability | P0 |
| FR-TOOL-002 | Position baseline tool/MCP/skills; agent grants/exceptions; task scope; runtime authorization | P0 |
| FR-TOOL-003 | Hiển thị tool call request/approval/response/errors/cost, che secrets | P0 |
| FR-MOD-001 | Central provider/model registry & encrypted API credential references | P0 |
| FR-MOD-002 | Capability filter tool-call/context/structured output + fallback policy | P0 |
| FR-MOD-003 | Dataset benchmark, model comparison, change proposal + approve + rollback | P1 |
| FR-BUD-001 | Limits USD/token/RPM/TPM/time/task, per company/department/team/employee/project/task | P0 |
| FR-BUD-002 | Atomic reserve/settle/expire, immutable ledger, alerts, hard/soft stops | P0 |
| FR-WORK-001 | Portfolio projects: owner, members, departments/teams, milestones, budget, risks | P0 |
| FR-WORK-002 | Task backlog, priorities, subtasks, dependencies, assignments, criteria, deadlines | P0 |
| FR-WORK-003 | One accountable owner, collaborators, executing agent and review/acceptance distinct | P0 |
| FR-WORK-004 | Unified board scopes: company, department, team, project, employee, manager, role | P0 |
| FR-WORK-005 | Kanban, table, timeline, calendar; filters, saved views, drill-down | Kanban/table P0; timeline/calendar P1 |
| FR-WORK-006 | Web assign task / reassign / comment / add artifact / review / approval | P0 |
| FR-AI-001 | Agent orchestration, task state, checkpoint, delegation, events | P0 |
| FR-AI-002 | Role-scoped memory and tool execution sandbox; retry and escalation | P0 |
| FR-AI-003 | Scheduled/event-triggered initiative inside authorization/budget | P1 |
| FR-EVAL-001 | KPI per role, objective evidence, feedback & evaluation metadata | P0 |
| FR-EVAL-002 | Model/provider performance benchmarks & recommendation score | P1 |
| FR-UI-001 | Clickable org chart showing title/role/manager/status/agent identity | P0 |
| FR-UI-002 | Employee drawer: identity, direct manager, current tasks, MCP/skills, budget, KPI, trace | P0 |
| FR-UI-003 | Company/project/team/employee overview dashboard and live execution feed | P0 |
| FR-SEC-001 | Human login, RBAC+resource-scope ABAC, approvals & append-only audit | P0 |

### V1.1 confirmed functional requirements
| ID | Requirement | Delivery stage |
|---|---|---|
| FR-HR-001 | AI Manager tạo proposal tuyển/bổ nhiệm/chuyển quản lý/cơ cấu; human approve và apply an toàn | P1, basic proposal path P0 |
| FR-TRIG-001 | Lưu task provenance từ HUMAN/MANAGER/SCHEDULE/EVENT; trigger permission & dedup | P0 provenance; P1 trigger management |
| FR-TRIG-002 | Agent tạo in-scope task, ngoài scope chỉ đề xuất; cron/event limits | P1 |
| FR-COL-001 | Multi-agent collaborators/subtasks và một accountable owner; tránh double-count | P0 |
| FR-CAP-001 | Position ceiling, individual MCP/tool/skill config; exceptions require approval | P0 |
| FR-MEM-001 | Long-term memory Employee/Team/Department/Project/Organization ACL + provenance | P0 simple namespaces; P1 curation/TTL |
| FR-BUD-003 | 80% threshold event once/window, hard stop; request budget top-up via human approval | P0 |
| FR-OPT-001 | Evidence-based prompt/skill/model suggestions; impactful changes human approved and versioned | P1 |
| FR-GOAL-001 | Goal, measurable KPI, initiative/plan version, continuous review & bounded replan | P2 **Deferred** |

**Operating modes:** Task-driven is MVP deliverable; Goal-driven is target architecture, not a feature acceptance gate for MVP. Task-generated work (schedule/event) is still task-driven if bounded by an existing grant/trigger/work order. See `TASK-AND-GOAL-ROADMAP.md`.

## 4. User journeys
1. **Bootstrap Solo:** Admin adds provider -> validates model/tool-calling -> creates Research Analyst position with approved web MCP & research skill -> creates AI employee -> gives $5 and 100k token cap -> assigns research task -> inspects output, call trace, cost and approves.
2. **Scale Team:** Founder creates department/team, AI Manager and several agents -> reporting lines -> manager can delegate within allowed child budget -> board shows WIP per person, department and project.
3. **Cross Department Project:** One project spans Sales and Technology teams. Project owner sees all linked tasks; department heads see department-attributed tasks only; company view de-duplicates work items and ledger.
4. **Optimizer:** evaluation compares candidate model against frozen dataset and live canary -> AI Manager creates change request -> human approves -> new model assignment is versioned -> monitor and roll back if SLO regresses.
5. **Approval:** An agent requests privileged tool action. Engine checkpoints work, approval inbox shows diff/target/impact/estimated cost; reject resumes with rejection context and audit.

6. **Hiring capacity:** Manager detects work backlog -> drafts AI hire proposal -> human reviews spend, position, tools and authority -> only approved proposal may create/activate employee.
7. **Proactive task:** Approved schedule/event creates a single task per occurrence, respect quota and allowlist; high-risk action creates approval request.
8. **Memory:** Researcher stores candidate finding with citations in Project scope; another project's agent cannot read without grant.
9. **Optimization:** Agent proposes skill/prompt/model improvement using frozen dataset; approval and versioned activation; no self-approval.
10. **Future goal:** Owner sets measurable business goal; AI Manager drafts initiative plan and requests activation approval. This is **not MVP**.

## 5. Explicit exclusions from initial MVP
- Public signup/multi-org billing & marketplace.
- Employment payroll or legal HR records for human employees.
- Generic AI agents unrestricted browser/shell permissions.
- Completely autonomous employment decisions or production deployment. AI-generated hiring proposals and delegated task creation **are** allowed within approval/policy controls.
- Model auto-switch without human approval.
- Full Gantt resource scheduling, complex matrix org, advanced finance chargebacks.

## 6. Nonfunctional requirements (NFR) — proposed test targets
- NFR-SEC-001: 100% privileged tool calls checked server-side; secrets absent from prompts/logs.
- NFR-REL-001: at-least-once queue delivery with dedup; replayable runs, no duplicated irreversible external effects.
- NFR-BUD-001: concurrent reserve never allows reserved+settled to exceed hard cap based on authorized ledger.
- NFR-UI-001: company/project/team/person filters consistently reconcile same underlying task IDs.
- NFR-UI-002: 95th percentile board read < 1.5s on 10,000 tasks/1,000 employees under defined staging load (target, not measured).
- NFR-OPS-001: SLO targets to confirm through load tests; traces carry org/employee/task/run/model IDs.
- NFR-SEC-002: all risky tool actions either denied or approved by appropriate human with expiry.

## 7. Data semantics & reporting
- **Assigned owner:** one accountable employee for a task (may be AI or human).
- **Current executor:** runtime agent doing one execution; may differ from owner when delegated.
- **Project ownership:** one owner; projects can have multi-team members.
- **Scoped boards:** compute visible task ID set then aggregate; avoid duplicating records across joins.
- **Accepted:** distinct from completed and with defined reviewer/criteria; KPI numerator uses accepted unless explicitly indicated.
- **Token usage:** input/output/cache separately and normalized into estimated charge, measured charge, provider bill delta.

## 8. Open decisions (not blockers for documentation)
- Which tasks require mandatory human acceptance versus automatic deterministic acceptance?
- Which identity provider (local OIDC/enterprise SSO) and data retention period?
- Annual/monthly default budget amounts, monetary base currency, risk-tier definitions per tool.
- Preferred default MCP catalogue and first business workflow for proving MVP value.
