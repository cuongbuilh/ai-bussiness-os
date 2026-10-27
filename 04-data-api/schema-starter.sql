-- AI Business OS schema starter (PostgreSQL 15+ example)
-- NOT full production migration. Add auth, RLS/policies, audit, migration management.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TYPE employee_kind AS ENUM ('HUMAN','AI');
CREATE TYPE employee_state AS ENUM ('DRAFT','ACTIVE','SUSPENDED','RETIRED');
CREATE TYPE task_state AS ENUM ('DRAFT','BACKLOG','READY','IN_PROGRESS','BLOCKED','IN_REVIEW','COMPLETED','ACCEPTED','REJECTED','CANCELLED');
CREATE TYPE execution_state AS ENUM ('CREATED','QUEUED','RUNNING','WAITING_APPROVAL','WAITING_TOOL','SUCCEEDED','FAILED','TIMED_OUT','CANCELLED');

CREATE TABLE organizations (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE departments (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 name text NOT NULL, parent_department_id uuid REFERENCES departments(id),
 UNIQUE (organization_id, name)
);
CREATE TABLE teams (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 name text NOT NULL, lead_employee_id uuid NULL,
 UNIQUE (organization_id, name)
);
CREATE TABLE positions (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 department_id uuid REFERENCES departments(id), title text NOT NULL,
 role_code text NOT NULL, description text NOT NULL DEFAULT ''
);
CREATE TABLE employees (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 department_id uuid REFERENCES departments(id), position_id uuid REFERENCES positions(id),
 display_name text NOT NULL, kind employee_kind NOT NULL, state employee_state NOT NULL DEFAULT 'DRAFT',
 version bigint NOT NULL DEFAULT 0, created_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE teams ADD CONSTRAINT fk_team_lead FOREIGN KEY (lead_employee_id) REFERENCES employees(id);
CREATE TABLE reporting_lines (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 subordinate_id uuid NOT NULL REFERENCES employees(id), manager_id uuid NOT NULL REFERENCES employees(id),
 valid_from timestamptz NOT NULL, valid_to timestamptz,
 CHECK (subordinate_id <> manager_id), CHECK (valid_to IS NULL OR valid_to > valid_from)
);
CREATE UNIQUE INDEX uq_active_direct_manager ON reporting_lines(subordinate_id) WHERE valid_to IS NULL;
-- Also enforce effective interval non-overlap and transitive no-cycle in transactional domain service.
CREATE TABLE team_memberships (
 team_id uuid NOT NULL REFERENCES teams(id), employee_id uuid NOT NULL REFERENCES employees(id),
 PRIMARY KEY (team_id, employee_id)
);
CREATE TABLE agent_profiles (
 employee_id uuid PRIMARY KEY REFERENCES employees(id), prompt_version text NOT NULL,
 memory_scope text NOT NULL DEFAULT 'EMPLOYEE', max_delegation_depth integer NOT NULL DEFAULT 0
 CHECK (max_delegation_depth >= 0)
);
CREATE TABLE model_versions (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), provider_code text NOT NULL, model_id text NOT NULL,
 revision text NOT NULL, supports_tools boolean NOT NULL DEFAULT false,
 supports_structured_output boolean NOT NULL DEFAULT false,
 context_window integer, price_currency char(3) NOT NULL DEFAULT 'USD',
 UNIQUE (provider_code, model_id, revision)
);
CREATE TABLE model_assignments (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), employee_id uuid NOT NULL REFERENCES employees(id),
 model_version_id uuid NOT NULL REFERENCES model_versions(id), valid_from timestamptz NOT NULL,
 valid_to timestamptz, approval_id uuid, CHECK (valid_to IS NULL OR valid_to > valid_from)
);
CREATE UNIQUE INDEX uq_current_primary_model ON model_assignments(employee_id) WHERE valid_to IS NULL;
CREATE TABLE projects (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 code text NOT NULL, name text NOT NULL, description text NOT NULL DEFAULT '',
 accountable_owner_id uuid NOT NULL REFERENCES employees(id), status text NOT NULL DEFAULT 'DRAFT',
 due_at timestamptz, created_at timestamptz NOT NULL DEFAULT now(), version bigint NOT NULL DEFAULT 0,
 UNIQUE (organization_id, code)
);
CREATE TABLE project_memberships (
 project_id uuid NOT NULL REFERENCES projects(id), employee_id uuid NOT NULL REFERENCES employees(id),
 membership_role text NOT NULL DEFAULT 'CONTRIBUTOR', PRIMARY KEY(project_id,employee_id)
);
CREATE TABLE tasks (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 project_id uuid REFERENCES projects(id), parent_task_id uuid REFERENCES tasks(id),
 title text NOT NULL, description text NOT NULL DEFAULT '',
 accountable_owner_id uuid REFERENCES employees(id), status task_state NOT NULL DEFAULT 'DRAFT',
 priority text NOT NULL DEFAULT 'MEDIUM', acceptance_criteria text NOT NULL DEFAULT '',
 due_at timestamptz, accepted_at timestamptz, created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(), version bigint NOT NULL DEFAULT 0,
 CHECK (parent_task_id IS NULL OR id <> parent_task_id),
 CHECK (status NOT IN ('READY','IN_PROGRESS','IN_REVIEW','COMPLETED','ACCEPTED') OR accountable_owner_id IS NOT NULL)
);
CREATE TABLE task_participants (
 task_id uuid NOT NULL REFERENCES tasks(id), employee_id uuid NOT NULL REFERENCES employees(id),
 relation text NOT NULL, PRIMARY KEY(task_id,employee_id,relation)
);
CREATE TABLE task_dependencies (
 blocked_task_id uuid NOT NULL REFERENCES tasks(id), prerequisite_task_id uuid NOT NULL REFERENCES tasks(id),
 dependency_type text NOT NULL DEFAULT 'BLOCKS', PRIMARY KEY(blocked_task_id, prerequisite_task_id),
 CHECK (blocked_task_id <> prerequisite_task_id)
);
CREATE TABLE executions (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 task_id uuid NOT NULL REFERENCES tasks(id), executing_employee_id uuid NOT NULL REFERENCES employees(id),
 model_assignment_id uuid REFERENCES model_assignments(id),
 status execution_state NOT NULL DEFAULT 'CREATED', attempt_no int NOT NULL DEFAULT 1,
 created_at timestamptz NOT NULL DEFAULT now(), started_at timestamptz, ended_at timestamptz,
 idempotency_key text NOT NULL, UNIQUE(task_id,idempotency_key)
);
CREATE TABLE budget_accounts (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 scope_type text NOT NULL, scope_id uuid NOT NULL, window_start timestamptz NOT NULL,
 window_end timestamptz NOT NULL, max_usd numeric(20,8), max_tokens bigint,
 max_tasks bigint, max_runtime_seconds bigint,
 CHECK (window_end > window_start),
 CHECK (max_usd IS NULL OR max_usd >= 0), CHECK (max_tokens IS NULL OR max_tokens >= 0)
);
CREATE TABLE budget_reservations (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 execution_id uuid NOT NULL REFERENCES executions(id), step_key text NOT NULL,
 reserved_usd numeric(20,8) NOT NULL DEFAULT 0, reserved_tokens bigint NOT NULL DEFAULT 0,
 status text NOT NULL DEFAULT 'ACTIVE', expires_at timestamptz NOT NULL,
 UNIQUE(execution_id,step_key)
);
CREATE TABLE usage_ledger (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 execution_id uuid NOT NULL REFERENCES executions(id), task_id uuid NOT NULL REFERENCES tasks(id),
 project_id uuid REFERENCES projects(id), employee_id uuid NOT NULL REFERENCES employees(id),
 model_version_id uuid REFERENCES model_versions(id),
 charge_reference text NOT NULL, entry_type text NOT NULL, posted_at timestamptz NOT NULL DEFAULT now(),
 input_tokens bigint NOT NULL DEFAULT 0, output_tokens bigint NOT NULL DEFAULT 0,
 cache_read_tokens bigint NOT NULL DEFAULT 0, cost_usd numeric(20,8) NOT NULL,
 UNIQUE(execution_id,charge_reference,entry_type)
);
CREATE TABLE approval_requests (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 task_id uuid REFERENCES tasks(id), requested_by_employee_id uuid REFERENCES employees(id),
 action_kind text NOT NULL, payload_json jsonb NOT NULL, status text NOT NULL DEFAULT 'PENDING',
 expires_at timestamptz, resolved_by_employee_id uuid REFERENCES employees(id), resolved_at timestamptz,
 CHECK (requested_by_employee_id IS DISTINCT FROM resolved_by_employee_id)
);
CREATE TABLE outbox_events (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 event_type text NOT NULL, aggregate_id uuid NOT NULL, payload_json jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(), sent_at timestamptz
);
CREATE TABLE inbox_processed (
 event_id uuid NOT NULL, consumer_name text NOT NULL, processed_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(event_id,consumer_name)
);
CREATE INDEX ix_tasks_company_status ON tasks(organization_id,status,due_at);
CREATE INDEX ix_tasks_project ON tasks(project_id,status,updated_at DESC);
CREATE INDEX ix_tasks_owner ON tasks(accountable_owner_id,status,updated_at DESC);
CREATE INDEX ix_exec_task ON executions(task_id,created_at DESC);
CREATE INDEX ix_ledger_org_time ON usage_ledger(organization_id,posted_at DESC);
CREATE INDEX ix_ledger_task_time ON usage_ledger(task_id,posted_at DESC);
-- TODO: org FK consistency: composite (organization_id,id) constraints or triggers.
-- TODO: enforce direct-manager effective interval exclusion constraint; task dependency/reporting cycles.
-- TODO: enforce AI profile only for kind='AI', owner/project org consistency, policy/RLS.
-- TODO: atomic aggregate budget reservation across all scoped budget_accounts.

-- ===== V1.1 starter extension (illustrative, not a migration) =====
-- Approved in domain baseline, but P1 features can be behind feature flags.
ALTER TABLE tasks ADD COLUMN origin_type text NOT NULL DEFAULT 'HUMAN'
  CHECK (origin_type IN ('HUMAN','MANAGER','SCHEDULE','EVENT','GOAL'));
ALTER TABLE tasks ADD COLUMN origin_ref text;
ALTER TABLE tasks ADD COLUMN initiated_by_employee_id uuid REFERENCES employees(id);
ALTER TABLE tasks ADD COLUMN trigger_occurrence_id uuid;
-- GOAL is reserved for future use; MUST remain disabled until Goal-driven release.

CREATE TABLE organization_change_proposals (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 proposer_employee_id uuid NOT NULL REFERENCES employees(id),
 change_type text NOT NULL CHECK (change_type IN ('HIRE_AI','ASSIGN_POSITION','CHANGE_MANAGER','REORGANIZE')),
 status text NOT NULL DEFAULT 'DRAFT', reason text NOT NULL,
 sanitized_payload jsonb NOT NULL, estimated_monthly_usd numeric(20,8),
 approval_id uuid REFERENCES approval_requests(id),
 applied_at timestamptz, created_at timestamptz NOT NULL DEFAULT now(), version bigint NOT NULL DEFAULT 0,
 CHECK (status IN ('DRAFT','SUBMITTED','REVIEWING','APPROVED','REJECTED','EXPIRED','APPLIED'))
);
CREATE TABLE automation_triggers (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 owner_employee_id uuid NOT NULL REFERENCES employees(id), name text NOT NULL,
 trigger_kind text NOT NULL CHECK (trigger_kind IN ('SCHEDULE','EVENT')),
 trigger_mode text NOT NULL CHECK (trigger_mode IN ('PROPOSE_ONLY','AUTO_CREATE_IN_POLICY')),
 cron_expr text, timezone text, event_type_allowlist jsonb NOT NULL DEFAULT '[]'::jsonb,
 max_tasks_per_window integer NOT NULL CHECK (max_tasks_per_window > 0),
 max_usd_per_window numeric(20,8) NOT NULL CHECK (max_usd_per_window >= 0),
 expiry_at timestamptz NOT NULL, enabled boolean NOT NULL DEFAULT false,
 definition jsonb NOT NULL, version bigint NOT NULL DEFAULT 0,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE trigger_occurrences (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), trigger_id uuid NOT NULL REFERENCES automation_triggers(id),
 occurrence_key text NOT NULL, event_hash text, status text NOT NULL DEFAULT 'RECEIVED',
 linked_task_id uuid REFERENCES tasks(id),
 error_code text, received_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(trigger_id,occurrence_key)
);
ALTER TABLE tasks ADD CONSTRAINT fk_task_trigger_occurrence
 FOREIGN KEY (trigger_occurrence_id) REFERENCES trigger_occurrences(id);
CREATE TABLE task_proposals (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 proposed_by_employee_id uuid NOT NULL REFERENCES employees(id),
 parent_task_id uuid REFERENCES tasks(id), materialized_task_id uuid REFERENCES tasks(id),
 status text NOT NULL DEFAULT 'DRAFT', proposed_data jsonb NOT NULL,
 reason text NOT NULL, approval_id uuid REFERENCES approval_requests(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 CHECK (status IN ('DRAFT','PENDING_REVIEW','APPROVED','REJECTED','MATERIALIZED','EXPIRED'))
);
CREATE TABLE memory_items (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 scope_type text NOT NULL CHECK (scope_type IN ('EMPLOYEE','TEAM','DEPARTMENT','PROJECT','ORGANIZATION')),
 scope_id uuid NOT NULL, created_by_employee_id uuid NOT NULL REFERENCES employees(id),
 classification text NOT NULL, content_ref text NOT NULL, provenance_json jsonb NOT NULL,
 content_hash text NOT NULL, status text NOT NULL DEFAULT 'CANDIDATE',
 reviewed_by_employee_id uuid REFERENCES employees(id),
 expires_at timestamptz, revoked_at timestamptz, revision bigint NOT NULL DEFAULT 0,
 created_at timestamptz NOT NULL DEFAULT now(),
 CHECK (status IN ('CANDIDATE','VERIFIED','REJECTED','ACTIVE','REVOKED','EXPIRED'))
);
CREATE INDEX ix_memory_scope ON memory_items(organization_id,scope_type,scope_id,status);
-- ACL/ABAC must be evaluated in query/tool-gateway and vector index, not inferred from scope_type.

CREATE TABLE budget_topup_requests (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), budget_account_id uuid NOT NULL REFERENCES budget_accounts(id),
 requested_by_employee_id uuid NOT NULL REFERENCES employees(id),
 increase_usd numeric(20,8) NOT NULL CHECK (increase_usd > 0),
 reason text NOT NULL, status text NOT NULL DEFAULT 'PENDING',
 approval_id uuid REFERENCES approval_requests(id), applied_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(), version bigint NOT NULL DEFAULT 0
);
CREATE TABLE budget_threshold_notifications (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), budget_account_id uuid NOT NULL REFERENCES budget_accounts(id),
 threshold_percent smallint NOT NULL CHECK (threshold_percent BETWEEN 1 AND 100),
 metric text NOT NULL CHECK (metric IN ('USD','TOKENS','TASKS','RUNTIME_SECONDS')),
 crossed_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(budget_account_id,threshold_percent,metric)
);
CREATE TABLE optimization_proposals (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 employee_id uuid NOT NULL REFERENCES employees(id), proposer_employee_id uuid NOT NULL REFERENCES employees(id),
 change_type text NOT NULL CHECK (change_type IN ('MODEL','PROVIDER','PROMPT','SKILL','TOOL_POLICY')),
 baseline_version text NOT NULL, candidate_version text NOT NULL,
 evaluation_reference text NOT NULL, rollback_version text,
 evidence_json jsonb NOT NULL DEFAULT '{}'::jsonb, status text NOT NULL DEFAULT 'DRAFT',
 approval_id uuid REFERENCES approval_requests(id),
 created_at timestamptz NOT NULL DEFAULT now(), version bigint NOT NULL DEFAULT 0
);
CREATE INDEX ix_trigger_owner ON automation_triggers(organization_id, owner_employee_id);
CREATE INDEX ix_proposals_employee ON optimization_proposals(employee_id,created_at DESC);
-- TODO: enforce organization FK consistency and schema-level validity in production migrations.
-- TODO: permission checks, trigger rate + budget atomicity, validated state transitions, crypto, RLS.
-- TODO: ensure threshold notifications match policy version/window when later budget revisions are added.

-- ===== Existing platform registry reference tables (starter only) =====
-- Retains platform-managed credential *references* rather than plaintext secrets.
CREATE TABLE provider_registry (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 provider_alias text NOT NULL, endpoint_alias text, secret_ref text NOT NULL,
 status text NOT NULL DEFAULT 'DISABLED', created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE (organization_id, provider_alias)
);
CREATE TABLE mcp_servers (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 name text NOT NULL, transport text NOT NULL, endpoint_alias text,
 credential_ref text, schema_fingerprint text NOT NULL,
 review_status text NOT NULL DEFAULT 'PENDING', enabled boolean NOT NULL DEFAULT false,
 UNIQUE (organization_id,name)
);
CREATE TABLE mcp_tools (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), server_id uuid NOT NULL REFERENCES mcp_servers(id),
 qualified_name text NOT NULL, schema_hash text NOT NULL, risk_tier text NOT NULL,
 operation_type text NOT NULL, approval_required boolean NOT NULL DEFAULT false,
 UNIQUE(server_id,qualified_name,schema_hash)
);
CREATE TABLE agent_skills (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 skill_name text NOT NULL, skill_version text NOT NULL, content_hash text NOT NULL,
 content_ref text NOT NULL, review_status text NOT NULL DEFAULT 'PENDING',
 required_tool_refs jsonb NOT NULL DEFAULT '[]'::jsonb,
 UNIQUE(organization_id,skill_name,skill_version)
);
CREATE TABLE capability_grants (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), organization_id uuid NOT NULL REFERENCES organizations(id),
 subject_type text NOT NULL CHECK (subject_type IN ('POSITION','EMPLOYEE','TASK')),
 subject_id uuid NOT NULL, capability_type text NOT NULL CHECK(capability_type IN ('MCP','TOOL','SKILL')),
 capability_ref text NOT NULL, access_type text NOT NULL CHECK(access_type IN ('ALLOW','DENY')),
 scope_policy jsonb NOT NULL DEFAULT '{}'::jsonb, approval_id uuid REFERENCES approval_requests(id),
 expires_at timestamptz, policy_version bigint NOT NULL DEFAULT 0
);
-- TODO: enforce subject FK and same-org integrity via entity-specific FKs/policy service.
-- TODO: grant should never overrun position ceiling unless human-approved scoped exception.
