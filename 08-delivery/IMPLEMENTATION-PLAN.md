# Implementation roadmap, epics, agent coding handoff

## 1. Delivery strategy
Begin with **vertical slice** that proves one human founder, one AI employee, one project/task, one tool-call, one budget reservation, one review/accepted outcome, one web board and trace. Expand role/org management, team collaboration and model eval only after the slice is reliable.

## Version 1.1 phased-mode boundary
- **MVP/P0 Task-driven:** actual task assignment and delegation, employee position + tool/skill ceiling, budget 80%/hard, org/board drilldown, memory namespaces, approval/human acceptance, model assignment policy; task records carry provenance.
- **P1 Task-driven enhancements:** trigger registration/UI/scheduler connectors, HR/top-up/optimization proposal workflows, memory verification/retention automation, benchmark comparison and model change recommendations. Each may be feature-flagged independently.
- **Goal-driven after MVP (delivery Phase P6):** Goal Registry, KPI baselines, Plan/Initiative Builder, planned/replan reviewer, goal budget and provenance -> same canonical Project/Task/Run pipeline. **No claim of Goal-driven in MVP.**

## 2. Phases and deliverables
| Phase | Scope | Main acceptance |
|---|---|---|
| P0. Foundation | repositories, CI, OIDC, Postgres migrations, event outbox, audit | log in / authorized action / audit trail |
| P1. Single Employee | position/employee/manager, model/provider registry, skills/MCP catalog, policy enforcement | AI employee executes authorized task; unauthorized tool denied |
| P2. Project & Board | projects/team/department, task lifecycle, unified scope board, employee/project drawers | from company to project/team/employee and assign task; no duplicates |
| P3. Budget & Execution | reservations/settle, queue, checkpoint, approval, model/tool traces, SSE | concurrent reserve safe; resume approval; run trace |
| P4. Evaluation & Proposals | KPI rubric, benchmark datasets, experiments, hiring/top-up/prompt/skill/model proposals, approval/canary/rollback | evidence-based changes; human approval and version history retained |
| P5. Hardening | security/prompt-injection, load, disaster recovery, observability, K8s-compatible packaging | SLO & risk gates pass |
| P6. Goal-driven (post-MVP) | goals, KPI targets, plans, initiatives, controlled re-planning | goal-to-task trace, bounded cost, human approval, accepted KPI evidence |

## 3. Suggested development ownership
- **Product/Domain Agent:** Own PRD/BRs, traceability, acceptance definitions; cannot silently change C-decisions.
- **Backend Agent (Java):** Spring modules org/task/project/finance/policy, migrations, REST API; owns ledger consistency.
- **AI Runtime Agent (Python):** Executor graphs/checkpoints, tool-call adapter, event processing; uses public control contracts.
- **Frontend Agent (Angular):** Org chart, board, detail panels, assignment, approvals, KPIs; mocks contract before backend integration.
- **Security Agent:** Threat model, tool scope/authz/approval tests, secrets, logs and runtime isolation.
- **QA/Integration Agent:** Cross-module API, parity between board scopes, budget race and reconnect tests.
- **DevOps Agent:** Compose environments, test CI, load/telemetry, upgrade migration plan.

## 4. Parallelization sequence
1. Product/domain freezes shared entity terms, task states, IDs, auth claims and API/events envelope.
2. UI and backend can proceed in parallel against OpenAPI mocks; runtime works against reserve and tool authorization contracts.
3. Backend finance and policy tests pass before enabling real external providers/MCP tools.
4. Integration agent assembles vertical slice and confirms one canonical task across all boards.
5. Runtime worker tool side effects and approval gates are tested before allowing automation/schedule.

## 5. Repository structure (suggested)
```
ai-business-os/
  apps/web-console/             # Angular
  services/control-plane/       # Java Spring Boot modular monolith
  services/agent-runtime/       # Python
  services/tool-gateway/        # Python or Java secured connector gateway
  contracts/openapi/            # versioned REST contracts
  contracts/events/             # event JSON Schemas
  infra/compose/                # dev/staging
  infra/k8s/                    # future
  packages/design-system/       # UI components/tokens
  docs/                         # this spec
  test/contract/                # Pact/OpenAPI validators
  test/e2e/                     # Playwright UI
  test/chaos/                   # retry / budget concurrency tests
```

## 6. Definition of Done for each epic
- Business rule IDs mapped to code + tests, security/permission tests, audit events, metric tags.
- Docs and API/DB migration backward-compatibility reviewed.
- Positive, negative and concurrency/retry scenarios passing in CI.
- Observability: task correlation end-to-end, no sensitive payload leaks.
- UI (if relevant) supports loading, empty, error, forbidden, stale data, and keyboard.
- No mocks shipped as real model quality/spend metrics.

## 7. Release gates for first functional MVP
- Founder creates position and AI employee via UI, chooses manager, provider model and skills.
- Can assign task on web, worker handles at least one approved read-only MCP tool; all actions authorized.
- Task appears in company, project, department, team and employee board scopes where appropriate.
- Employee detail displays role/title, direct manager, active tasks, model, MCP/skills, token/cost and real run trace.
- Budget hard caps block new model calls and support reconciliation for failure.
- Human can approve tool call and accept task, after which accepted KPI and ledger update.
- Model change from primary assignment requires approved request; benchmark phase can be P4 but no direct unapproved switch.

## 8. Suggested feature flags
`ff_goal_driven=false`, `ff_trigger_automation=false`, `ff_ai_hiring_apply=false`, `ff_improvement_auto_activate=false`, `ff_external_tools_write=false`, `ff_schedule_automation=false`, `ff_multi_model_canary=false`, `ff_auto_accept_deterministic=false`, `ff_project_timeline=false`, `ff_hiring_suggestions=false` until their corresponding tests pass.
