#!/usr/bin/env python3
"""Evaluate captured Compose/PostgreSQL responses without changing database data."""
import csv
import json
from pathlib import Path

base = Path(__file__).resolve().parent
with (base / "http-status.csv").open() as handle:
    codes = {row["name"]: int(row["status"]) for row in csv.DictReader(handle)}


def response(name):
    return json.loads((base / f"{name}.json").read_text())


results = []


def check(case_id, expected, actual, passed, requirements):
    results.append({"id": case_id, "status": "PASS" if passed else "FAIL",
                    "expected": expected, "actual": actual, "requirements": requirements})


health = response("health")
check("QA-CP-01", "HTTP 200 and status=ok", f"HTTP {codes['health']}; {health.get('status')}",
      codes["health"] == 200 and health.get("status") == "ok", ["NFR-REL-001"])

seed = dict(line.split("=", 1) for line in (base / "seed-counts.txt").read_text().splitlines())
company = response("company")
check("QA-CP-02", "Postgres seed: 9 employees, 15 tasks, 3 projects; API 15 distinct tasks",
      f"DB {seed}; API {len(company)} tasks, {len({task['id'] for task in company})} IDs",
      seed == {"employees": "9", "tasks": "15", "projects": "3"}
      and len(company) == len({task["id"] for task in company}) == 15,
      ["FR-WORK-004", "AT-03"])

p1 = response("project_p1")
check("QA-CP-03", "Project p1 returns exactly its 6 task IDs",
      f"HTTP {codes['project_p1']}; IDs={[task['id'] for task in p1]}",
      codes["project_p1"] == 200 and len(p1) == 6
      and all(task["project"] == "p1" for task in p1), ["AT-04"])

insights = response("team_insights")
expected_insights = {"T-107", "T-108", "T-109", "T-110", "T-111", "T-115"}
check("QA-CP-04", "Team insights sees six seeded linked tasks",
      f"HTTP {codes['team_insights']}; IDs={sorted(task['id'] for task in insights)}",
      codes["team_insights"] == 200 and {task["id"] for task in insights} == expected_insights,
      ["AT-03", "AT-04"])

summary = response("summary_p1")
check("QA-CP-05", "Project summary count=6, accepted=1, spent=18.00",
      f"HTTP {codes['summary_p1']}; {summary}",
      codes["summary_p1"] == 200 and summary["totalTasks"] == 6
      and summary["acceptedTasks"] == 1 and abs(summary["totalSpentUsd"] - 18) < 1e-6,
      ["AT-04", "AT-19"])

for case_id, name, req in [
    ("QA-CP-06", "tasks_anonymous", "FR-SEC-001"),
    ("QA-CP-07", "employees_fake", "AT-17"),
    ("QA-CP-08", "org_anonymous", "FR-SEC-001"),
]:
    check(case_id, "HTTP 401/403", f"HTTP {codes[name]}",
          codes[name] in (401, 403), [req])

check("QA-CP-09", "Malformed/unauthorized scope rejected with HTTP 400/403",
      f"HTTP {codes['invalid_scope']}; body={response('invalid_scope')}",
      codes["invalid_scope"] in (400, 403), ["BR-WORK-007"])

check("QA-CP-10", "company:anything is not an authorized all-company scope",
      f"HTTP {codes['company_fake']}; task count={len(response('company_fake'))}",
      codes["company_fake"] in (400, 403), ["BR-WORK-007"])

tables = set((base / "schema-tables.txt").read_text().splitlines())
required_finance = {"budget_accounts", "budget_reservations", "usage_ledger"}
check("QA-CP-11", "Postgres has reservation and immutable usage ledger tables",
      f"missing={sorted(required_finance - tables)}", required_finance <= tables,
      ["FR-BUD-001", "FR-BUD-002", "AT-07", "AT-08"])

build_log = (base / "frontend-docker-build.log").read_text()
check("QA-CP-12", "Docker frontend image builds with npm ci",
      "npm ci lockfile mismatch" if "npm error Invalid: lock file" in build_log else "see build log",
      "Image ai-business-os-spec-v11-frontend Built" in build_log,
      ["C-08", "08-delivery release gate"])

created = response("qa_create_task")
readback = response("qa_created_task_readback")
check("QA-CP-13", "Task created by API persists in separate PostgreSQL QA database",
      f"POST {codes['qa_create_task']}; GET {codes['qa_created_task_readback']}; id={created.get('id')}",
      codes["qa_create_task"] == 201 and codes["qa_created_task_readback"] == 200
      and created.get("id") == readback.get("id"), ["FR-WORK-006", "AT-05"])
check("QA-CP-14", "Create response returns persisted organization, timestamp, version",
      f"POST org={created.get('organizationId')}, createdAt={created.get('createdAt')}, version={created.get('version')}; GET org={readback.get('organizationId')}, version={readback.get('version')}",
      all(created.get(key) == readback.get(key) for key in ("organizationId", "createdAt", "version")),
      ["FR-WORK-006"])

check("QA-CP-15", "Unknown manager rejected in PostgreSQL",
      f"HTTP {codes['qa_bad_manager']}", codes["qa_bad_manager"] in (400, 409),
      ["BR-ORG-001", "AT-02"])
check("QA-CP-16", "BACKLOG accept and direct ACCEPTED patch rejected",
      f"accept={codes['qa_accept_backlog']}; patch={codes['qa_patch_accepted']}",
      codes["qa_accept_backlog"] in (400, 403, 409)
      and codes["qa_patch_accepted"] in (400, 403, 409), ["BR-WORK-003", "AT-16"])
check("QA-CP-17", "READY to COMPLETED jump rejected",
      f"HTTP {codes['qa_patch_completed']}", codes["qa_patch_completed"] in (400, 403, 409),
      ["BR-WORK-006"])
check("QA-CP-18", "Anonymous in-review acceptance rejected",
      f"HTTP {codes['qa_accept_anonymous']}", codes["qa_accept_anonymous"] in (401, 403),
      ["BR-WORK-003", "AT-16"])
check("QA-CP-19", "Forged token cannot approve as Founder",
      f"HTTP {codes['qa_approval_fake']}", codes["qa_approval_fake"] in (401, 403),
      ["BR-AUTH-001", "AT-22"])

low_emp = response("qa_low_budget_employee")
check("QA-CP-20", "Single run estimated over $0.10 cap denied, spent stays zero",
      f"HTTP {codes['qa_execute_low_budget']}; spent={low_emp.get('spent')}",
      codes["qa_execute_low_budget"] in (400, 403) and low_emp.get("spent") == 0,
      ["BR-BUD-003"])
parallel_codes = [codes[f"qa_create_parallel_{i}"] for i in (1, 2, 3)]
check("QA-CP-21", "Three concurrent create-task requests get distinct IDs",
      f"HTTP statuses={parallel_codes}; errors={[response(f'qa_create_parallel_{i}').get('error') for i in (1, 2, 3)]}",
      parallel_codes == [201, 201, 201]
      and len({response(f"qa_create_parallel_{i}").get("id") for i in (1, 2, 3)}) == 3,
      ["FR-WORK-002", "FR-WORK-006"])

(base / "results.json").write_text(json.dumps(results, ensure_ascii=False, indent=2) + "\n")
print(f"{sum(r['status'] == 'PASS' for r in results)} PASS, "
      f"{sum(r['status'] == 'FAIL' for r in results)} FAIL")
