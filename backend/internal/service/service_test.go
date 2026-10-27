package service_test

import (
	"context"
	"testing"

	"github.com/cuongbuilh/ai-business-os/backend/internal/domain"
	"github.com/cuongbuilh/ai-business-os/backend/internal/repository"
	"github.com/cuongbuilh/ai-business-os/backend/internal/service"
)

func setupTestServices() *service.Services {
	repo := repository.NewMemoryRepo()
	broker := service.NewEventBroker()
	return service.NewServices(repo, broker)
}

func TestOrgChartAndHierarchy(t *testing.T) {
	svc := setupTestServices()
	ctx := context.Background()

	chart, err := svc.GetOrgChart(ctx)
	if err != nil {
		t.Fatalf("failed to get org chart: %v", err)
	}

	if chart.Name != "Founder" {
		t.Errorf("expected root founder, got %s", chart.Name)
	}

	if len(chart.DirectSub) == 0 {
		t.Errorf("expected direct reports under Founder")
	}
}

func TestBR_ORG_002_SelfReportingPrevented(t *testing.T) {
	svc := setupTestServices()
	ctx := context.Background()

	selfID := "emp-test"
	emp := &domain.Employee{
		ID:          selfID,
		Name:        "Self Manager",
		DisplayName: "Self Manager",
		ManagerID:   &selfID,
	}

	err := svc.CreateEmployee(ctx, emp)
	if err == nil {
		t.Fatalf("expected error for self-reporting employee, got nil")
	}
}

func TestBR_WORK_002_TaskAccountableOwnerRequired(t *testing.T) {
	svc := setupTestServices()
	ctx := context.Background()

	task := &domain.Task{
		Title:              "Unassigned Task",
		AccountableOwnerID: "",
	}

	err := svc.CreateTask(ctx, task)
	if err == nil {
		t.Fatalf("expected error when accountable owner is missing")
	}
}

func TestBR_AUTH_001_NoSelfApproval(t *testing.T) {
	svc := setupTestServices()
	ctx := context.Background()

	// APR-12 was requested by m1
	err := svc.DecideApproval(ctx, "APR-12", "APPROVE", "m1")
	if err == nil {
		t.Fatalf("expected error for self-approval, got nil")
	}

	// Approving by Founder (h1) should succeed
	err = svc.DecideApproval(ctx, "APR-12", "APPROVE", "h1")
	if err != nil {
		t.Fatalf("unexpected error when Founder approves: %v", err)
	}
}

func TestTaskAcceptance(t *testing.T) {
	svc := setupTestServices()
	ctx := context.Background()

	// BR-WORK-003: Acceptance from in_progress should be rejected
	err := svc.AcceptTask(ctx, "T-102", "premature acceptance")
	if err == nil {
		t.Fatalf("expected error when accepting in_progress task, got nil")
	}

	// T-103 is in in_review, so accepting it should succeed
	err = svc.AcceptTask(ctx, "T-103", "Looks good, tests pass")
	if err != nil {
		t.Fatalf("failed to accept task in_review: %v", err)
	}

	task, err := svc.GetTaskDetail(ctx, "T-103")
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if task.Status != domain.TaskStateAccepted {
		t.Errorf("expected status accepted, got %s", task.Status)
	}
	if task.AcceptedAt == nil {
		t.Errorf("expected acceptedAt timestamp to be set")
	}
}

func TestTaskExecutionRun(t *testing.T) {
	svc := setupTestServices()
	ctx := context.Background()

	exec, err := svc.ExecuteTaskRun(ctx, "T-104")
	if err != nil {
		t.Fatalf("failed to execute task run: %v", err)
	}

	if exec.Status != domain.ExecutionStateSucceeded {
		t.Errorf("expected succeeded status, got %s", exec.Status)
	}

	if len(exec.Steps) == 0 {
		t.Errorf("expected execution steps to be generated")
	}

	if exec.TotalCostUSD <= 0 {
		t.Errorf("expected cost to be recorded, got %f", exec.TotalCostUSD)
	}

	task, _ := svc.GetTaskDetail(ctx, "T-104")
	if task.Status != domain.TaskStateInReview {
		t.Errorf("expected task status to move to in_review, got %s", task.Status)
	}
}

func TestScopeFiltering(t *testing.T) {
	svc := setupTestServices()
	ctx := context.Background()

	// Company scope: all tasks
	tasksAll, err := svc.GetTasks(ctx, "company:all", "", "", "", "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(tasksAll) == 0 {
		t.Errorf("expected non-empty company tasks")
	}

	// Project scope
	tasksP1, err := svc.GetTasks(ctx, "project:p1", "", "", "", "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	for _, tsk := range tasksP1 {
		if tsk.ProjectID == nil || *tsk.ProjectID != "p1" {
			t.Errorf("expected project p1, got %v", tsk.ProjectID)
		}
	}
}

func TestUpdateReportingLine_CycleDetection(t *testing.T) {
	svc := setupTestServices()
	ctx := context.Background()

	// h1 is Founder (manager = nil)
	// m1 reports to h1
	// r1 reports to m1
	// Attempt 1: Setting m1 to report to itself should fail BR-ORG-002
	m1ID := "m1"
	err := svc.UpdateReportingLine(ctx, "m1", &m1ID)
	if err == nil {
		t.Fatalf("expected error for self reporting, got nil")
	}

	// Attempt 2: Setting m1 to report to r1 (its subordinate) creates cycle r1 -> m1 -> r1!
	subID := "r1"
	err = svc.UpdateReportingLine(ctx, "m1", &subID)
	if err == nil {
		t.Fatalf("expected cycle detection error for m1 -> r1, got nil")
	}

	// Attempt 3: Setting r1 to report to m2 (valid peer manager) should succeed
	m2ID := "m2"
	err = svc.UpdateReportingLine(ctx, "r1", &m2ID)
	if err != nil {
		t.Fatalf("expected valid reassignment to succeed, got: %v", err)
	}

	emp, _ := svc.GetEmployeeDetail(ctx, "r1")
	if emp.ManagerID == nil || *emp.ManagerID != "m2" {
		t.Errorf("expected manager to be m2, got %v", emp.ManagerID)
	}
}

func TestCapabilitiesAndEntitiesCRUD(t *testing.T) {
	svc := setupTestServices()
	ctx := context.Background()

	// 1. Model CRUD
	m := &domain.ModelInfo{
		ID:           "model-custom-1",
		ProviderCode: "Provider A",
		ModelID:      "model-custom",
		Revision:     "v1",
		Features:     []string{"Tool Calling", "Vision"},
		Status:       "ACTIVE",
	}
	if err := svc.CreateModel(ctx, m); err != nil {
		t.Fatalf("failed to create model: %v", err)
	}
	m.Status = "DEPRECATED"
	if err := svc.UpdateModel(ctx, m); err != nil {
		t.Fatalf("failed to update model: %v", err)
	}
	if err := svc.DeleteModel(ctx, m.ID); err != nil {
		t.Fatalf("failed to delete model: %v", err)
	}

	// 2. MCP Server CRUD
	server := &domain.MCPServer{
		ID:          "mcp-custom-test",
		Name:        "custom-test-mcp",
		Transport:   "stdio",
		Description: "Testing MCP",
		RiskTier:    "R2",
		Enabled:     true,
	}
	if err := svc.CreateMCPServer(ctx, server); err != nil {
		t.Fatalf("failed to create mcp server: %v", err)
	}
	server.Description = "Updated description"
	if err := svc.UpdateMCPServer(ctx, server); err != nil {
		t.Fatalf("failed to update mcp server: %v", err)
	}
	if err := svc.DeleteMCPServer(ctx, server.ID); err != nil {
		t.Fatalf("failed to delete mcp server: %v", err)
	}

	// 3. Department & Team CRUD
	dept := &domain.Department{
		ID:   "dept-test",
		Name: "Test Department",
	}
	if err := svc.CreateDepartment(ctx, dept); err != nil {
		t.Fatalf("failed to create dept: %v", err)
	}
	dept.Name = "Test Department Updated"
	if err := svc.UpdateDepartment(ctx, dept); err != nil {
		t.Fatalf("failed to update dept: %v", err)
	}

	team := &domain.Team{
		ID:           "team-test",
		DepartmentID: dept.ID,
		Name:         "Test Team",
	}
	if err := svc.CreateTeam(ctx, team); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	if err := svc.DeleteTeam(ctx, team.ID); err != nil {
		t.Fatalf("failed to delete team: %v", err)
	}
	if err := svc.DeleteDepartment(ctx, dept.ID); err != nil {
		t.Fatalf("failed to delete dept: %v", err)
	}
}
