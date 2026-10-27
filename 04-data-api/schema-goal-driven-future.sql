-- FUTURE DESIGN ONLY (P2 Goal-driven). DO NOT execute as part of MVP migrations.
-- Requires review of FK tenancy, approval semantics, metric evaluator source, and limits.
CREATE TABLE business_goals (
 id uuid PRIMARY KEY, organization_id uuid NOT NULL, accountable_owner_id uuid NOT NULL,
 statement text NOT NULL, status text NOT NULL, autonomy_tier text NOT NULL,
 target_date timestamptz, budget_cap_usd numeric(20,8) NOT NULL,
 approved_plan_id uuid, created_at timestamptz NOT NULL, version bigint NOT NULL
);
CREATE TABLE goal_metrics (
 id uuid PRIMARY KEY, goal_id uuid NOT NULL, metric_code text NOT NULL,
 baseline_value numeric, target_value numeric NOT NULL, source_ref text NOT NULL,
 evaluator_version text NOT NULL, measurement_window jsonb NOT NULL
);
CREATE TABLE goal_plan_versions (
 id uuid PRIMARY KEY, goal_id uuid NOT NULL, author_employee_id uuid NOT NULL,
 plan_version integer NOT NULL, plan_json jsonb NOT NULL,
 budget_forecast_usd numeric(20,8), approval_id uuid,
 status text NOT NULL, UNIQUE(goal_id,plan_version)
);
CREATE TABLE initiatives (
 id uuid PRIMARY KEY, goal_id uuid NOT NULL, accountable_manager_id uuid NOT NULL,
 active_plan_id uuid NOT NULL, expected_impact jsonb NOT NULL,
 budget_cap_usd numeric(20,8), status text NOT NULL
);
CREATE TABLE goal_task_links (
 goal_id uuid NOT NULL, initiative_id uuid NOT NULL, task_id uuid NOT NULL,
 origin_plan_version_id uuid NOT NULL,
 PRIMARY KEY(goal_id,task_id)
);
-- Follow-up design: add full same-tenant FK constraints, event persistence, correct cascades,
-- no data duplication, goal budget overlay, stop/pause and human-approved activate/replan.
