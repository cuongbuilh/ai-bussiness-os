# Budget, allocation, KPI and model optimization

## A. Financial policy
**Dimensions**: USD money cap, input/output/cache tokens, TPM/RPM, maximum concurrent runs, runtime minutes, tool calls, task counts, daily/weekly/monthly windows.
**Scope constraints**: organization AND primary accountable department/team(s) as policy, project, employee, task, execution. One expense is counted once in one charge ledger, but reduces available headroom for every matching cap.
**Budget inheritance**: department/team allocation must be <= unallocated parent allocatable allowance. Cross-cut employee/project limits are parallel spend constraints, not additional company allocations. Define account hierarchy explicitly to prevent double allocation.

### Reservation/settlement state
`REQUESTED -> RESERVED -> PARTIALLY_SETTLED/SETTLED` or `EXPIRED/RELEASED/FAILED`. Use database transaction with lock ordering, unique `(execution_id, step_key)` and dedup provider charge reference. Requested estimated max with output-token cap. At settlement, reconcile charged usage to original reservation and ledger; record unmatched/provider-late charge exceptions for manual reconciliation.

```text
remaining(account) = hard_cap - settled_spend - active_reserved
allow(step) iff min_remaining across all relevant accounts >= estimated_max_cost
             AND token, request, time, task constraints pass
```
The policy must consider denied requests, escaped side effects, provider billing delays. Reserve based on maximum plausible cost for a capped request, not expected average.

### Budget exception
Soft limit triggers 80%/90% warnings (thresholds configurable); hard limit queues task as `BLOCKED` with code `BUDGET_EXHAUSTED`, requests manager approval for top-up; no silent fallback to expensive provider. `max_tokens` enforced on input estimation + max output cap + reconciliation.

### Currency
Persist provider source currency and immutable USD rate-card snapshot. If USD-to-VND conversion desired, use a dated FX rate snapshot and never retrospectively rewrite ledger amounts. Monetary tolerance/rounding methodology versioned.

### Confirmed 80% threshold / top-up governance (C-17)
- A scope/window with positive hard cap fires 80% warning on transition into `settled + active_reserved >= 0.8 * limit`, **once per threshold and account/window**. A second optional 90% warning remains *proposed*, configurable off.
- Crossing 80% alone does **not** block authorized new work; hard cap blocks admissions (and requests cannot bypass it). For tokens/time/task counters apply corresponding thresholds and their own reserve/consumption semantics.
- AI Manager may `REQUEST_TOP_UP` with reason, expected cost and originating task/project, but has **no** direct change to effective hard cap. Human with budget authority must approve and the apply stage re-checks parent headroom and concurrent allocations.
- Emitting a warning event / notifying manager/owner must be idempotent and robust to retry; budget snapshots show `spent`, `reserved`, `remaining`, `threshold_crossed`, `as_of`.

## B. KPI contracts
Each job role defines rubric (`KPISpec`): metric name, applicability, weight, unit, data source, evaluator version, minimum sample size, threshold, failure consequence.
- **Research:** verified source ratio, unsupported claims, report acceptance, citation correction rate, cost per accepted report.
- **Engineering:** reproducible test pass, regression rate, security findings, review rework, cost/accepted ticket.
- **QA:** true-positive defects, critical defect escape, false-positive, test coverage relevance.
- **Manager:** accepted throughput, cost/runaway reduction, under/over delegation, SLA, employee quality trend.
- **Finance:** reconciliation correctness, forecast deviation, audit trail completeness.

Exclude metrics susceptible to easy gaming (raw token count, number of task messages, agent self-scored confidence as sole evidence).

### Basic formulas
```
AcceptanceRate = accepted_tasks / reviewed_tasks
CostPerAcceptedTask = (LLM + tools + sandbox compute + evaluation + retry costs) / accepted_tasks
SLAHitRate = accepted_before_deadline / eligible_accepted_tasks
ReworkRate = revisions_required / reviewed_tasks
```
Always display time window, population definition, confidence/sample size, missing data and who reviewed. Avoid comparing different workload complexity classes without stratification.

## C. Model Evaluation Engine
1. **Observe:** trace run inputs, output, costs, latency, skills/tools/prompt version, evaluator, task category/complexity (masked where private).
2. **Discover:** candidate provider/model models in registry with tool/structured/context requirements + privacy policy + ceiling.
3. **Benchmark:** run same frozen evaluation dataset (golden tasks + adversarial tool/use cases); blind evaluation with deterministic rubrics and human calibration.
4. **Normalize:** compare within matching task class/risk, accounting for retries/tool costs; report quality threshold compliance, confidence interval and variance.
5. **Recommend:** generate proposal with delta quality, cost, latency, safety, reliability, rollback conditions. No auto-change of primary assignment.
6. **Approve & Canary:** human validates and selects limited traffic for low-risk tasks only. Telemetry watch SLO, quality, spending.
7. **Activate or Rollback:** versioned effective model assignments; rollback on threshold breach, retain all run provenance.

### Recommendation gating (proposal)
Quality threshold is a constraint; **do not trade away required quality for price**. Among candidates satisfying safety/tool/quality requirements, choose by weighted normalized cost and latency plus uncertainty penalty. Sample size and low-confidence estimates should produce `INSUFFICIENT_EVIDENCE`, not false precision.

### Data leakage defense
Benchmark dataset contains approved or sanitized work items. Provider residency/data retention and external processing agreement are policy filters ahead of quality/cost comparison. Human rating/evaluators must not be fed secrets.

### Confirmed optimization proposal policy (C-18)
- Proposal types: `MODEL`, `PROVIDER`, `PROMPT`, `SKILL`, `TOOL_POLICY`. Each has stable baseline, experiment/test dataset version, candidate revision, risk class, prediction, rollback ref and actual approve/apply state.
- AI may suggest changes and run **authorized evaluation** within budget; cannot unilaterally activate impactful changes (including primary model, shared skill/system prompt or tool policy).
- A KPI rise with inadequate sample size, untrusted evaluator or changing data mix produces `INSUFFICIENT_EVIDENCE`. Never compare unequal task-risk/workload classes as a definitive model ranking.
- Improvement that affects tool permissions requires capability review separately from model benchmark, even if recommended by the same AI Manager.

## D. Proposed model experiment artifacts
`experiment_id, task_class, dataset_version, evaluator_version, prompt_version, tools_version, candidate_model_revision, baseline_model_revision, sample_count, accepted_count, cost_usd, mean/p95 latency, severity-weighted failures, confidence bounds, recommendation, reviewer, approvals`.

## E. Existing tools vs platform ownership
- LiteLLM: provider abstraction, virtual keys, rate limits, model usage statistics, certain MCP permissions (requires PostgreSQL for enforced budgets). Platform remains authoritative for org/project/employee/task ledgers and reservations. https://docs.litellm.ai/docs/proxy/users
- Langfuse: traces, datasets, experiments, evaluation scores; platform translates output into business KPI policy. https://langfuse.com/docs/evaluation/experiments/datasets
- LangGraph: durable checkpoint / interrupt human approval. https://docs.langchain.com/oss/python/langgraph/interrupts
