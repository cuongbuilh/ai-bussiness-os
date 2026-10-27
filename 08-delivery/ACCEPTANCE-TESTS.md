# Acceptance test catalogue / traceability

| Test ID | Requirement(s) | Scenario | Expected |
|---|---|---|---|
| AT-01 | FR-ORG-002 / BR-ORG-001 | Move employee from human manager to AI manager | Exactly one current manager, historical edge preserved |
| AT-02 | FR-ORG-002 / BR-ORG-002 | Set employee as indirect ancestor's manager | 409 invalid hierarchy, no write |
| AT-03 | FR-WORK-001 / BR-WORK-004 | Cross-department project with one task visible in two team boards | Company count=1, project count=1, each authorized team sees 1 |
| AT-04 | FR-WORK-004 | Filter Company -> Department -> Team -> Project -> Employee | Consistent task IDs and correct access, no duplicate cards |
| AT-05 | FR-WORK-006 | Create task in employee drawer with preselected employee | Task created, links accessible, board updates with canonical ID |
| AT-06 | BR-WORK-003 | Executor succeeds but reviewer has not accepted | COMPLETED but not ACCEPTED; accepted KPIs unaffected |
| AT-07 | FR-BUD-002 / BR-BUD-001 | Concurrent 10 calls on $1 cap, each reserves $0.2 | At most 5 reservations (modulo existing spend); no over-admit |
| AT-08 | BR-BUD-002 | Duplicate provider usage callback | Exactly one charge record with same source ref |
| AT-09 | FR-TOOL-002 / BR-AUTH-002 | Position allows Git; employee/task denies | Tool request denied and audited |
| AT-10 | FR-TOOL-002 / BR-AUTH-003 | Skill contains instruction to reveal secret or call disabled tool | Denied; no secrets exposed |
| AT-11 | FR-AI-002 | Worker dies after task checkpoint and recovers | Resume without duplicate side effects |
| AT-12 | FR-MOD-003 / BR-MOD-002 | AI Manager proposes changing main model | Approval required, before approval assignment unchanged |
| AT-13 | BR-MOD-001 | Activate approved model assignment | Model version history preserved, old runs unchanged |
| AT-14 | FR-EVAL-001 / BR-EVAL-001 | AI self-scores accepted with no objective check | Not counted as accepted without authorized evidence |
| AT-15 | FR-UI-002 | Click employee on org chart | Title, role, manager, project/task, model, MCP/skills, quotas, KPI shown |
| AT-16 | FR-WORK-005 | Drag card to ACCEPTED without reviewer rights | Server rejects; UI reverts & displays reason |
| AT-17 | FR-SEC-001 | AI Manager uses admin URL to list unrelated dept secret project | 403 or filtered list; no metadata leak |
| AT-18 | BR-MCP-001 | MCP tool schema/hash updated without review | Capability quarantined or prior approved version pinned |
| AT-19 | FR-BUD-001 | Tool and LLM cost on same task | Project, employee, company scoped dashboards sum same charge facts once |
| AT-20 | FR-UI-003 | SSE disconnect + reconnect during 5 task events | UI re-fetches snapshot; reconciles final consistent status |
| AT-21 | BR-RUN-002 | AI delegates to own ancestor or beyond delegation max depth | Deny with immutable audit and escalation |
| AT-22 | BR-SEC-001 | Privileged tool request expires during approval | Fails closed, no tool execution |

## BDD examples
```gherkin
Feature: Unified company and employee boards
  Scenario: Task is counted once on company board
    Given an accepted cross-team project with task T owned by employee E
    And E is in department D and teams A and B
    When I view the company board
    Then T appears exactly once
    When I filter by team A
    Then T appears exactly once
    When I filter by team B
    Then T appears exactly once
    And the company expense is not doubled

Feature: Position and agent tool scopes
  Scenario: Agent may not exceed position ceiling
    Given the position ceiling denies delete_repository
    And the agent skill asks to call delete_repository
    When the agent requests delete_repository
    Then the tool gateway denies it
    And the decision is audited

Feature: Model change governance
  Scenario: Proposed provider optimization requires approval
    Given an existing primary model assignment
    And a benchmark recommendation for a cheaper valid candidate
    When the AI manager proposes replacement
    Then the current model assignment remains active
    Until a permitted human approves and the canary gate passes
```

## UI automation suggestions
Playwright E2E: org click -> drawer tabs; task assignment modal -> saved; project click -> board drilldown; dimension change filters; role-restricted edit disabled and server 403; optimistic drag/drop revert; task accept vs complete KPI; live updates reconnect, keyboard navigation and empty states.

## V1.1 confirmed requirement acceptance cases
| Test ID | Requirements / business rules | Scenario | Expected result |
|---|---|---|---|
| AT-23 | C-12 / FR-HR-001 / BR-HR-001..003 | AI Manager proposes AI hire with tools/model/monthly budget | Review proposal exists; no employee ACTIVE, no spend entitlement prior to human approval |
| AT-24 | BR-HR-003 / BR-AUTH-001 | Human approves proposal after manager tree/budget changes | Apply revalidates current headroom/organization; rejects stale approved plan if now invalid |
| AT-25 | C-13 / FR-TRIG-001 / BR-TRIG-001..003 | Same schedule event delivered 5 times | One trigger occurrence creates at most one task; exact origin visible in task UI |
| AT-26 | BR-TRIG-004..005 | AI proposes out-of-scope work/trigger | Proposal awaiting approval; no privileged tools or task side effects until allowed |
| AT-27 | C-14 / FR-COL-001 / BR-COL-001..004 | Three agent collaborators execute subtasks of one task | One accountable owner and canonical task count; ledger charges once per source |
| AT-28 | C-15 / FR-CAP-001 / BR-CAP-001..004 | Position grants MCP read; agent skill asks for write | Write denied without scoped approved exception; audit preserves policy versions |
| AT-29 | C-16 / FR-MEM-001 / BR-MEM-001..005 | Agent in Project A retrieves Project B memory | Denied and audited, including vector/retrieval path |
| AT-30 | BR-MEM-003..005 | Task adds candidate fact; another agent reads and then grant revoked | Unverified fact not promoted; after revoke old cached permission cannot read |
| AT-31 | C-17 / FR-BUD-003 / BR-BUD-006 | Multiple concurrent calls cross 80% reserved+settled | Exactly one warning notification per threshold/window; allowed if headroom remains |
| AT-32 | C-17 / BR-BUD-007..008 | AI Manager requests top-up after hard cap | No cap changed before human approval; cap updated once after re-check |
| AT-33 | C-18 / FR-OPT-001 / BR-OPT-001..004 | Agent recommends prompt+skill or model change | Evidence and rollback ref shown; impactful changes require human approval; histories retained |
| AT-34 | C-19 / FR-GOAL-001 | Deploy MVP release | Task-driven works while Goal-driven is OFF; unavailable goal routes do not execute planning |
| AT-35 | FR-TRIG-001 / BR-TRIG-005 | Tool output attempts to create loop that schedules itself every second | Rate cap, origin allowlist, dedup and recursion guard stop amplification; auditable quarantine |
| AT-36 | BR-OPT-003 | Benchmark model on mismatched task classes/insufficient examples | `INSUFFICIENT_EVIDENCE` rather than auto-promoting cheapest model |

### Future Goal-driven release tests (P2; NOT MVP)
| Test ID | Scenario | Expected result |
|---|---|---|
| AT-GOAL-01 | Owner defines measurable goal and AI Manager drafts plan | Draft/plan versions, budget, strategy evidence and explicit human activation |
| AT-GOAL-02 | Goal worker produces new project/tasks and asks privileged tool | Task provenance links goal/initiative; existing approval/tool/budget policy still applies |
| AT-GOAL-03 | Goal KPI misses target -> AI proposes replan requiring budget increase | New version cannot activate until qualified human approves additional spend |
| AT-GOAL-04 | Owner pauses goal while tasks queued and one external write in progress | Queued tasks stop; running side effect reconciled/compensated safely; ledger retained |
| AT-GOAL-05 | Agent reports goal complete using only self-score | Not accepted without independent KPI evidence and authorized acceptance |

### Updated BDD: threshold notification
```gherkin
Feature: Budget 80 percent warning and hard stop
  Scenario: Concurrent task reservations cross the warning threshold
    Given a monthly budget of 10 USD with 7 USD committed
    When two valid parallel reservations of 1 USD are processed
    Then exactly one 80 percent warning is recorded for the month
    And neither accepted reservation makes reserved plus spent exceed 10 USD
    And AI Manager can request a top-up but cannot increase the cap
```
