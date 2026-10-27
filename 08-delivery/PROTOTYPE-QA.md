# Prototype smoke-test results (2026-10-27)

## Environment
- Single self-contained HTML with CSS/JavaScript, no external web requests/dependencies.
- Ran automated Chromium/Playwright interactions by injecting local HTML into sandboxed browser. Browser policy blocked navigating to local files and localhost, so tests used `page.set_content` with localStorage shim.

## Verified
- Overview renders with four KPIs, portfolio list, workforce and task health.
- Org chart contains 9 mock employees, click agent profile drawer.
- Employee detail tabs switch to MCP/Skills; position ceiling and agent-specific MCP grants visible.
- Assign task from employee profile -> new task visible in company board; scoped employee filtering shows relevant tasks.
- Kanban and Table switch; project detail opens; approval action controls rendered.
- JavaScript `node --check` passes; Playwright captured no uncaught page errors in tested flows.
- Screenshots: `prototype/preview.png`, `prototype/preview-board.png`, `prototype/preview-org.png`.

## Not implemented in prototype
- Real Spring Boot API, persisted PostgreSQL data, queue, LLM execution, actual MCP tool calls, identity, RBAC/ABAC, real finance ledger, benchmark, observability.
- Drag/status and approval are **demo-only actions**; production backend must validate lifecycle, reviewer authority and budget.

## Remaining implementation verification
- End-to-end backend API contract tests, SQL migrations, security tests, concurrency stress, state machine transition tests and production deployment. See `ACCEPTANCE-TESTS.md`.
