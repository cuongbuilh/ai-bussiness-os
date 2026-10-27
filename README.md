# AI Business OS — Internal Enterprise AI Workforce Platform

**Spec version:** 1.1-confirmed-baseline | **Baseline date:** 2026-10-27 | **Language:** Vietnamese | **Tech Stack:** Golang 1.24 + Angular 18 + PostgreSQL 16 + Docker Compose.

---

## 🚀 Khởi động nhanh bằng Docker Compose

Hệ thống đã được đóng gói và cấu hình đầy đủ bằng Docker Compose:

```bash
# 1. Khởi chạy toàn bộ hệ thống (PostgreSQL, Golang Backend, Angular Web Console)
docker compose up --build -d

# 2. Kiểm tra trạng thái các container
docker compose ps
```

| Dịch vụ | URL / Cổng kết nối | Mô tả |
|---|---|---|
| **Web Console** | [http://localhost:3000](http://localhost:3000) | Giao diện điều hành Angular SPA trực quan |
| **Control Plane API** | [http://localhost:8080/api/v1](http://localhost:8080/api/v1) | Backend REST API và luồng SSE realtime |
| **PostgreSQL Database** | `localhost:5432` | Cơ sở dữ liệu quan hệ (`abos_db`, user: `abos`) |

---

## 🔐 Hướng dẫn Đăng nhập & Xác thực (Authentication)

Hệ thống hỗ trợ cơ chế xác thực Bearer token kết hợp trang đăng nhập thân thiện trên Web Console:

### 1. Đăng nhập qua Web Console ([http://localhost:3000](http://localhost:3000))
Truy cập trình duyệt tại `http://localhost:3000`:
- **Tài khoản mặc định Founder / CEO:**
  - **Tên đăng nhập:** `founder` (hoặc `admin`, `h1`)
  - **Mật khẩu:** `founder123`
  - Hoặc nhấp nút **⚡ Đăng nhập nhanh tài khoản Founder / CEO** để đăng nhập trực tiếp.
- **Đăng nhập bằng Access Token:** Chuyển sang tab *Access Token* để nhập token tùy chỉnh.
- **Đăng xuất:** Nhấp nút **Đăng xuất** cạnh tên người dùng trên thanh toolbar trên cùng.

### 2. Đăng nhập qua REST API
```bash
# Đăng nhập nhận Bearer token và thông tin người dùng
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"founder","password":"founder123"}'

# Phản hồi:
# {"token":"founder-h1-token","user":{"id":"h1","displayName":"Founder","role":"Executive","kind":"HUMAN","title":"Founder / CEO"}}

# Gọi API với token nhận được
curl -s http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer founder-h1-token"
```

---

## Product intent
Nền tảng vận hành doanh nghiệp nội bộ có 1..N nhân sự AI/người thật; sơ đồ quản lý trực tiếp; giao việc, dự án, team; MCP/tool/skills theo position và agent; ngân sách tiền/token/quota/thời gian/task; đánh giá nhân viên và benchmark model/provider; phê duyệt thay model; điều hành trực quan trên web.

## Confirmed decisions (C)
- C-01: Một organization nội bộ trước mắt (single-tenant), dữ liệu có khóa organization_id để dễ mở rộng.
- C-02: Đa lĩnh vực, position/role/skills tự định nghĩa, không hard-code công ty phần mềm.
- C-03: Autonomous in-policy, human approval for sensitive actions.
- C-04: Mỗi employee có tối đa một direct manager tại một thời điểm; cộng tác chéo team/phòng/dự án.
- C-05: Budget đồng thời tiền, token, quotas, time, task counts.
- C-06: Provider/API credentials do platform quản lý tập trung.
- C-07: Benchmark tự động, đề xuất model mới, thay model chính cần duyệt.
- C-08: Self-host Docker Compose trước, Kubernetes sau.
- C-09: MCP/tools/skills ở level position + agent; capability gates theo model; tool-call audit.
- C-10: Web console có org-chart click-through, agent profile, model, KPI, task và giao việc.
- C-11: Unified portfolio/project/task boards, drill-down theo company, department, team, project, employee, manager, role.
- C-12: AI Manager đề xuất hiring/position/org changes; thay đổi quan trọng cần human approval.
- C-13: Task được tạo bởi human, manager, schedule hoặc event trong policy; proactive work được quản lý.
- C-14: Nhiều agent cộng tác trên task; một owner chịu trách nhiệm chính và có thể có collaborators.
- C-15: Position có MCP/tool/skills defaults và ceiling; agent-specific grants không vượt trần khi chưa được duyệt.
- C-16: Persistent scoped memory theo Employee/Team/Department/Project/Organization, có ACL.
- C-17: Cảnh báo ngân sách tại 80%, dừng hard limit, AI manager chỉ được đề xuất top-up.
- C-18: Evidence-based self-evaluation và proposals tối ưu prompt/skill/model, thay đổi quan trọng cần duyệt.
- C-19: Hỗ trợ cả Task-driven và Goal-driven; **MVP Task-driven; Goal-driven deferred**.

## Proposed implementation policies (P) — không thay thế các confirmed decisions
- P-01: Frontend Angular 18+ (đánh giá compatibility thực tế lúc build); Spring Boot control plane; Python agent runtime; PostgreSQL; queue; LiteLLM; LangGraph; Langfuse; MinIO. Không khóa version đến khi lập lockfile và security scan.
- P-02: Task mặc định có đúng 1 accountable owner và 0..N collaborators; project có 1 primary accountable project owner, được liên phòng/team.
- P-03: Các task status: DRAFT, BACKLOG, READY, IN_PROGRESS, BLOCKED, IN_REVIEW, COMPLETED, ACCEPTED, REJECTED, CANCELLED; execution status tách biệt.
- P-04: Cấu hình role baseline là ceiling; agent grants con không tăng quyền nếu chưa có exception được duyệt.
- P-05: Chi phí sổ cái đơn vị USD chuẩn, số tiền VND chỉ hiển thị theo tỷ giá snapshot; không nên lấy số token làm proxy duy nhất.
- P-06: Event streaming SSE trước, WebSocket tùy nhu cầu; polling fallback.
- P-07: Không tự deploy production, xuất dữ liệu nhạy cảm, xóa dữ liệu, thanh toán, cấp thêm quyền, đổi model chính, hoặc self-hire ngoài ngưỡng mà không duyệt.
- P-08: Một task chỉ có một primary project (nullable); xem theo scope khác là query/view, không sao chép task; cross-project work qua linked work items.

## Two-mode delivery contract
- **MVP (Task-driven):** user/manager assigned work, budget/tool-gated execution, approvals, event/schedule hooks có thể triển khai dần theo P0/P1.
- **Later (Goal-driven):** owner-defined measurable goals -> plans -> initiatives -> projects/tasks -> review/replan. Future schema/API là thiết kế, chưa được triển khai.

## Document index
0. `00-governance/BASELINE-AND-DECISIONS.md`, `CHANGELOG.md`, `REQUIREMENT-TRACEABILITY.md` — **confirmed decisions, change control and traceability**.
1. `01-product/PRD.md` — goals, personas, scope, functional/nonfunctional requirements.
1a. `01-product/TASK-AND-GOAL-ROADMAP.md` — dual-mode scope and Goal-driven future design.
2. `02-domain/DOMAIN-AND-BUSINESS-RULES.md` — bounded contexts, entities, invariants, state machines.
2a. `02-domain/INITIATIVES-TRIGGERS-AND-MEMORY.md` — confirmed HR, triggers, collaboration, memory, ceiling, threshold, improvement BRs.
3. `03-architecture/SOLUTION-ARCHITECTURE.md` — component and deployment architecture, trade-offs, reliability.
4. `04-data-api/DATA-MODEL.md` — ERD, constraints, aggregates, cost attribution.
5. `04-data-api/openapi-starter.yaml` — starter API contract (not exhaustive OpenAPI).
6. `04-data-api/schema-starter.sql` — illustrative schema and critical constraints (not production migration).
7. `04-data-api/EVENT-CONTRACTS.md` — event envelopes, idempotency and outbox.
7a. `04-data-api/CONTRACTS-V1-1.md` — new contract examples, future Goal-driven API registry.
7b. `04-data-api/schema-goal-driven-future.sql` — **future design only**, DO NOT run in MVP.
8. `05-agent-platform/AGENT-RUNTIME-MCP-SKILLS.md` — position/agent/model tools, grant resolution, executor.
8a. `05-agent-platform/POSITION-AGENT-CAPABILITY-EXAMPLES.md` — worked examples for role/agent MCP, skills, model permission gates.
9. `05-agent-platform/BUDGET-AND-EVALUATION.md` — ledger, hard cap, model scoring/selection & benchmark.
10. `06-web-console/UX-WEB-CONSOLE.md` — pages, board scopes, drawer details, interactions, acceptance.
11. `07-security-ops/SECURITY-OPERATIONS.md` — threat model, secrets, authz, observability, deployment.
12. `08-delivery/IMPLEMENTATION-PLAN.md` — milestone, backlog, agents, definition of done.
13. `08-delivery/ACCEPTANCE-TESTS.md` — executable-test outlines and requirements traceability.
14. `prototype/index.html` — clickable standalone UI mock (demo data, local-only, no backend).

### Source-of-truth policy
The canonical decisions are in `00-governance/BASELINE-AND-DECISIONS.md`, confirmed rules in `02-domain/INITIATIVES-TRIGGERS-AND-MEMORY.md`. Source API/SQL files remain illustrative starter contracts. Features in prototype are mocked until integration tests against runtime pass.

## Reading order & handoff
PRD -> Domain & BRs -> Architecture -> Data/API -> AI platform -> UI -> Security -> Delivery/Tests.

> **Important:** HTML prototype is a visual/interaction proof of concept, not actual secured production software. Mock token, cost and performance data are fabricated for demonstration. Database schema and OpenAPI are starter contracts, not equivalent to fully implemented services.

## Verified engineering references (as of 2026-10-27)
- LangGraph interrupts/durable pause: https://docs.langchain.com/oss/python/langgraph/interrupts
- MCP tools spec: https://modelcontextprotocol.io/specification/2025-06-18/server/tools
- LiteLLM budgets and database caveat: https://docs.litellm.ai/docs/proxy/users
- LiteLLM virtual keys and MCP entitlements: https://docs.litellm.ai/docs/proxy/virtual_keys
- Langfuse datasets/evaluation: https://langfuse.com/docs/evaluation/experiments/datasets
