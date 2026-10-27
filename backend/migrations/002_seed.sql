-- AI Business OS Canonical Seed Data

INSERT INTO organizations (id, name) VALUES
('org-default', 'AI Business OS Enterprise')
ON CONFLICT (id) DO NOTHING;

INSERT INTO departments (id, organization_id, name) VALUES
('tech', 'org-default', 'Technology'),
('biz', 'org-default', 'Business & Research'),
('ops', 'org-default', 'Operations')
ON CONFLICT (id) DO NOTHING;

INSERT INTO teams (id, organization_id, department_id, name) VALUES
('platform', 'org-default', 'tech', 'Platform'),
('quality', 'org-default', 'tech', 'Quality Engineering'),
('insights', 'org-default', 'biz', 'Market Insights'),
('revenue', 'org-default', 'biz', 'Growth'),
('finance', 'org-default', 'ops', 'Finance Ops')
ON CONFLICT (id) DO NOTHING;

INSERT INTO positions (id, organization_id, department_id, title, role_code, description, cap_mcps) VALUES
('pos-exec', 'org-default', 'ops', 'Founder / CEO', 'Executive', 'Executive corporate governance', '["task-mcp","budget-mcp","docs-mcp"]'),
('pos-biz-mgr', 'org-default', 'biz', 'Business Manager', 'AI Manager', 'Business management & agent coordination', '["task-mcp","budget-mcp","docs-mcp"]'),
('pos-cto', 'org-default', 'tech', 'Head of Technology', 'Tech Lead', 'Technology leadership and architecture', '["git-mcp","task-mcp","test-mcp"]'),
('pos-res', 'org-default', 'biz', 'Research Analyst', 'Researcher', 'Market analysis and data synthesis', '["web-mcp","docs-mcp"]'),
('pos-mkt', 'org-default', 'biz', 'Marketing Specialist', 'Marketing', 'Content, campaigns, and audience acquisition', '["web-mcp","docs-mcp"]'),
('pos-fin', 'org-default', 'ops', 'Cost Analyst', 'Finance', 'Budgeting, ledger auditing, and variance', '["budget-mcp","docs-mcp"]'),
('pos-be', 'org-default', 'tech', 'Backend Engineer', 'Java Engineer', 'Service implementation and integrations', '["git-mcp","shell-mcp","test-mcp"]'),
('pos-fe', 'org-default', 'tech', 'Frontend Engineer', 'Angular Engineer', 'User experience and web console architecture', '["git-mcp","browser-mcp","test-mcp"]'),
('pos-qa', 'org-default', 'tech', 'QA Engineer', 'QA Automation', 'Testing, regression, and quality gates', '["git-mcp","test-mcp","browser-mcp"]')
ON CONFLICT (id) DO NOTHING;

INSERT INTO employees (id, organization_id, department_id, position_id, display_name, kind, state, title, role_code, manager_id, model_name, provider_name, skills, mcp_grants, team_ids, kpi_score, budget_usd, spent_usd) VALUES
('h1', 'org-default', 'ops', 'pos-exec', 'Founder', 'HUMAN', 'ACTIVE', 'Founder / CEO', 'Executive', NULL, 'Human', '-', '["Strategy","Governance"]', '[]', '[]', 98, 0, 0),
('m1', 'org-default', 'biz', 'pos-biz-mgr', 'Maya Manager', 'AI', 'ACTIVE', 'Business Manager', 'AI Manager', 'h1', 'Model X', 'Provider A', '["Planning","Delegation","Team Review"]', '["task-mcp","budget-mcp"]', '["insights","revenue"]', 93, 45, 23.40),
('m2', 'org-default', 'tech', 'pos-cto', 'Theo CTO', 'AI', 'ACTIVE', 'Head of Technology', 'Tech Lead', 'h1', 'Model Z', 'Provider B', '["Architecture","Code Review"]', '["git-mcp","task-mcp"]', '["platform","quality"]', 91, 75, 42.30),
('r1', 'org-default', 'biz', 'pos-res', 'Nora Research', 'AI', 'ACTIVE', 'Research Analyst', 'Researcher', 'm1', 'Model Y', 'Provider A', '["Market Research","Source Verification"]', '["web-mcp","docs-mcp"]', '["insights"]', 96, 20, 12.82),
('mk1', 'org-default', 'biz', 'pos-mkt', 'Liam Growth', 'AI', 'ACTIVE', 'Marketing Specialist', 'Marketing', 'm1', 'Model Y', 'Provider A', '["Copywriting","SEO"]', '["web-mcp","docs-mcp"]', '["revenue"]', 90, 20, 9.43),
('f1', 'org-default', 'ops', 'pos-fin', 'Iris Finance', 'AI', 'ACTIVE', 'Cost Analyst', 'Finance', 'h1', 'Local 9B', 'Ollama', '["Reconciliation","Budget Forecast"]', '["budget-mcp"]', '["finance"]', 94, 8, 2.61),
('b1', 'org-default', 'tech', 'pos-be', 'Kai Backend', 'AI', 'ACTIVE', 'Backend Engineer', 'Java Engineer', 'm2', 'Model Z', 'Provider B', '["Spring Boot","SQL","Unit Testing"]', '["git-mcp","shell-mcp"]', '["platform"]', 92, 35, 19.35),
('b2', 'org-default', 'tech', 'pos-fe', 'Ava Frontend', 'AI', 'ACTIVE', 'Frontend Engineer', 'Angular Engineer', 'm2', 'Model X', 'Provider A', '["Angular","UX","Accessibility"]', '["git-mcp","browser-mcp"]', '["platform"]', 95, 30, 17.20),
('q1', 'org-default', 'tech', 'pos-qa', 'Zed Quality', 'AI', 'ACTIVE', 'QA Engineer', 'QA Automation', 'm2', 'Model Y', 'Provider A', '["Playwright","Regression","Test Design"]', '["git-mcp","test-mcp"]', '["quality"]', 97, 18, 8.15)
ON CONFLICT (id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    title = EXCLUDED.title,
    role_code = EXCLUDED.role_code,
    manager_id = EXCLUDED.manager_id,
    model_name = EXCLUDED.model_name,
    provider_name = EXCLUDED.provider_name,
    skills = EXCLUDED.skills,
    mcp_grants = EXCLUDED.mcp_grants,
    team_ids = EXCLUDED.team_ids,
    kpi_score = EXCLUDED.kpi_score,
    budget_usd = EXCLUDED.budget_usd,
    spent_usd = EXCLUDED.spent_usd;

INSERT INTO projects (id, organization_id, code, name, description, accountable_owner_id, status, budget_usd, spent_usd, due_at, members, teams, milestones) VALUES
('p1', 'org-default', 'ABOS-CORE', 'AI Business OS Platform', 'Xây lõi tổ chức, quản trị agent, task, budget và web console.', 'm2', 'ACTIVE', 130, 78.9, '2026-12-15', '["m2","b1","b2","q1","m1"]', '["platform","quality","insights"]', '[{"name":"Domain & API contract","done":true},{"name":"Employee + task vertical slice","done":true},{"name":"Budget + MCP authorization","done":false},{"name":"Acceptance and observability","done":false}]'),
('p2', 'org-default', 'MARKET-26', 'AI Market Intelligence', 'Thử nghiệm team AI nghiên cứu thị trường và tổng hợp báo cáo đa lĩnh vực.', 'm1', 'ACTIVE', 55, 27.4, '2026-11-30', '["m1","r1","mk1"]', '["insights","revenue"]', '[{"name":"Research protocol","done":true},{"name":"Weekly monitoring","done":false},{"name":"Executive report","done":false}]'),
('p3', 'org-default', 'COST-OPT', 'Model Cost Optimization', 'Benchmark đa provider, tối ưu cost per accepted task, approval và rollback.', 'f1', 'PLANNED', 30, 11.2, '2026-12-30', '["f1","r1","m2","q1"]', '["finance","insights","quality"]', '[{"name":"Evaluation dataset","done":true},{"name":"Provider benchmark","done":false},{"name":"Canary dashboard","done":false}]')
ON CONFLICT (id) DO UPDATE SET
    code = EXCLUDED.code,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    accountable_owner_id = EXCLUDED.accountable_owner_id,
    status = EXCLUDED.status,
    budget_usd = EXCLUDED.budget_usd,
    spent_usd = EXCLUDED.spent_usd,
    due_at = EXCLUDED.due_at,
    members = EXCLUDED.members,
    teams = EXCLUDED.teams,
    milestones = EXCLUDED.milestones;

INSERT INTO tasks (id, organization_id, project_id, title, description, accountable_owner_id, collaborators, status, priority, acceptance_criteria, due_at, cost_usd, execs_count) VALUES
('T-101', 'org-default', 'p1', 'Thiết kế Employee / Role domain', 'Xác định bounded context Employee, Position, Role và Reporting lines', 'm2', '["b1"]', 'accepted', 'HIGH', 'Entity rõ và có tests', '2026-10-30', 2.6, 2),
('T-102', 'org-default', 'p1', 'Build organization chart API', 'API trả cấu trúc phân cấp công ty, kiểm tra quyền xem', 'b1', '["b2"]', 'in_progress', 'HIGH', 'API permission-filtered trả node/edge', '2026-11-02', 4.3, 3),
('T-103', 'org-default', 'p1', 'Triển khai Angular agent detail drawer', 'Drawer hiển thị thông tin nhân sự AI, model, MCP và KPIs', 'b2', '[]', 'in_review', 'HIGH', 'Profile với model MCP skills tasks KPI', '2026-11-03', 3.2, 2),
('T-104', 'org-default', 'p1', 'Test phân quyền Tool Gateway', 'Kiểm tra rule Org ∩ Position ∩ Employee ∩ Task', 'q1', '["b1"]', 'ready', 'CRITICAL', 'Permission denied / audit tests', '2026-11-06', 0.4, 0),
('T-105', 'org-default', 'p1', 'Budget reservation concurrency', 'Đảm bảo khóa ngân sách nguyên tử chống vượt hạn mức đồng thời', 'b1', '["f1"]', 'blocked', 'HIGH', 'No overspend under concurrent calls', '2026-11-05', 5.7, 4),
('T-106', 'org-default', 'p1', 'CI/CD Compose environment', 'Thiết lập Docker Compose reproducible staging cho dev và test', 'm2', '["b1"]', 'completed', 'MEDIUM', 'Repeatable staging setup', '2026-10-29', 1.8, 2),
('T-107', 'org-default', 'p2', 'Market research competitor shortlist', 'Khảo sát và lập danh sách 10 đối thủ hàng đầu trên thị trường', 'r1', '["mk1"]', 'in_progress', 'HIGH', '10 verified competitors + sources', '2026-11-02', 2.4, 3),
('T-108', 'org-default', 'p2', 'Weekly social trends report', 'Theo dõi xu hướng thảo luận và từ khóa tuần', 'r1', '[]', 'backlog', 'MEDIUM', 'Report with confidence label', '2026-11-10', 0.0, 0),
('T-109', 'org-default', 'p2', 'Landing page positioning draft', 'Bản thảo định vị sản phẩm cho landing page B2B', 'mk1', '["m1"]', 'in_review', 'MEDIUM', '2 tested messages and targeting rationale', '2026-11-05', 1.25, 1),
('T-110', 'org-default', 'p2', 'Research approval for external publishing', 'Yêu cầu phê duyệt công bố báo cáo ra công chúng', 'm1', '[]', 'blocked', 'HIGH', 'Human approval record and publication evidence', '2026-11-04', 0.3, 1),
('T-111', 'org-default', 'p3', 'Build benchmark golden dataset', 'Xây dựng bộ dữ liệu đánh giá chất lượng model cho các task chuẩn', 'q1', '["r1"]', 'accepted', 'HIGH', 'Versioned dataset with rubrics', '2026-10-31', 1.65, 1),
('T-112', 'org-default', 'p3', 'Provider price and quality analysis', 'Phân tích bảng giá và tỷ lệ thành công của các nhà cung cấp', 'f1', '["m2"]', 'ready', 'HIGH', 'Quality threshold and cost comparison', '2026-11-06', 0.5, 0),
('T-113', 'org-default', 'p3', 'Model swap approval workflow', 'Quy trình đề xuất và phê duyệt chuyển đổi model mặc định', 'm2', '["f1"]', 'backlog', 'MEDIUM', 'Audit approved change only', '2026-11-14', 0.0, 0),
('T-114', 'org-default', NULL, 'Review internal operating expenses', 'Đối soát chi phí vận hành hạ tầng và quota mô hình nội bộ', 'f1', '[]', 'ready', 'LOW', 'Reconciled expenses report', '2026-11-08', 0.31, 0),
('T-115', 'org-default', NULL, 'Executive weekly progress report', 'Báo cáo tổng kết tiến độ tuần gửi Founder / CEO', 'm1', '["m2"]', 'in_progress', 'MEDIUM', 'All projects summarised without double-count', '2026-10-30', 0.7, 1)
ON CONFLICT (id) DO NOTHING;

INSERT INTO approval_requests (id, organization_id, task_id, requested_by, action_kind, risk_tier, status, title) VALUES
('APR-12', 'org-default', 'T-110', 'm1', 'External write via docs-mcp', 'HIGH', 'PENDING', 'Publish external research report'),
('APR-13', 'org-default', NULL, 'm1', 'Model assignment change', 'MEDIUM', 'PENDING', 'Change primary model for Nora Research'),
('APR-14', 'org-default', 'T-105', 'm2', 'Budget increase', 'HIGH', 'PENDING', 'Increase project budget ABOS-CORE +$20')
ON CONFLICT (id) DO NOTHING;

INSERT INTO mcp_servers (id, organization_id, name, transport, description, risk_tier, enabled, review_status) VALUES
('mcp-1', 'org-default', 'web-mcp', 'stdio', 'Search, crawl and retrieve external web content', 'R1', true, 'APPROVED'),
('mcp-2', 'org-default', 'git-mcp', 'stdio', 'Read repo, commit changes, branch and PR management', 'R2', true, 'APPROVED'),
('mcp-3', 'org-default', 'shell-mcp', 'docker-sandbox', 'Isolated shell command runner within policy', 'R3', true, 'APPROVED'),
('mcp-4', 'org-default', 'budget-mcp', 'stdio', 'Ledger query, spend forecasts and ratecard inspect', 'R1', true, 'APPROVED'),
('mcp-5', 'org-default', 'docs-mcp', 'stdio', 'Read and write enterprise documentation repository', 'R2', true, 'APPROVED'),
('mcp-6', 'org-default', 'task-mcp', 'internal', 'Query, delegate and report task lifecycle status', 'R2', true, 'APPROVED')
ON CONFLICT (id) DO NOTHING;

INSERT INTO model_versions (id, provider_code, model_id, revision, features, status) VALUES
('mv-1', 'Provider A', 'Model X', '2026-q3', '["Tool Calling","JSON Mode","Fast Latency"]', 'ACTIVE'),
('mv-2', 'Provider A', 'Model Y', '2026-q3', '["Long Context","Reasoning","Tool Calling"]', 'ACTIVE'),
('mv-3', 'Provider B', 'Model Z', '2026-q4', '["Parallel Tools","High Coding Benchmark"]', 'ACTIVE'),
('mv-4', 'Ollama', 'Local 9B', '2026-v2', '["Local Privacy","Low Cost","Basic Tools"]', 'ACTIVE')
ON CONFLICT (id) DO NOTHING;
