-- AI Business OS Initial Schema
CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$ BEGIN
    CREATE TYPE employee_kind AS ENUM ('HUMAN', 'AI');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE employee_state AS ENUM ('DRAFT', 'ACTIVE', 'SUSPENDED', 'RETIRED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE task_state AS ENUM ('DRAFT', 'backlog', 'ready', 'in_progress', 'blocked', 'in_review', 'completed', 'accepted', 'rejected', 'cancelled');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE execution_state AS ENUM ('CREATED', 'QUEUED', 'RUNNING', 'WAITING_APPROVAL', 'WAITING_TOOL', 'SUCCEEDED', 'FAILED', 'TIMED_OUT', 'CANCELLED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS organizations (
    id text PRIMARY KEY,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS departments (
    id text PRIMARY KEY,
    organization_id text NOT NULL REFERENCES organizations(id),
    name text NOT NULL,
    parent_department_id text REFERENCES departments(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, name)
);

CREATE TABLE IF NOT EXISTS teams (
    id text PRIMARY KEY,
    organization_id text NOT NULL REFERENCES organizations(id),
    department_id text NOT NULL REFERENCES departments(id),
    name text NOT NULL,
    lead_employee_id text,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, name)
);

CREATE TABLE IF NOT EXISTS positions (
    id text PRIMARY KEY,
    organization_id text NOT NULL REFERENCES organizations(id),
    department_id text REFERENCES departments(id),
    title text NOT NULL,
    role_code text NOT NULL,
    description text NOT NULL DEFAULT '',
    cap_mcps jsonb NOT NULL DEFAULT '[]'::jsonb
);

CREATE TABLE IF NOT EXISTS employees (
    id text PRIMARY KEY,
    organization_id text NOT NULL REFERENCES organizations(id),
    department_id text REFERENCES departments(id),
    position_id text REFERENCES positions(id),
    display_name text NOT NULL,
    kind employee_kind NOT NULL DEFAULT 'AI',
    state employee_state NOT NULL DEFAULT 'ACTIVE',
    title text NOT NULL,
    role_code text NOT NULL,
    manager_id text REFERENCES employees(id),
    model_name text NOT NULL DEFAULT '',
    provider_name text NOT NULL DEFAULT '',
    skills jsonb NOT NULL DEFAULT '[]'::jsonb,
    mcp_grants jsonb NOT NULL DEFAULT '[]'::jsonb,
    team_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
    kpi_score integer NOT NULL DEFAULT 90,
    budget_usd numeric(20,4) NOT NULL DEFAULT 0,
    spent_usd numeric(20,4) NOT NULL DEFAULT 0,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS projects (
    id text PRIMARY KEY,
    organization_id text NOT NULL REFERENCES organizations(id),
    code text NOT NULL,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    accountable_owner_id text NOT NULL REFERENCES employees(id),
    status text NOT NULL DEFAULT 'ACTIVE',
    budget_usd numeric(20,4) NOT NULL DEFAULT 0,
    spent_usd numeric(20,4) NOT NULL DEFAULT 0,
    due_at text NOT NULL,
    members jsonb NOT NULL DEFAULT '[]'::jsonb,
    teams jsonb NOT NULL DEFAULT '[]'::jsonb,
    milestones jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    version bigint NOT NULL DEFAULT 1,
    UNIQUE (organization_id, code)
);

CREATE TABLE IF NOT EXISTS tasks (
    id text PRIMARY KEY,
    organization_id text NOT NULL REFERENCES organizations(id),
    project_id text REFERENCES projects(id),
    parent_task_id text REFERENCES tasks(id),
    title text NOT NULL,
    description text NOT NULL DEFAULT '',
    accountable_owner_id text NOT NULL REFERENCES employees(id),
    collaborators jsonb NOT NULL DEFAULT '[]'::jsonb,
    status task_state NOT NULL DEFAULT 'ready',
    priority text NOT NULL DEFAULT 'MEDIUM',
    acceptance_criteria text NOT NULL DEFAULT '',
    due_at text NOT NULL,
    accepted_at timestamptz,
    cost_usd numeric(20,4) NOT NULL DEFAULT 0,
    execs_count integer NOT NULL DEFAULT 0,
    origin_type text NOT NULL DEFAULT 'HUMAN',
    origin_ref text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    version bigint NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS executions (
    id text PRIMARY KEY,
    organization_id text NOT NULL REFERENCES organizations(id),
    task_id text NOT NULL REFERENCES tasks(id),
    executing_employee_id text NOT NULL REFERENCES employees(id),
    status execution_state NOT NULL DEFAULT 'CREATED',
    attempt_no integer NOT NULL DEFAULT 1,
    idempotency_key text NOT NULL,
    logs jsonb NOT NULL DEFAULT '[]'::jsonb,
    steps jsonb NOT NULL DEFAULT '[]'::jsonb,
    total_tokens bigint NOT NULL DEFAULT 0,
    total_cost_usd numeric(20,4) NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    started_at timestamptz,
    ended_at timestamptz
);

CREATE TABLE IF NOT EXISTS approval_requests (
    id text PRIMARY KEY,
    organization_id text NOT NULL REFERENCES organizations(id),
    task_id text REFERENCES tasks(id),
    requested_by text NOT NULL REFERENCES employees(id),
    action_kind text NOT NULL,
    payload_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    risk_tier text NOT NULL DEFAULT 'MEDIUM',
    status text NOT NULL DEFAULT 'PENDING',
    title text NOT NULL,
    expires_at timestamptz,
    resolved_by text REFERENCES employees(id),
    resolved_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS mcp_servers (
    id text PRIMARY KEY,
    organization_id text NOT NULL REFERENCES organizations(id),
    name text NOT NULL,
    transport text NOT NULL DEFAULT 'stdio',
    description text NOT NULL DEFAULT '',
    risk_tier text NOT NULL DEFAULT 'R1',
    enabled boolean NOT NULL DEFAULT true,
    review_status text NOT NULL DEFAULT 'APPROVED',
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, name)
);

CREATE TABLE IF NOT EXISTS model_versions (
    id text PRIMARY KEY,
    provider_code text NOT NULL,
    model_id text NOT NULL,
    revision text NOT NULL,
    supports_tools boolean NOT NULL DEFAULT true,
    supports_structured_output boolean NOT NULL DEFAULT true,
    context_window integer DEFAULT 128000,
    price_currency char(3) NOT NULL DEFAULT 'USD',
    features jsonb NOT NULL DEFAULT '[]'::jsonb,
    status text NOT NULL DEFAULT 'ACTIVE'
);
