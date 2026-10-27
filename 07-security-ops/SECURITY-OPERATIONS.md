# Security, approval, operations, deployment

## 1. Trust boundaries and threat model
- Human browser is untrusted input; can request tasks or approvals only through authenticated API.
- Agent response and tool results are untrusted content even if the model is high-performing. Never promote an instruction from user-provided docs/web/skill to platform policy.
- Tool Gateway hosts privileged connector access; API key/credential references remain server-side.
- Worker is not trusted to self-authorize a privileged action or create/raise its own budget.
- MCP servers may change tool schemas and behaviors; hash-pin/approve and isolate network egress.

## 2. Identity, policy, principals
- Human identities via chosen OIDC/enterprise IdP; service accounts for worker, MCP Tool Gateway, gateway; rotate tokens.
- AI employee `employee_id` is a business identity, not a human login; it acts with signed, short-lived delegation context for task/tool scope.
- Authorization uses RBAC for UI functions + ABAC/resource scope for department/team/project/subtree/task/operation/data classification.
- Sensitive approvals require human principal and separation of duties. Policy evaluation server-side on all writes and sensitive reads.
- Replay protection / signed task context; CSRF for cookie sessions; scoped CORS; origin/network restrictions.

## 3. Protected tools and data
- Per-tool risk classification R0..R4, operations read/write/delete/execute, resource scoping and human gate.
- Commands executed in ephemeral non-root sandbox, resource caps, network allowlists, deny host Docker socket and prod K8s credentials.
- Prompt injection protections, schema validation, anti-SSRF/private IP blocking for browser/network tools, malicious files handled as data.
- Secrets in Vault/secret manager or encrypted credential storage, never Git/SQL cleartext; no raw secrets in prompts or analytics.
- Role/project/employee-scoped memory, retention + deletion redaction process, prevent cross-team visibility leaks.

### V1.1 abuse cases & controls
- **Self-hire / budget escalation:** a manager's proposal cannot create active employment, add grants or raise caps until human approval; revalidate on apply, prohibit requester as approver and circular approval chains.
- **Trigger amplification:** enforce occurrence dedup, task rate, spend, expiry, allowlisted events, origin signatures, loop prevention (self-generated events cannot endlessly retrigger same policy).
- **Skill or memory injection:** skill package and memory are untrusted instructions; no secret retrieval from memory without policy; active skills content-hash pinned and reviewed.
- **Memory leakage:** ACL at retrieval, writes and embeddings; scope changes revoke future access and invalidate caches. Prevent cross-department or cross-project data reads by tool browsing.
- **Eval gaming:** AI cannot erase fail traces, change rubric/thresholds, approve own improvements; integrity of objective evidence and versioned evaluation snapshots.
- **Approval replay / TOCTOU:** approved action binds exact target, diffs/args, time window, scope and amount; re-check grants, manager tree, budget and policy when executing.

## 4. Audit contract
Append-only events for: login, person/agent creation, manager changes, grants, model assignment, budget changes, task status/assignment, reviews, approvals, tool decisions, provider calls, charges, export, suspend. Required fields actor (human/service/agent), target, old/new hash or sanitized diff, reason, correlation ID, timestamp, source IP/device where applicable, approval reference. Audit must withstand agent-controlled edits; export to protected storage.

## 5. SLO and alerting (proposed targets, to validate)
- Budget overspend detected/blocked in app-controlled transactions; provider-late overrun goes to alert/reconciliation.
- Tool policy unknown/unavailable -> fail closed, not bypass.
- Execution broker lag threshold, worker heartbeat lost, stale SSE, orphaned reservations, pending approvals, model error rate and cost spike alerts.
- Correlation: org, project, task, run, agent, tool, model, ledger IDs.
- Backups: PostgreSQL PITR, MinIO bucket versioning, secret backups; documented restore exercise quarterly (proposal).

## 6. Docker Compose deployment shape
```
edge-nginx
  ├── angular-web
  └── spring-control-api
spring-control-api -> postgres, queue, object-store, secret-store
python-agent-worker -> queue, control-api, litellm, tool-gateway
litellm -> provider APIs/Ollama + separate persisted database/config
mcp/tool-gateway -> approved connectors/sandbox (restricted egress)
telemetry -> OTEL collector/Langfuse (optional early)
```
Production ready Compose requires HTTPS, auth, non-root, persistent volumes, backups, resource limit and bind control. Do not expose DB, Redis or service admin routes to public Internet.

## 7. Kubernetes migration contract
- Stateless services configurable by env/secrets/ConfigMaps; agent executions checkpoint to durable DB/object store, not local disk.
- NetworkPolicy isolates worker/tool egress; namespace and service account RBAC least privilege; scale workers by queued task count/latency; separate rate limits by model/provider.
- Database migrations backward compatible (`expand -> deploy -> contract`), pin versions and rollback plan.
- Load testing on canonical company board, cross-department filters, concurrent reservations, and tool policy fail-closed.

## 8. Acceptance security gates
- Unit + integration + policy property tests for scope, model capabilities, role grant intersections.
- Concurrent budget stress and key leak scan.
- Adversarial prompt-injection attempts in MCP/browser/document results.
- External write tool crash/retry scenarios tested with idem/reconcile.
- Security review before enabling production-impacting tool categories.
