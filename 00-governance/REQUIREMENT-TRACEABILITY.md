# Traceability matrix — Confirmed V1.1

| Confirmed decision | FRs | Main BRs | Data/API evidence | Acceptance | Delivery |
|---|---|---|---|---|---|
| C-01..C-04 | FR-ORG-001..003 | BR-ORG-001..003 | `employees`, `reporting_lines`, org chart | AT-01, AT-02 | P0 |
| C-05 | FR-BUD-001..002 | BR-BUD-001..005 | `budget_accounts`, reservations, usage ledger | AT-07, AT-08, AT-19 | P0 |
| C-06 | FR-MOD-001..002 | BR-MOD-001..002 | Model/provider registry and policy | AT-12, AT-13 | P0/P1 |
| C-07 | FR-MOD-003 | BR-MOD-001..002, BR-EVAL-002 | benchmark and approved model assignment | AT-12, AT-13, AT-36 | P1 |
| C-08 | FR-AI-001, NFR-REL-001 | BR-RUN-001..002 | outbox/worker checkpoints | AT-11, AT-20 | P0 |
| C-09, C-15 | FR-TOOL-001..003, FR-CAP-001 | BR-CAP-001..004, BR-MCP-001..002 | grants/skills/tool invocations | AT-09, AT-10, AT-18, AT-28 | P0 |
| C-10..C-11 | FR-WORK-001..006, FR-UI-001..003 | BR-WORK-001..009 | boards/task/member/project | AT-03..06, AT-15, AT-16 | P0 |
| C-12 | FR-HR-001 | BR-HR-001..003 | HR change proposal + approval | AT-23, AT-24 | P0 basic/P1 full |
| C-13 | FR-TRIG-001..002 | BR-TRIG-001..005 | triggers/occurrences/task origin | AT-25, AT-26, AT-35 | P0 origin/P1 trigger |
| C-14 | FR-COL-001 | BR-COL-001..004 | participants, subtasks, cost attribution | AT-27 | P0 |
| C-16 | FR-MEM-001 | BR-MEM-001..005 | scoped memory, version/ACL/provenance | AT-29, AT-30 | P0 basic/P1 verification |
| C-17 | FR-BUD-003 | BR-BUD-006..008 | threshold notifications + top-up | AT-31, AT-32 | P0 |
| C-18 | FR-OPT-001 | BR-OPT-001..004 | optimization proposals + benchmark | AT-33, AT-36 | P1 |
| C-19 | FR-GOAL-001 | Goal future design (no MVP rule activated) | `TASK-AND-GOAL-ROADMAP`, future goal schema | AT-34, AT-GOAL-01..05 | MVP Task-driven; Goal P6 |

## Mandatory handoff policies
- Implementation agents **must not** mutate confirmed C-decisions to simplify code or remove server-side grants/budget checks.
- Every MVP PR touching a confirmed rule references `C-*`, `FR-*`, `BR-*`, `AT-*` in PR description and tests.
- Schema examples are not executable database migrations as-is, and OpenAPI YAML is **starter**, not an exhaustive business API.
- Data in standalone HTML `prototype/index.html` are demo only; do not present fake KPI/token figures as real.
