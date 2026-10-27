# Position, Agent and Model — policy configuration examples

## 1. Distinct concepts
- **Job position**: can have a title and 1..N business roles; defines duties, MCP/skill *defaults*, tool permission ceiling and KPI rubrics.
- **Agent employee**: identity with prompt versions, explicit MCP grants/denies, skills, manager, delegated budget, memory ACL, model assignments and concurrency.
- **Model**: a provider-backed inference engine with tool-call/structured-output support, parameters, price context, health; provides capability, not authorization.
- **Task/Run**: narrows all capabilities, holds scope/budget/approval context; no agent can call an MCP just because its instruction references that MCP.

## 2. Example position registry
| Position | Default MCP/tools (ceiling) | Default skills | High-risk boundaries |
|---|---|---|---|
| Research Analyst | Search (allowed domains), read-only documents, read project memory | source-verification, market-research, writing | No sending external data without human approval |
| Backend Developer | Git repo read/write assigned branch, isolated terminal + test runner | java-spring, code-review, secure-coding | No production deploy, credential reads or deleting repository |
| AI Project Manager | work-board read, in-scope task/subtask creation, budget status read, employee profile scoped | planning, task-estimation, risk-review | Hiring/top-up/model switch = proposal, NOT apply |
| AI Finance Analyst | scoped ledger read + reporting sandbox | cost-analysis, audit, forecasting | No payments, no direct cap edits |
| QA Analyst | read repo, isolated test runner, issue tracker | test-design, bug-triage | No prod credentials, no direct merges |

## 3. Canonical YAML example (illustrative config; not yet importable)
```yaml
position:
  code: RESEARCH_ANALYST
  title: Research Analyst
  business_roles: [RESEARCH, REPORT_WRITING]
  capability_ceiling:
    mcp_servers:
      - id: WEB_SEARCH
        tools: [search]
        conditions: {domain_allowlist: [example.org], data_classification_max: INTERNAL}
      - id: PROJECT_DOCS
        tools: [read_document]
    skills: [source_verification@1.0, market_research@2.1]
agent_employee:
  employee_code: RES-001
  manager_employee_id: HUMAN_OWNER_UUID
  enabled_mcp_servers: [WEB_SEARCH, PROJECT_DOCS]
  allowed_tools: [WEB_SEARCH/search, PROJECT_DOCS/read_document]
  denied_tools: [PROJECT_DOCS/delete_document]
  activated_skills: [source_verification@1.0]
  memory_scopes: [EMPLOYEE:RES-001, PROJECT:PRJ-001]
  monthly_budget: {hard_usd: 20.00, hard_tokens: 2000000, warning_percent: 80}
  max_parallel_runs: 1
model_assignment:
  provider_alias: provider_a
  model_alias: model_x
  required_capabilities: {tool_calling: true, structured_output: true}
  primary_version: rev-pinned
  change_policy: HUMAN_APPROVAL
```

## 4. Permission resolution
```
authorize(tool_call) iff
  organization_allows(tool, data, network)
  AND position_ceiling_allows(tool, operation, target)
  AND agent_grant_allows(tool, operation, target)
  AND task_scope_allows(tool, operation, target)
  AND risk_approval_valid_if_required()
  AND current_budget_reservation_valid()
  AND selected_model_can_emit_supported_call()
```
A denied tool MUST stay denied even when invoked by an approved skill. Model capability does not elevate security authority. Human-approved exception has explicit target, operation, scope, expiry and audit reference.

## 5. Runtime execution trace checklist
For each run: employee ID, position/policy revision, task, project, direct manager snapshot, model/provider/revision, prompt and skill hash, MCP schema versions, tool call request + approval result, retry, token categories, estimated/reserved/settled cost, artifacts, review decision, measurable KPI.

## 6. Example of proposal vs authority
AI Project Manager may issue: `POST /hr/change-proposals`, `POST /task-proposals`, `POST /budgets/{id}/top-up-requests`, or `POST /optimization/proposals`, with supporting evidence. These operations never imply direct authority to create ACTIVE employees, raise caps or activate a new primary model. A distinct authorized human decision and apply/revalidation step is required.
