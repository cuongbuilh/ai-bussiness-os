package repository

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cuongbuilh/ai-business-os/backend/internal/domain"
	"github.com/google/uuid"
)

type MemoryRepo struct {
	mu          sync.RWMutex
	org         domain.Organization
	departments map[string]domain.Department
	teams       map[string]domain.Team
	positions   map[string]domain.Position
	employees   map[string]domain.Employee
	projects    map[string]domain.Project
	tasks       map[string]domain.Task
	executions  map[string]domain.Execution
	approvals   map[string]domain.ApprovalRequest
	mcpServers  map[string]domain.MCPServer
	models      map[string]domain.ModelInfo
}

func NewMemoryRepo() *MemoryRepo {
	r := &MemoryRepo{
		departments: make(map[string]domain.Department),
		teams:       make(map[string]domain.Team),
		positions:   make(map[string]domain.Position),
		employees:   make(map[string]domain.Employee),
		projects:    make(map[string]domain.Project),
		tasks:       make(map[string]domain.Task),
		executions:  make(map[string]domain.Execution),
		approvals:   make(map[string]domain.ApprovalRequest),
		mcpServers:  make(map[string]domain.MCPServer),
		models:      make(map[string]domain.ModelInfo),
	}
	r.seed()
	return r
}

func (r *MemoryRepo) seed() {
	r.org = domain.Organization{
		ID:        "org-default",
		Name:      "AI Business OS Enterprise",
		CreatedAt: time.Now(),
	}

	depts := []domain.Department{
		{ID: "tech", OrganizationID: "org-default", Name: "Technology"},
		{ID: "biz", OrganizationID: "org-default", Name: "Business & Research"},
		{ID: "ops", OrganizationID: "org-default", Name: "Operations"},
	}
	for _, d := range depts {
		r.departments[d.ID] = d
	}

	teams := []domain.Team{
		{ID: "platform", OrganizationID: "org-default", DepartmentID: "tech", Name: "Platform"},
		{ID: "quality", OrganizationID: "org-default", DepartmentID: "tech", Name: "Quality Engineering"},
		{ID: "insights", OrganizationID: "org-default", DepartmentID: "biz", Name: "Market Insights"},
		{ID: "revenue", OrganizationID: "org-default", DepartmentID: "biz", Name: "Growth"},
		{ID: "finance", OrganizationID: "org-default", DepartmentID: "ops", Name: "Finance Ops"},
	}
	for _, t := range teams {
		r.teams[t.ID] = t
	}

	positions := []domain.Position{
		{ID: "pos-exec", OrganizationID: "org-default", Title: "Founder / CEO", RoleCode: "Executive", Description: "Executive governance", CapMCPs: []string{"task-mcp", "budget-mcp", "docs-mcp"}},
		{ID: "pos-biz-mgr", OrganizationID: "org-default", Title: "Business Manager", RoleCode: "AI Manager", Description: "Business operations coordination", CapMCPs: []string{"task-mcp", "budget-mcp", "docs-mcp"}},
		{ID: "pos-cto", OrganizationID: "org-default", Title: "Head of Technology", RoleCode: "Tech Lead", Description: "Technology architecture", CapMCPs: []string{"git-mcp", "task-mcp", "test-mcp"}},
		{ID: "pos-res", OrganizationID: "org-default", Title: "Research Analyst", RoleCode: "Researcher", Description: "Market insights", CapMCPs: []string{"web-mcp", "docs-mcp"}},
		{ID: "pos-mkt", OrganizationID: "org-default", Title: "Marketing Specialist", RoleCode: "Marketing", Description: "Growth & audience", CapMCPs: []string{"web-mcp", "docs-mcp"}},
		{ID: "pos-fin", OrganizationID: "org-default", Title: "Cost Analyst", RoleCode: "Finance", Description: "Budgets & accounting", CapMCPs: []string{"budget-mcp", "docs-mcp"}},
		{ID: "pos-be", OrganizationID: "org-default", Title: "Backend Engineer", RoleCode: "Java Engineer", Description: "Backend microservices", CapMCPs: []string{"git-mcp", "shell-mcp", "test-mcp"}},
		{ID: "pos-fe", OrganizationID: "org-default", Title: "Frontend Engineer", RoleCode: "Angular Engineer", Description: "Frontend UI", CapMCPs: []string{"git-mcp", "browser-mcp", "test-mcp"}},
		{ID: "pos-qa", OrganizationID: "org-default", Title: "QA Engineer", RoleCode: "QA Automation", Description: "Testing automation", CapMCPs: []string{"git-mcp", "test-mcp", "browser-mcp"}},
	}
	for _, p := range positions {
		r.positions[p.ID] = p
	}

	deptBiz := "biz"
	deptTech := "tech"
	deptOps := "ops"
	h1ID := "h1"
	m1ID := "m1"
	m2ID := "m2"

	employees := []domain.Employee{
		{ID: "h1", OrganizationID: "org-default", DepartmentID: &deptOps, DepartmentName: "Operations", DisplayName: "Founder", Name: "Founder", Kind: domain.EmployeeKindHuman, State: domain.EmployeeStateActive, Title: "Founder / CEO", Role: "Executive", ManagerID: nil, ManagerName: "None / Founder", Model: "Human", Provider: "-", Skills: []string{"Strategy", "Governance"}, MCP: []string{}, Teams: []string{}, KPI: 98, BudgetUSD: 0, SpentUSD: 0, Version: 1},
		{ID: "m1", OrganizationID: "org-default", DepartmentID: &deptBiz, DepartmentName: "Business & Research", DisplayName: "Maya Manager", Name: "Maya Manager", Kind: domain.EmployeeKindAI, State: domain.EmployeeStateActive, Title: "Business Manager", Role: "AI Manager", ManagerID: &h1ID, ManagerName: "Founder", Model: "Model X", Provider: "Provider A", Skills: []string{"Planning", "Delegation", "Team Review"}, MCP: []string{"task-mcp", "budget-mcp"}, Teams: []string{"insights", "revenue"}, KPI: 93, BudgetUSD: 45, SpentUSD: 23.40, Version: 1},
		{ID: "m2", OrganizationID: "org-default", DepartmentID: &deptTech, DepartmentName: "Technology", DisplayName: "Theo CTO", Name: "Theo CTO", Kind: domain.EmployeeKindAI, State: domain.EmployeeStateActive, Title: "Head of Technology", Role: "Tech Lead", ManagerID: &h1ID, ManagerName: "Founder", Model: "Model Z", Provider: "Provider B", Skills: []string{"Architecture", "Code Review"}, MCP: []string{"git-mcp", "task-mcp"}, Teams: []string{"platform", "quality"}, KPI: 91, BudgetUSD: 75, SpentUSD: 42.30, Version: 1},
		{ID: "r1", OrganizationID: "org-default", DepartmentID: &deptBiz, DepartmentName: "Business & Research", DisplayName: "Nora Research", Name: "Nora Research", Kind: domain.EmployeeKindAI, State: domain.EmployeeStateActive, Title: "Research Analyst", Role: "Researcher", ManagerID: &m1ID, ManagerName: "Maya Manager", Model: "Model Y", Provider: "Provider A", Skills: []string{"Market Research", "Source Verification"}, MCP: []string{"web-mcp", "docs-mcp"}, Teams: []string{"insights"}, KPI: 96, BudgetUSD: 20, SpentUSD: 12.82, Version: 1},
		{ID: "mk1", OrganizationID: "org-default", DepartmentID: &deptBiz, DepartmentName: "Business & Research", DisplayName: "Liam Growth", Name: "Liam Growth", Kind: domain.EmployeeKindAI, State: domain.EmployeeStateActive, Title: "Marketing Specialist", Role: "Marketing", ManagerID: &m1ID, ManagerName: "Maya Manager", Model: "Model Y", Provider: "Provider A", Skills: []string{"Copywriting", "SEO"}, MCP: []string{"web-mcp", "docs-mcp"}, Teams: []string{"revenue"}, KPI: 90, BudgetUSD: 20, SpentUSD: 9.43, Version: 1},
		{ID: "f1", OrganizationID: "org-default", DepartmentID: &deptOps, DepartmentName: "Operations", DisplayName: "Iris Finance", Name: "Iris Finance", Kind: domain.EmployeeKindAI, State: domain.EmployeeStateActive, Title: "Cost Analyst", Role: "Finance", ManagerID: &h1ID, ManagerName: "Founder", Model: "Local 9B", Provider: "Ollama", Skills: []string{"Reconciliation", "Budget Forecast"}, MCP: []string{"budget-mcp"}, Teams: []string{"finance"}, KPI: 94, BudgetUSD: 8, SpentUSD: 2.61, Version: 1},
		{ID: "b1", OrganizationID: "org-default", DepartmentID: &deptTech, DepartmentName: "Technology", DisplayName: "Kai Backend", Name: "Kai Backend", Kind: domain.EmployeeKindAI, State: domain.EmployeeStateActive, Title: "Backend Engineer", Role: "Java Engineer", ManagerID: &m2ID, ManagerName: "Theo CTO", Model: "Model Z", Provider: "Provider B", Skills: []string{"Spring Boot", "SQL", "Unit Testing"}, MCP: []string{"git-mcp", "shell-mcp"}, Teams: []string{"platform"}, KPI: 92, BudgetUSD: 35, SpentUSD: 19.35, Version: 1},
		{ID: "b2", OrganizationID: "org-default", DepartmentID: &deptTech, DepartmentName: "Technology", DisplayName: "Ava Frontend", Name: "Ava Frontend", Kind: domain.EmployeeKindAI, State: domain.EmployeeStateActive, Title: "Frontend Engineer", Role: "Angular Engineer", ManagerID: &m2ID, ManagerName: "Theo CTO", Model: "Model X", Provider: "Provider A", Skills: []string{"Angular", "UX", "Accessibility"}, MCP: []string{"git-mcp", "browser-mcp"}, Teams: []string{"platform"}, KPI: 95, BudgetUSD: 30, SpentUSD: 17.20, Version: 1},
		{ID: "q1", OrganizationID: "org-default", DepartmentID: &deptTech, DepartmentName: "Technology", DisplayName: "Zed Quality", Name: "Zed Quality", Kind: domain.EmployeeKindAI, State: domain.EmployeeStateActive, Title: "QA Engineer", Role: "QA Automation", ManagerID: &m2ID, ManagerName: "Theo CTO", Model: "Model Y", Provider: "Provider A", Skills: []string{"Playwright", "Regression", "Test Design"}, MCP: []string{"git-mcp", "test-mcp"}, Teams: []string{"quality"}, KPI: 97, BudgetUSD: 18, SpentUSD: 8.15, Version: 1},
	}
	for _, e := range employees {
		r.employees[e.ID] = e
	}

	projects := []domain.Project{
		{
			ID: "p1", OrganizationID: "org-default", Code: "ABOS-CORE", Name: "AI Business OS Platform",
			Description: "Xây lõi tổ chức, quản trị agent, task, budget và web console.",
			AccountableOwnerID: "m2", OwnerName: "Theo CTO", Status: "ACTIVE", BudgetUSD: 130, SpentUSD: 78.9, DueAt: "2026-12-15",
			Members: []string{"m2", "b1", "b2", "q1", "m1"}, Teams: []string{"platform", "quality", "insights"},
			Milestones: []domain.Milestone{
				{Name: "Domain & API contract", Done: true},
				{Name: "Employee + task vertical slice", Done: true},
				{Name: "Budget + MCP authorization", Done: false},
				{Name: "Acceptance and observability", Done: false},
			},
			Version: 1,
		},
		{
			ID: "p2", OrganizationID: "org-default", Code: "MARKET-26", Name: "AI Market Intelligence",
			Description: "Thử nghiệm team AI nghiên cứu thị trường và tổng hợp báo cáo đa lĩnh vực.",
			AccountableOwnerID: "m1", OwnerName: "Maya Manager", Status: "ACTIVE", BudgetUSD: 55, SpentUSD: 27.4, DueAt: "2026-11-30",
			Members: []string{"m1", "r1", "mk1"}, Teams: []string{"insights", "revenue"},
			Milestones: []domain.Milestone{
				{Name: "Research protocol", Done: true},
				{Name: "Weekly monitoring", Done: false},
				{Name: "Executive report", Done: false},
			},
			Version: 1,
		},
		{
			ID: "p3", OrganizationID: "org-default", Code: "COST-OPT", Name: "Model Cost Optimization",
			Description: "Benchmark đa provider, tối ưu cost per accepted task, approval và rollback.",
			AccountableOwnerID: "f1", OwnerName: "Iris Finance", Status: "PLANNED", BudgetUSD: 30, SpentUSD: 11.2, DueAt: "2026-12-30",
			Members: []string{"f1", "r1", "m2", "q1"}, Teams: []string{"finance", "insights", "quality"},
			Milestones: []domain.Milestone{
				{Name: "Evaluation dataset", Done: true},
				{Name: "Provider benchmark", Done: false},
				{Name: "Canary dashboard", Done: false},
			},
			Version: 1,
		},
	}
	for _, p := range projects {
		r.projects[p.ID] = p
	}

	p1 := "p1"
	p2 := "p2"
	p3 := "p3"

	tasks := []domain.Task{
		{ID: "T-101", Title: "Thiết kế Employee / Role domain", ProjectID: &p1, ProjectCode: "ABOS-CORE", AccountableOwnerID: "m2", OwnerName: "Theo CTO", Collaborators: []string{"b1"}, Status: domain.TaskStateAccepted, Priority: "HIGH", AcceptanceCriteria: "Entity rõ và có tests", DueAt: "2026-10-30", CostUSD: 2.6, ExecutionsCount: 2, Version: 1},
		{ID: "T-102", Title: "Build organization chart API", ProjectID: &p1, ProjectCode: "ABOS-CORE", AccountableOwnerID: "b1", OwnerName: "Kai Backend", Collaborators: []string{"b2"}, Status: domain.TaskStateInProgress, Priority: "HIGH", AcceptanceCriteria: "API permission-filtered trả node/edge", DueAt: "2026-11-02", CostUSD: 4.3, ExecutionsCount: 3, Version: 1},
		{ID: "T-103", Title: "Triển khai Angular agent detail drawer", ProjectID: &p1, ProjectCode: "ABOS-CORE", AccountableOwnerID: "b2", OwnerName: "Ava Frontend", Collaborators: []string{}, Status: domain.TaskStateInReview, Priority: "HIGH", AcceptanceCriteria: "Profile với model MCP skills tasks KPI", DueAt: "2026-11-03", CostUSD: 3.2, ExecutionsCount: 2, Version: 1},
		{ID: "T-104", Title: "Test phân quyền Tool Gateway", ProjectID: &p1, ProjectCode: "ABOS-CORE", AccountableOwnerID: "q1", OwnerName: "Zed Quality", Collaborators: []string{"b1"}, Status: domain.TaskStateReady, Priority: "CRITICAL", AcceptanceCriteria: "Permission denied / audit tests", DueAt: "2026-11-06", CostUSD: 0.4, ExecutionsCount: 0, Version: 1},
		{ID: "T-105", Title: "Budget reservation concurrency", ProjectID: &p1, ProjectCode: "ABOS-CORE", AccountableOwnerID: "b1", OwnerName: "Kai Backend", Collaborators: []string{"f1"}, Status: domain.TaskStateBlocked, Priority: "HIGH", AcceptanceCriteria: "No overspend under concurrent calls", DueAt: "2026-11-05", CostUSD: 5.7, ExecutionsCount: 4, Version: 1},
		{ID: "T-106", Title: "CI/CD Compose environment", ProjectID: &p1, ProjectCode: "ABOS-CORE", AccountableOwnerID: "m2", OwnerName: "Theo CTO", Collaborators: []string{"b1"}, Status: domain.TaskStateCompleted, Priority: "MEDIUM", AcceptanceCriteria: "Repeatable staging setup", DueAt: "2026-10-29", CostUSD: 1.8, ExecutionsCount: 2, Version: 1},
		{ID: "T-107", Title: "Market research competitor shortlist", ProjectID: &p2, ProjectCode: "MARKET-26", AccountableOwnerID: "r1", OwnerName: "Nora Research", Collaborators: []string{"mk1"}, Status: domain.TaskStateInProgress, Priority: "HIGH", AcceptanceCriteria: "10 verified competitors + sources", DueAt: "2026-11-02", CostUSD: 2.4, ExecutionsCount: 3, Version: 1},
		{ID: "T-108", Title: "Weekly social trends report", ProjectID: &p2, ProjectCode: "MARKET-26", AccountableOwnerID: "r1", OwnerName: "Nora Research", Collaborators: []string{}, Status: domain.TaskStateBacklog, Priority: "MEDIUM", AcceptanceCriteria: "Report with confidence label", DueAt: "2026-11-10", CostUSD: 0, ExecutionsCount: 0, Version: 1},
		{ID: "T-109", Title: "Landing page positioning draft", ProjectID: &p2, ProjectCode: "MARKET-26", AccountableOwnerID: "mk1", OwnerName: "Liam Growth", Collaborators: []string{"m1"}, Status: domain.TaskStateInReview, Priority: "MEDIUM", AcceptanceCriteria: "2 tested messages and targeting rationale", DueAt: "2026-11-05", CostUSD: 1.25, ExecutionsCount: 1, Version: 1},
		{ID: "T-110", Title: "Research approval for external publishing", ProjectID: &p2, ProjectCode: "MARKET-26", AccountableOwnerID: "m1", OwnerName: "Maya Manager", Collaborators: []string{}, Status: domain.TaskStateBlocked, Priority: "HIGH", AcceptanceCriteria: "Human approval record and publication evidence", DueAt: "2026-11-04", CostUSD: 0.3, ExecutionsCount: 1, Version: 1},
		{ID: "T-111", Title: "Build benchmark golden dataset", ProjectID: &p3, ProjectCode: "COST-OPT", AccountableOwnerID: "q1", OwnerName: "Zed Quality", Collaborators: []string{"r1"}, Status: domain.TaskStateAccepted, Priority: "HIGH", AcceptanceCriteria: "Versioned dataset with rubrics", DueAt: "2026-10-31", CostUSD: 1.65, ExecutionsCount: 1, Version: 1},
		{ID: "T-112", Title: "Provider price and quality analysis", ProjectID: &p3, ProjectCode: "COST-OPT", AccountableOwnerID: "f1", OwnerName: "Iris Finance", Collaborators: []string{"m2"}, Status: domain.TaskStateReady, Priority: "HIGH", AcceptanceCriteria: "Quality threshold and cost comparison", DueAt: "2026-11-06", CostUSD: 0.5, ExecutionsCount: 0, Version: 1},
		{ID: "T-113", Title: "Model swap approval workflow", ProjectID: &p3, ProjectCode: "COST-OPT", AccountableOwnerID: "m2", OwnerName: "Theo CTO", Collaborators: []string{"f1"}, Status: domain.TaskStateBacklog, Priority: "MEDIUM", AcceptanceCriteria: "Audit approved change only", DueAt: "2026-11-14", CostUSD: 0, ExecutionsCount: 0, Version: 1},
		{ID: "T-114", Title: "Review internal operating expenses", ProjectID: nil, ProjectCode: "Operations", AccountableOwnerID: "f1", OwnerName: "Iris Finance", Collaborators: []string{}, Status: domain.TaskStateReady, Priority: "LOW", AcceptanceCriteria: "Reconciled expenses report", DueAt: "2026-11-08", CostUSD: 0.31, ExecutionsCount: 0, Version: 1},
		{ID: "T-115", Title: "Executive weekly progress report", ProjectID: nil, ProjectCode: "Operations", AccountableOwnerID: "m1", OwnerName: "Maya Manager", Collaborators: []string{"m2"}, Status: domain.TaskStateInProgress, Priority: "MEDIUM", AcceptanceCriteria: "All projects summarised without double-count", DueAt: "2026-10-30", CostUSD: 0.7, ExecutionsCount: 1, Version: 1},
	}
	for _, t := range tasks {
		t.OrganizationID = "org-default"
		t.CreatedAt = time.Now().Add(-48 * time.Hour)
		t.UpdatedAt = time.Now()
		r.tasks[t.ID] = t
	}

	approvals := []domain.ApprovalRequest{
		{ID: "APR-12", OrganizationID: "org-default", TaskID: &tasks[9].ID, TaskTitle: tasks[9].Title, RequestedByEmployeeID: "m1", RequesterName: "Maya Manager", ActionKind: "External write via docs-mcp", RiskTier: "HIGH", Status: domain.ApprovalStatePending, Title: "Publish external research report", CreatedAt: time.Now().Add(-12 * time.Hour)},
		{ID: "APR-13", OrganizationID: "org-default", RequestedByEmployeeID: "m1", RequesterName: "Maya Manager", ActionKind: "Model assignment change", RiskTier: "MEDIUM", Status: domain.ApprovalStatePending, Title: "Change primary model for Nora Research", CreatedAt: time.Now().Add(-8 * time.Hour)},
		{ID: "APR-14", OrganizationID: "org-default", TaskID: &tasks[4].ID, TaskTitle: tasks[4].Title, RequestedByEmployeeID: "m2", RequesterName: "Theo CTO", ActionKind: "Budget increase", RiskTier: "HIGH", Status: domain.ApprovalStatePending, Title: "Increase project budget ABOS-CORE +$20", CreatedAt: time.Now().Add(-4 * time.Hour)},
	}
	for _, a := range approvals {
		r.approvals[a.ID] = a
	}

	mcpServers := []domain.MCPServer{
		{ID: "mcp-1", OrganizationID: "org-default", Name: "web-mcp", Transport: "stdio", Description: "Search, crawl and retrieve external web content", RiskTier: "R1", Enabled: true, ReviewStatus: "APPROVED", CreatedAt: time.Now()},
		{ID: "mcp-2", OrganizationID: "org-default", Name: "git-mcp", Transport: "stdio", Description: "Read repo, commit changes, branch and PR management", RiskTier: "R2", Enabled: true, ReviewStatus: "APPROVED", CreatedAt: time.Now()},
		{ID: "mcp-3", OrganizationID: "org-default", Name: "shell-mcp", Transport: "docker-sandbox", Description: "Isolated shell command runner within policy", RiskTier: "R3", Enabled: true, ReviewStatus: "APPROVED", CreatedAt: time.Now()},
		{ID: "mcp-4", OrganizationID: "org-default", Name: "budget-mcp", Transport: "stdio", Description: "Ledger query, spend forecasts and ratecard inspect", RiskTier: "R1", Enabled: true, ReviewStatus: "APPROVED", CreatedAt: time.Now()},
		{ID: "mcp-5", OrganizationID: "org-default", Name: "docs-mcp", Transport: "stdio", Description: "Read and write enterprise documentation repository", RiskTier: "R2", Enabled: true, ReviewStatus: "APPROVED", CreatedAt: time.Now()},
		{ID: "mcp-6", OrganizationID: "org-default", Name: "task-mcp", Transport: "internal", Description: "Query, delegate and report task lifecycle status", RiskTier: "R2", Enabled: true, ReviewStatus: "APPROVED", CreatedAt: time.Now()},
	}
	for _, m := range mcpServers {
		r.mcpServers[m.ID] = m
	}

	models := []domain.ModelInfo{
		{ID: "mv-1", ProviderCode: "Provider A", ModelID: "Model X", Revision: "2026-q3", Features: []string{"Tool Calling", "JSON Mode", "Fast Latency"}, Status: "ACTIVE"},
		{ID: "mv-2", ProviderCode: "Provider A", ModelID: "Model Y", Revision: "2026-q3", Features: []string{"Long Context", "Reasoning", "Tool Calling"}, Status: "ACTIVE"},
		{ID: "mv-3", ProviderCode: "Provider B", ModelID: "Model Z", Revision: "2026-q4", Features: []string{"Parallel Tools", "High Coding Benchmark"}, Status: "ACTIVE"},
		{ID: "mv-4", ProviderCode: "Ollama", ModelID: "Local 9B", Revision: "2026-v2", Features: []string{"Local Privacy", "Low Cost", "Basic Tools"}, Status: "ACTIVE"},
	}
	for _, m := range models {
		r.models[m.ID] = m
	}
}

func (r *MemoryRepo) GetDepartments(ctx context.Context) ([]domain.Department, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]domain.Department, 0, len(r.departments))
	for _, d := range r.departments {
		res = append(res, d)
	}
	return res, nil
}

func (r *MemoryRepo) GetTeams(ctx context.Context) ([]domain.Team, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]domain.Team, 0, len(r.teams))
	for _, t := range r.teams {
		res = append(res, t)
	}
	return res, nil
}

func (r *MemoryRepo) GetPositions(ctx context.Context) ([]domain.Position, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]domain.Position, 0, len(r.positions))
	for _, p := range r.positions {
		res = append(res, p)
	}
	return res, nil
}

func (r *MemoryRepo) GetEmployees(ctx context.Context) ([]domain.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]domain.Employee, 0, len(r.employees))
	for _, e := range r.employees {
		res = append(res, e)
	}
	return res, nil
}

func (r *MemoryRepo) GetEmployeeByID(ctx context.Context, id string) (*domain.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.employees[id]
	if !ok {
		return nil, fmt.Errorf("employee %s not found", id)
	}
	return &e, nil
}

func (r *MemoryRepo) CreateEmployee(ctx context.Context, emp *domain.Employee) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	emp.Version = 1
	emp.CreatedAt = time.Now()
	r.employees[emp.ID] = *emp
	return nil
}

func (r *MemoryRepo) UpdateEmployee(ctx context.Context, emp *domain.Employee) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.employees[emp.ID]
	if !ok {
		return fmt.Errorf("employee %s not found", emp.ID)
	}
	emp.Version = existing.Version + 1
	r.employees[emp.ID] = *emp
	return nil
}

func (r *MemoryRepo) GetOrgTree(ctx context.Context) (*domain.OrgNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var root *domain.Employee
	for _, e := range r.employees {
		if e.ManagerID == nil {
			copy := e
			root = &copy
			break
		}
	}
	if root == nil {
		return nil, fmt.Errorf("root employee not found")
	}

	var buildNode func(e *domain.Employee) *domain.OrgNode
	buildNode = func(e *domain.Employee) *domain.OrgNode {
		node := &domain.OrgNode{
			ID:         e.ID,
			Name:       e.Name,
			Title:      e.Title,
			Role:       e.Role,
			Kind:       string(e.Kind),
			Model:      e.Model,
			Provider:   e.Provider,
			Department: e.DepartmentName,
			DirectSub:  make([]*domain.OrgNode, 0),
		}
		for _, child := range r.employees {
			if child.ManagerID != nil && *child.ManagerID == e.ID {
				childCopy := child
				node.DirectSub = append(node.DirectSub, buildNode(&childCopy))
			}
		}
		return node
	}

	return buildNode(root), nil
}

func (r *MemoryRepo) GetProjects(ctx context.Context) ([]domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]domain.Project, 0, len(r.projects))
	for _, p := range r.projects {
		res = append(res, p)
	}
	return res, nil
}

func (r *MemoryRepo) GetProjectByID(ctx context.Context, id string) (*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.projects[id]
	if !ok {
		return nil, fmt.Errorf("project %s not found", id)
	}
	return &p, nil
}

func (r *MemoryRepo) CreateProject(ctx context.Context, p *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.Version = 1
	p.CreatedAt = time.Now()
	r.projects[p.ID] = *p
	return nil
}

func (r *MemoryRepo) UpdateProject(ctx context.Context, p *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.projects[p.ID]
	if !ok {
		return fmt.Errorf("project %s not found", p.ID)
	}
	p.Version = existing.Version + 1
	r.projects[p.ID] = *p
	return nil
}

func (r *MemoryRepo) inScope(t *domain.Task, scope string) bool {
	if scope == "" || scope == "company:all" {
		return true
	}
	parts := strings.Split(scope, ":")
	if len(parts) != 2 {
		return true
	}
	scopeType, scopeID := parts[0], parts[1]
	related := append([]string{t.AccountableOwnerID}, t.Collaborators...)

	switch scopeType {
	case "company":
		return true
	case "project":
		return t.ProjectID != nil && *t.ProjectID == scopeID
	case "department":
		for _, id := range related {
			if emp, ok := r.employees[id]; ok && emp.DepartmentID != nil && *emp.DepartmentID == scopeID {
				return true
			}
		}
	case "team":
		for _, id := range related {
			if emp, ok := r.employees[id]; ok {
				for _, tm := range emp.Teams {
					if tm == scopeID {
						return true
					}
				}
			}
		}
	case "employee":
		for _, id := range related {
			if id == scopeID {
				return true
			}
		}
	case "manager":
		subordinates := r.reportingSubtree(scopeID)
		for _, id := range related {
			if subordinates[id] {
				return true
			}
		}
	case "role":
		for _, id := range related {
			if emp, ok := r.employees[id]; ok && emp.Role == scopeID {
				return true
			}
		}
	}
	return false
}

func (r *MemoryRepo) reportingSubtree(rootID string) map[string]bool {
	sub := make(map[string]bool)
	sub[rootID] = true
	changed := true
	for changed {
		changed = false
		for _, e := range r.employees {
			if e.ManagerID != nil && sub[*e.ManagerID] && !sub[e.ID] {
				sub[e.ID] = true
				changed = true
			}
		}
	}
	return sub
}

func (r *MemoryRepo) GetTasks(ctx context.Context, scope, projectFilter, ownerFilter, statusFilter, search string) ([]domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []domain.Task
	searchLower := strings.ToLower(search)

	for _, t := range r.tasks {
		if !r.inScope(&t, scope) {
			continue
		}
		if projectFilter != "" && (t.ProjectID == nil || *t.ProjectID != projectFilter) {
			continue
		}
		if ownerFilter != "" {
			matched := t.AccountableOwnerID == ownerFilter
			if !matched {
				for _, c := range t.Collaborators {
					if c == ownerFilter {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
		}
		if statusFilter != "" && string(t.Status) != statusFilter {
			continue
		}
		if searchLower != "" {
			txt := strings.ToLower(fmt.Sprintf("%s %s %s %s", t.ID, t.Title, t.ProjectCode, t.OwnerName))
			if !strings.Contains(txt, searchLower) {
				continue
			}
		}
		res = append(res, t)
	}
	return res, nil
}

func (r *MemoryRepo) GetTaskByID(ctx context.Context, id string) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task %s not found", id)
	}
	return &t, nil
}

func (r *MemoryRepo) CreateTask(ctx context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if task.ID == "" {
		task.ID = fmt.Sprintf("T-%s", strings.ToUpper(uuid.New().String()[:8]))
	}
	if task.OrganizationID == "" {
		task.OrganizationID = "org-default"
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	task.UpdatedAt = now
	if task.Version == 0 {
		task.Version = 1
	}
	if task.AccountableOwnerID != "" {
		if emp, ok := r.employees[task.AccountableOwnerID]; ok {
			task.OwnerName = emp.Name
		}
	}
	if task.ProjectID != nil && *task.ProjectID != "" {
		if p, ok := r.projects[*task.ProjectID]; ok {
			task.ProjectCode = p.Code
			task.ProjectName = p.Name
		}
	} else {
		task.ProjectCode = "Operations"
	}
	task.Version = 1
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = *task
	return nil
}

func (r *MemoryRepo) UpdateTask(ctx context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.tasks[task.ID]
	if !ok {
		return fmt.Errorf("task %s not found", task.ID)
	}
	task.Version = existing.Version + 1
	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = *task
	return nil
}

func (r *MemoryRepo) GetExecutionsByTaskID(ctx context.Context, taskID string) ([]domain.Execution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []domain.Execution
	for _, e := range r.executions {
		if e.TaskID == taskID {
			res = append(res, e)
		}
	}
	return res, nil
}

func (r *MemoryRepo) CreateExecution(ctx context.Context, exec *domain.Execution) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executions[exec.ID] = *exec
	return nil
}

func (r *MemoryRepo) UpdateExecution(ctx context.Context, exec *domain.Execution) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executions[exec.ID] = *exec
	return nil
}

func (r *MemoryRepo) GetApprovals(ctx context.Context) ([]domain.ApprovalRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]domain.ApprovalRequest, 0, len(r.approvals))
	for _, a := range r.approvals {
		res = append(res, a)
	}
	return res, nil
}

func (r *MemoryRepo) GetApprovalByID(ctx context.Context, id string) (*domain.ApprovalRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.approvals[id]
	if !ok {
		return nil, fmt.Errorf("approval %s not found", id)
	}
	return &a, nil
}

func (r *MemoryRepo) UpdateApproval(ctx context.Context, approval *domain.ApprovalRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.approvals[approval.ID] = *approval
	return nil
}

func (r *MemoryRepo) GetMCPServers(ctx context.Context) ([]domain.MCPServer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]domain.MCPServer, 0, len(r.mcpServers))
	for _, m := range r.mcpServers {
		res = append(res, m)
	}
	return res, nil
}

func (r *MemoryRepo) GetModels(ctx context.Context) ([]domain.ModelInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]domain.ModelInfo, 0, len(r.models))
	for _, m := range r.models {
		res = append(res, m)
	}
	return res, nil
}

func (r *MemoryRepo) GetBoardSummary(ctx context.Context, scope string) (*domain.BoardSummary, error) {
	tasks, err := r.GetTasks(ctx, scope, "", "", "", "")
	if err != nil {
		return nil, err
	}
	summary := &domain.BoardSummary{
		TotalTasks: len(tasks),
	}
	projects := make(map[string]bool)
	for _, t := range tasks {
		if t.Status == domain.TaskStateAccepted {
			summary.AcceptedTasks++
		}
		if t.Status == domain.TaskStateInProgress || t.Status == domain.TaskStateInReview {
			summary.ActiveTasks++
		}
		if t.Status == domain.TaskStateBlocked {
			summary.BlockedTasks++
		}
		if t.Status == domain.TaskStateBacklog {
			summary.BacklogTasks++
		}
		summary.TotalSpentUSD += t.CostUSD
		if t.ProjectID != nil && *t.ProjectID != "" {
			projects[*t.ProjectID] = true
		}
	}
	summary.ProjectsCount = len(projects)
	return summary, nil
}
