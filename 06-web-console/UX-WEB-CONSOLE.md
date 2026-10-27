# UX and web-console detailed specification

## 1. Product UX goals
**Primary interface:** Browser-based responsive web console (Angular). Core tasks must be completed from UI, not only CLI/API. For initial MVP desktop-first 1440px, adapts tablet, accessible keyboard/navigation. UI is a projection of server-controlled business state, not source of truth.

## 2. Information architecture
- **Executive Overview:** Company KPIs, portfolio health, active/blocked/overdue, spends vs caps, approval alerts, model quality changes, recent executions.
- **Organization / Org Chart:** hierarchical graph showing employee avatar/name/ID/title/role/department/status, model badge and primary manager edge. Pan, zoom, search, collapse/expand; click node opens employee detail drawer. Zero, one, or N AI employees allowed.
- **Work Board:** universal board, scope selector `Company | Department | Team | Project | Employee | Manager (reports subtree) | Position`; optional filters project, employee, department, team, status, priority, due, model, tag, risk; saved views.
- **Projects:** Portfolio grid; project detail with summary KPIs, milestones, budget/spend, members, dependencies/risks, linked tasks, executions and audit.
- **Workforce:** Employees table, position, manager, skills, MCP, models, spending and KPI comparisons; Employee Detail.
- **Model & Tools:** Registry and model capability matrix; MCP server catalog, tool grants by position/agent/task, skill bundles; model change proposals.
- **Approvals:** Human approval inbox with action preview and scoped approve/reject; no self-approval.
- **Audit & Operations:** timeline, execution trace drill-down, quotas, logs, errors, incident alerts, export.

## 3. Global scope switcher
**Scope is a query dimension, not a data owner.** A task linked to project P, completed by employee E, in department D/team T appears in each relevant scope but exists once and is counted once at company level. Managers see reporting subtree only when permitted; project participants see project-granted items despite another direct manager; privacy filters still apply.

### Board views
| View | Display | Interaction |
|---|---|---|
| Kanban P0 | Columns `BACKLOG`, `READY`, `IN_PROGRESS`, `BLOCKED`, `IN_REVIEW`, `COMPLETED`, `ACCEPTED` | status transition via drag/drop or action (server validates), WIP-limit warning |
| Table P0 | sortable/paged task rows, visible ID/title/project/owner/priority/status/deadline/cost/model | quick filters, details, selection, export within permission |
| Timeline/Gantt P1 | duration, milestones, dependency chains, blockers, team allocation | date shift requires revalidation |
| Calendar P1 | deadlines, schedule, on-call tasks | click event to detail |
| Executive KPI P0 | counts/spend/acceptance/lead time; per-scope comparisons and drill-down | click card opens filtered board; data-as-of timestamp |

## 4. Detail surfaces
### Employee drawer
Header: name, AI/HUMAN, employee ID, working status, job title/position, role(s), department, teams, direct manager and reportees. Tabs:
1. Overview: KPI trend, current activity, budget remaining/used, model version/provider.
2. Tasks: owned/assigned/delegated/pending-review, task board and history.
3. Capabilities: MCP servers, individual tools, effective allows/denies with inherited position ceiling, skill versions, approved exceptions.
4. Models: current model, history, tool-capability, fallback, benchmark outcomes, switch proposals.
5. Executions: tool call timeline, token, latency, exceptions, artifact previews and redactions.
6. Permissions & history: approval, transfers, audit events.
Actions: Assign Task, Edit Employee, Change Manager, Suspend, Request Model Switch; hide/disable unauthorized.

### Project detail
Header: project code, objective, primary owner, status, dates, risk, team/department membership. Metrics: accepted tasks/total, overdue, blocked, budget cap/spent/reserved, tool/runtime cost, milestones, critical path. Tabs: Board, Milestones, People, Budget, Activity, Artifacts, Approvals. From project to task, task to execution, execution to tool/employee.

### Task detail
Header: task ID/title, status, priority, deadline, project; description, acceptance criteria, accountable owner, actual executor, collaborators, manager chain; dependencies/blockers; budget and ledger breakdown; execution timeline with model/tool/skill; subtasks/comments/artifacts; review/accept/reject actions with audit.

## 5. Assign task on web
1. Click **Assign Task** from company board, project or employee profile. If launched from employee/project, preselect contextual values.
2. Enter title, description, project (nullable), accountable owner, optional actual executor/collaborators, priority, due date, success/acceptance criteria, risk class, task budget/time/steps.
3. Client previews permission, available budget/capability, potential conflict; server validates authoritatively.
4. Submit idempotently; UI receives task ID, status READY/BACKLOG, first execution when queued.
5. View live board update, prompt/tool steps, checkpoints, approval and artifacts; reviewer marks ACCEPTED when criteria fulfilled.
6. Changes to an existing task use ETag/If-Match version to prevent lost updates.

## 6. UI data contracts and errors
- `GET /organization/org-chart` -> nodes/edges, visible subset, effective date.
- `GET /boards/summary?scopeType&scopeId` -> count/status/spent/reserved/accepted/blocked/overdue, `calculatedAt`, data freshness.
- `GET /tasks?scopeType&scopeId&status&cursor` -> paginated tasks (canonical IDs).
- `GET /projects?scopeType&scopeId` -> portfolio query.
- `GET /employees/{id}` & `.../capabilities` -> detailed drawers.
- SSE `/boards/stream` delivers invalidation. UI has reconnect/poll fallback, loading/empty/error/partial-data/forbidden states.
- Show hard stop on budget denied, expired approvals, unauthorized tool, worker offline; never show falsely green "completed" just because HTTP submit succeeded.

## 6.1 V1.1 web requirements: proactive work, proposals, memory, and future goals
- **Company task board (MVP):** filter `originType=HUMAN|MANAGER|SCHEDULE|EVENT`; show original work-order/trigger on task detail and whether auto-generated task was approved.
- **Employee detail (MVP):** show scoped memory access metadata (namespaces/grant status, never restricted content); show improvement/hiring requests raised by employee/manager where authorized.
- **Budget board (MVP):** 80% warning for monetary and token caps, spent/reserved/remaining, blocked actions and `Request Top-up` form; disabled direct increase for AI users.
- **Approval inbox (MVP/P1):** categories HIRE/TRANSFER/BUDGET_TOPUP/PROMPT/SKILL/MODEL/PRIVILEGED_TOOL; compare versions, budget delta, impact, approve/reject with reason.
- **Trigger console (P1):** schedule/event sources, owner, enable toggle, frequency, cap, last runs, idempotency duplicates, dry-run and suspension on anomaly.
- **Goal portfolio (P2 — NOT MVP):** goal KPI baseline/target/actual, plans/initiative DAG, owner, budget, last review/replan, lineage to company/project/team/person board and audit. May display a disabled 'Planned' navigation only if clearly labeled.
- KPI and cost report must distinguish task source and accepted/denied/queued outcomes without double counting.

## 7. Accessibility/responsiveness
- Keyboard accessible org graph (alternative list/tree), status expressed as text not solely color.
- Tables with visible sort/filters and mobile-friendly drawer/stacked layout.
- Tool outputs sanitized, escape untrusted markup/markdown and filenames; strip executable HTML.
- Zero employees/projects/task state should guide first employee/project creation.

## 8. Acceptance criteria — user-visible
- From company dashboard, click `BLOCKED` KPI -> board filtered blocked across authorized company.
- Select Department D, then Team T, then Project P, then Employee E -> filters reconcile task IDs with authorized membership.
- Org chart click employee -> modal/drawer shows title, role, manager, model, MCP/skills, active tasks, token/cost and performance metrics.
- Project click -> project detail with progress, task list, members, budget, milestones and tool trace drilldowns.
- Assign task to employee from drawer -> task appears in company, employee, department, relevant team and project boards (no duplicate).
- Changing task status with insufficient permission is denied by backend, UI reverts optimistic transition.
- Auditors can trace KPI/card -> canonical tasks -> run -> tool calls and single authoritative ledger charges.

## 9. Prototype limitation
`prototype/index.html` simulates filters, profiles, task create/status changes using in-browser data/localStorage. It does **not** implement real auth, LLM executions, MCP servers, provider usage, SSR/sandboxing or billing, and must not be mistaken for production platform.
