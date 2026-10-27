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
