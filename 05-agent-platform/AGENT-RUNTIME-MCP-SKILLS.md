# Agent runtime, MCP, tools, skills, entitlement policies

## 1. Separation of concerns
- **Position:** Job title, responsibilities, competence/skill profile and policy **ceiling** for tools. Position does not execute.
- **Agent/AI Employee:** Stable identity, manager, unique per-employee working instructions, MCP grants/skills, memory namespace and model assignment history.
- **Model/Provider:** Inference engine; flags for `tool_calling`, `parallel_tool_calling`, JSON/structured outputs, multimodality, context/token limits, parameters, pricing, health. Does not receive direct long-term MCP credentials.
- **Tool:** Named callable capability with JSON input/output schema; may originate from MCP server or internal API; has version/risk/side-effect classification.
- **Skill:** Versioned instructions/procedures/resources with prerequisites; may request tool usage but cannot independently authorize it.
- **Task/Execution:** Current operation; narrows entitlements and allows exception only after explicit approved elevation.

## 2. MCP and tool metadata (registration catalog)
### MCPServer
`id, name, base_uri/transport, trusted_publisher, signature_or_hash, auth_method, credential_ref, enabled, health, server_config_version, discovery_timestamp, allowed_network_destinations, risk_default, exposure_policy`.
### ToolDefinition
`id, mcp_server_id, qualified_name (server/tool), tool_schema_json, schema_hash, read_or_write, side_effect, risk_level, idempotent, mandatory_approval, resource_scope_template, data_classification, invoke_timeout_ms, max_response_bytes, active_version`.
### SkillDefinition
`id, name, version, content_hash, instruction_bundle_ref, owner, review_status, required_tools[], prereqs[], test_cases[], source_provenance, expiry/review_at`.

Tool names must be qualified by server identity and schema version to avoid collisions. MCP-discovered schema change invalidates cached approval until reviewed where risk warrants it. Tools are exposed to model only after capability filtering and policy intersection.

## 3. Default entitlement evaluation
**Allow only if all conditions hold:**
```
companyAllow(tool, target)
AND positionCeiling(employee.position, tool, operation)
AND employeeGrant(employee, tool, operation)
AND taskScope(task, tool, target)
AND modelSupportsToolCalling(selectedModel)
AND executionPolicy(run, risk, budget, time, rate_limit)
AND approvedIfRequired(tool, target, action)
```
Deny overrides allow. If employee has no individual rule, issue only **default grants explicitly inherited** under approved position policy; never silently elevate. A position ceiling is strict: exceptional access requires scoped, time-limited elevated policy approved by human and audited. For all decisions show breakdown of inherited grants/denies in UI.

### Example
**Backend Engineer position**: git read/write on allowed development repos, shell test commands in sandbox, code artifact upload; production deploy denied.
**DEV-007 employee**: Git read + branch write, `mvn test`, no network egress; Java/Spring and unit testing skills; runs tool-capable model.
**Task DEV-234**: only repo `abc`, branch `feature/dev-234`, max 20 git writes, no SSH or prod deployment.
**Effective**: tool granted by position but not employee grant or outside task scope -> DENY. If agent attempts `rm -rf /` -> DENY at sandbox/tool gateway even if LLM asks.

## 4. Role and agent configuration example
```yaml
position:
  key: RESEARCH_ANALYST
  title: Research Analyst
  roleIds: [research, reporting]
  ceiling:
    - {tool: web/search, operations: [read]}
    - {tool: docs/read, operations: [read]}
    - {tool: docs/create, operations: [create], approval: when_external_publish}
  requiredSkills: [source-verification, market-analysis]
employee:
  key: EMP-001
  kind: AI
  managerEmployeeId: HUMAN-FOUNDER
  modelAssignment: {provider: provider-a, model: model-x, fallbackGroup: approved-low-risk}
  mcpServers: [web-mcp-v1, docs-mcp-v1]
  toolGrants: [{tool: web/search, operations: [read]}, {tool: docs/create, operations: [create]}]
  skills: [{key: source-verification, version: '1.2'}, {key: market-analysis, version: '1.0'}]
  memory: {namespace: employee/EMP-001, projectAccess: grant-only}
  budget: {monthlyUsd: 20, monthlyTokens: 2000000, maxConcurrentRuns: 1}
task:
  key: TASK-001
  allowedToolScopes: [{tool: web/search, domainAllowlist: ['example.org']}]
  maxUsd: 2
  maxSteps: 25
  acceptance: {mode: HUMAN_REVIEW}
```
All model/provider identifiers and budget numbers are illustrative.

## 4.1 Confirmed permission & memory extensions (C-15/C-16)
- Position defines default MCP servers, tool scopes and skill catalog with a ceiling. Employee/agent grants can only narrow by default. Time-limited elevation demands human approval + independent capability-token check.
- Position skills are **templates**: an agent may activate only approved versions, and tool prerequisites must be explicitly granted to that employee/task; skill text never implies permission.
- Memory namespaces: `EMPLOYEE`, `TEAM`, `DEPARTMENT`, `PROJECT`, `ORGANIZATION`, each with read/write policy, data classification, provenance, retention. Model swaps preserve employee identity but not old permission caches.
- Inference memory retrieval must be task-scoped and re-authorized at the time of retrieval; retrieved content is untrusted data. Candidate memory is not verified truth; periodic review/expiry applies.
- For all tool calls, persist effective `position_policy_version`, `employee_grant_version`, `task_scope_version`, `skill_version`, `model_revision`, `tool_schema_hash`, approval and denial reason to support audit/replay.

## 4.2 Proactive tasks vs unlimited autonomy (C-13/C-19)
- Task-driven worker executes signed work orders; it may create subtasks within parent budget/depth/permissions and request out-of-scope task proposals; it **cannot** start an infinite business goal loop independently.
- Schedule/event triggers are registered with owner, max tasks, spend, expiry, frequency and allowed operations. The scheduler cannot bypass the business control plane.
- Goal-driven planning is future: only produces draft plan/initiative that must go through policies and Work API; no direct side-effect bypass.

## 5. Secure tool-call lifecycle
1. Worker fetches effective approved MCP tools and schema within execution context.
2. Model proposes `tool_call(name, arguments)`; never directly invokes server with credentials.
3. Tool Gateway parses/validates input JSON, length and target resources; performs independent authorization and risk checks.
4. If approval needed, persist exact sanitized action/target/change preview, checkpoint and create approval request with expiry; user can approve or reject in web.
5. On allow, issue scoped short-lived internal capability token; connector fetches secrets by reference, runs in sandbox/network allowlist.
6. Capture tool request fingerprint, result metadata, artifact refs, error, duration, charge; redact secrets/PII.
7. Emit event and settle resource usage; worker resumes. On ambiguous external effect, reconcile instead of blind retry.

## 6. Tool-call risk levels (proposal)
| Risk | Example | Default |
|---|---|---|
| R0 | Read non-sensitive approved docs | auto-allow within scope |
| R1 | Read internal repo or customer metadata | conditional allow, audit |
| R2 | Write development branch, create internal task | bounded allow or approval on novel scope |
| R3 | Delete records, send external email, modify infrastructure | human approval |
| R4 | Payments, deploy production, change credentials/policies/budget | explicit privileged approval, dual control optional |

## 7. Run-level state, tool safety and skills
- Max steps/delegation depth/retries/time/token/cost per run; model response guard for function format and hallucinated tool names.
- Separate `planner` and `executor` capabilities when model doesn't support tool calling; planner cannot bypass executor policy.
- Prompt injection defense: external web/doc/MCP responses are untrusted, never grant permissions from content; harden network egress, files, shell, SSRF targets, domain allowlists.
- Skill installations/upgrades signed or hash pinned; never execute arbitrary skill scripts except in approved sandbox.
- Memory: personal, task, project, department, org namespaces, use access checks for read/write; version and retention policy; no cross-role leakage.
- Model failover: allowed only within approved and benchmarked fallback pool; significant change to primary model requires approval.

## 8. Tool/skill API contract examples
```
GET /employees/{employeeId}/capabilities
GET /positions/{positionId}/capability-policy
POST /mcp-servers             (admin + secret ref)
POST /mcp-servers/{id}/refresh-tools  (inspect schema delta)
POST /skills                 (version, source hash, prereqs)
POST /capability-grants      (subject, target, scope, expiry, approval ref)
POST /tool-authorizations/evaluate (runId, toolQualifiedName, argsHash, target)
POST /tool-invocations       (worker service identity, signed run token, idempotency key)
GET /executions/{runId}/tool-calls
```
MCP itself covers discovering/invoking tools; position/employee/task authorization is **platform governance** layered on top of MCP.

## 9. Verification tests
- Position grants Git but employee revoked => deny.
- Employee grants Git, position does not => deny, unless human-approved scoped exception.
- Task disallows target project or repository => deny even with other allows.
- Disabled MCP server or changed schema => unexposed/blocked until reviewed.
- Model without function calling => filtered/routed to supported runtime; no fake implied capability.
- Skills that ask for extra privileges => do not change effective grants.
- Side-effecting tool retry with same idempotency key => one external effect or reconciliation case.
