# Domain model and business rules

## 1. Bounded contexts
**Organization:** Organization, Department, Team, Position, Role, Employee, ReportingLine, TeamMembership, SkillProfile.
**Work Management:** Project, ProjectMembership, Milestone, Task, TaskParticipant, TaskDependency, Comment, Acceptance, WorkView.
**Agent Platform:** AgentProfile, AgentRuntimeConfiguration, Provider, ModelVersion, ModelAssignment, MCPCatalog, ToolDefinition, CapabilityGrant, SkillDefinition, ToolInvocation.
**Finance:** BudgetAccount, BudgetPolicy, BudgetReservation, UsageLedger, RateCardSnapshot, ChargeAttribution.
**Governance:** Principal, Permission, Policy, ApprovalRequest, AuditEvent.
**Evaluation:** KPISpec, EvaluationDataset, EvalRun, TaskEvaluation, ModelExperiment, Recommendation, DeploymentDecision.

## 2. Core definitions
- `Position`: organizational seat/job title, usually in department; skills and ceiling tool permissions.
- `Role`: reusable duty/permission taxonomy, possibly many per position; do not equate job role with security role.
- `Employee`: stable identity, `kind=HUMAN|AI`; primary position, department, manager; agent profile iff AI.
- `AgentProfile`: prompt/skill/memory configuration of AI employee; **does not equal a model**.
- `ModelAssignment`: versioned effective provider/model/params for AI employee; swap without changing employee ID.
- `Manager`: same Employee identity, not magical global privileges; leader's authority governed by access policy and delegation budget.
- `Team`: collaboration grouping, membership M:N, optional lead; distinct from department and reporting line.
- `Project`: unit of business objective, primary project owner, 0..N team/employee/department participants, milestones, budget.
- `Task`: business work item, primary project nullable, accountable owner, optional executor/participants; task does NOT imply a single LLM invocation.
- `Execution`: individual attempt/run; contains model and tool calls, pause/continue checkpoints, accounting.
- `Budget`: constraint dimensions not independent charges. One underlying cost ledger, many scoped views.

### Version 1.1 additional domain concepts
- `WorkOrigin`: task created by HUMAN, MANAGER, SCHEDULE or EVENT (GOAL deferred). Original trigger/delegation and creator preserved.
- `OrganizationChangeProposal`: AI Manager candidate hire/appointment/transfer/reorganization request, cannot apply before authorized human approval.
- `TaskProposal` / `AutomationTrigger` / `TriggerOccurrence`: bounded proactive task generation, dedup and policy enforcement.
- `MemoryItem`: scoped, versioned, verifiable long-term knowledge with ACL, classification, provenance and retention.
- `ImprovementProposal`: evidence-backed prompt, skill, model, provider or tool-policy change with human gate for impactful changes.
- `BusinessGoal`, `GoalPlanVersion`, `Initiative`: **P2 design only**, they link to canonical work/task engines.

**Confirmed business rules BR-HR, BR-TRIG, BR-COL, BR-CAP, BR-MEM, BR-OPT and BR-BUD-006..008** are normative in `INITIATIVES-TRIGGERS-AND-MEMORY.md`. Existing BRs remain valid and additive.

## 3. Rules and invariants
| Rule | Invariant and failure behavior |
|---|---|
| BR-ORG-001 | At most one active manager per employee, enforced by DB history + transaction. Founder/root can have none. |
| BR-ORG-002 | No self-manager or transitive reporting cycle; changing manager checks tree under transaction advisory lock. |
| BR-ORG-003 | Historical managers/title/department/model assignments are effective-dated and never silently overwritten. |
| BR-TEAM-001 | Members can join multiple teams/projects; team membership does not auto-grant reporting authority. |
| BR-WORK-001 | Project must have accountable owner; cross-department visibility from project membership & grants, not org-chart alone. |
| BR-WORK-002 | Task has exactly one accountable owner before READY and one primary project at most; independent tasks supported. |
| BR-WORK-003 | Task ACCEPTED only by authorized reviewer or verifiable, approved deterministic rule. Completed != Accepted. |
| BR-WORK-004 | Project/task/employee/team/company metrics use distinct task IDs; ledger facts counted once per dimension/reporting axis. |
| BR-WORK-005 | Task dependencies cannot form cycles. BLOCKED reason and unblock conditions required. |
| BR-WORK-006 | Status transitions guarded; changes to task owner/status/deadline/budget are audit events with actor and reason. |
| BR-WORK-007 | All scoped boards must enforce row-level authorization server-side; URL filters alone are not authorization. |
| BR-WORK-008 | Overdue/SLA counts derive from UTC due times and explicit calendar/timezone policy; preserve historic timestamps. |
| BR-WORK-009 | Task cancellation releases unused reservations, keeps spent ledger and all executed side effects recorded. |
| BR-AUTH-001 | No agent can grant itself a role, budget, tool permission, or approve self-originated escalation. |
| BR-AUTH-002 | Effective tool permission requires org allow, position ceiling, employee grant, task scope, runtime/tool policy, approval where applicable. |
| BR-AUTH-003 | Model tool-call support is capability, never authorization; LLM output is untrusted instruction. |
| BR-AUTH-004 | Scope delegation carries parent task ID, delegation chain, max depth, amount, deadline; forbid ancestor/self recursive loops. |
| BR-MCP-001 | MCP server is registered and approved; changes in tools/schema trigger policy review before exposure. |
| BR-MCP-002 | Tool invocations have schema-validated args, resource/operation scope, allow/deny, audit and idempotency key where applicable. |
| BR-SKILL-001 | Skills are versioned instruction bundles without inherent credentials or tool grants; skill change requires provenance and allowed source. |
| BR-MOD-001 | One effective primary model assignment per AI employee at a given time; version and rollback target retained. |
| BR-MOD-002 | Switch primary assignment requires frozen benchmark evidence + authorized human approval; emergency provider failover limited to approved pool. |
| BR-EVAL-001 | Agent self-rating may be supporting evidence but not sole acceptance/KPI signal. |
| BR-EVAL-002 | Compare cost/accepted-task with retries, tool and evaluation charges; report confidence/sample size and quality threshold. |
| BR-BUD-001 | Each planned LLM/tool call reserves estimated max charge atomically across all relevant caps before issuing. |
| BR-BUD-002 | Per-call eventual provider usage settlement, reservation release/expiry reconciliation, source usage immutable and auditable. |
| BR-BUD-003 | Hard cap breach denies/requires approval; queued/blocked jobs do not start if unavailable funds. |
| BR-BUD-004 | Company, department, team, project, employee and task limits represent intersecting constraints; charges are not added for each scope. |
| BR-BUD-005 | Explicit charge attribution avoids overlap when agent works across teams; cost ledger stores one immutable event and attribution dimensions. |
| BR-RUN-001 | At-least-once messages; idempotent state transitions; external write actions require replay protection or reconciliation. |
| BR-RUN-002 | Timeouts, step depth, tool-call counts, run budget and maximum retries prevent runaway loops. |
| BR-SEC-001 | Restricted tool actions use human approval gates and can expire; deny-by-default if policy cannot be evaluated. |
| BR-SEC-002 | No credentials stored in model prompt, DB cleartext logs, traces or skill package. |

## 4. Lifecycle state machines
### Employee
`DRAFT -> ACTIVE -> SUSPENDED -> ACTIVE` or `ACTIVE/SUSPENDED -> RETIRED`. RETIRED immutable historical identity; historical works remain accessible to authorized auditors.

### Project
`DRAFT -> PLANNED -> ACTIVE -> ON_HOLD -> ACTIVE -> COMPLETED -> ARCHIVED`; cancellations are separate terminal states. Completion requires milestone/acceptance policy.

### Task
`DRAFT -> BACKLOG -> READY -> IN_PROGRESS -> (BLOCKED | IN_REVIEW) -> COMPLETED -> ACCEPTED`; return from IN_REVIEW/COMPLETED to IN_PROGRESS with reason; `REJECTED` triggers rework; `CANCELLED` terminal. Acceptance separate from execution completion.

### Execution
`CREATED -> QUEUED -> RUNNING -> WAITING_APPROVAL/WAITING_TOOL -> RUNNING -> SUCCEEDED/FAILED/TIMED_OUT/CANCELLED`. A succeeded execution does not automatically accept a task.

### Model change
`DRAFT -> BENCHMARKING -> PROPOSED -> APPROVAL_PENDING -> APPROVED -> CANARY -> ACTIVE` or `REJECTED/ROLLED_BACK` with experiment metadata.

### Additive V1.1 state machines
- HR proposal: `DRAFT -> SUBMITTED -> REVIEWING -> APPROVED/REJECTED/EXPIRED -> APPLIED` (approval≠apply, re-check conditions).
- Task proposal: `DRAFT -> PENDING_REVIEW -> APPROVED/REJECTED -> MATERIALIZED` or `EXPIRED`; in-scope auto-created task carries origin link.
- Trigger occurrence: `RECEIVED -> DEDUPED/SKIPPED/PENDING_APPROVAL/TASK_CREATED/FAILED`; redrive remains idempotent.
- Memory item: `CANDIDATE -> VERIFIED/REJECTED -> ACTIVE -> REVOKED/EXPIRED`; curation and provenance mandatory for high-impact facts.
- Improvement proposal: `DRAFT -> EVALUATING -> PROPOSED -> PENDING_APPROVAL -> APPROVED/REJECTED -> CANARY/ACTIVE/ROLLED_BACK`.
- Goal-driven lifecycle exists only as **future design** in `01-product/TASK-AND-GOAL-ROADMAP.md`.

## 5. Ambiguous cases resolved
- AI reports to HUMAN/AI: manager FK points to employee; founder is a human employee/principal, not separate untyped string.
- Empty company: one human founder, zero AI employees allowed; supports growing 1..N.
- Team manager vs reporting manager: team lead may coordinate work but does not automatically change direct manager.
- Multiple people working on one task: accountable owner is one, collaborators many, runtime executions many.
- Shared project budgets: project cap AND employee cap both enforced on one charge; project expense is **not** additional company charge.
- Role skills baseline can be overridden by exceptions only via approved elevated entitlement, with expiry and audit.
