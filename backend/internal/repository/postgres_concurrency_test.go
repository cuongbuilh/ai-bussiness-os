package repository_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cuongbuilh/ai-business-os/backend/internal/domain"
	"github.com/cuongbuilh/ai-business-os/backend/internal/repository"
	"github.com/google/uuid"
)

func TestPostgres_MultiInstance_AtomicBudgetConcurrency(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://abos:abos_secret@localhost:5432/abos_qa_20261027?sslmode=disable"
	}

	// Instance 1
	repo1, err := repository.NewPostgresRepo(dbURL)
	if err != nil {
		t.Skipf("Skipping live postgres concurrency test: %v", err)
		return
	}

	// Instance 2 (representing a completely separate backend process)
	repo2, err := repository.NewPostgresRepo(dbURL)
	if err != nil {
		t.Skipf("Skipping live postgres concurrency test: %v", err)
		return
	}

	ctx := context.Background()
	testEmpID := fmt.Sprintf("qa-multi-%s", uuid.New().String()[:8])
	managerID := "h1"

	// Create an employee with budget cap = $0.50
	emp := &domain.Employee{
		ID:             testEmpID,
		OrganizationID: "org-default",
		DisplayName:    "QA Multi-Instance Agent",
		Role:           "QA Concurrency Tester",
		Kind:           "AI",
		ManagerID:      &managerID,
		BudgetUSD:      0.50,
		SpentUSD:       0.0,
		State:          domain.EmployeeStateActive,
	}
	if err := repo1.CreateEmployee(ctx, emp); err != nil {
		t.Fatalf("failed to create test employee: %v", err)
	}

	// Create test task
	taskID := fmt.Sprintf("T-QA-%s", uuid.New().String()[:6])
	task := &domain.Task{
		ID:                 taskID,
		OrganizationID:     "org-default",
		Title:              "QA Multi Instance Task",
		AccountableOwnerID: testEmpID,
		Status:             domain.TaskStateReady,
		Priority:           "HIGH",
	}
	if err := repo1.CreateTask(ctx, task); err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}

	// Create executions for instance 1 and instance 2
	exec1ID := fmt.Sprintf("exec-inst1-%s", uuid.New().String()[:6])
	exec2ID := fmt.Sprintf("exec-inst2-%s", uuid.New().String()[:6])
	now := time.Now()

	exec1 := &domain.Execution{
		ID:                  exec1ID,
		OrganizationID:      "org-default",
		TaskID:              taskID,
		ExecutingEmployeeID: testEmpID,
		ExecutingEmpName:    emp.DisplayName,
		Status:              domain.ExecutionStateRunning,
		AttemptNo:           1,
		IdempotencyKey:      uuid.New().String(),
		CreatedAt:           now,
	}
	if err := repo1.CreateExecution(ctx, exec1); err != nil {
		t.Fatalf("failed to create exec1: %v", err)
	}

	exec2 := &domain.Execution{
		ID:                  exec2ID,
		OrganizationID:      "org-default",
		TaskID:              taskID,
		ExecutingEmployeeID: testEmpID,
		ExecutingEmpName:    emp.DisplayName,
		Status:              domain.ExecutionStateRunning,
		AttemptNo:           2,
		IdempotencyKey:      uuid.New().String(),
		CreatedAt:           now,
	}
	if err := repo2.CreateExecution(ctx, exec2); err != nil {
		t.Fatalf("failed to create exec2: %v", err)
	}

	// Try to reserve $0.37 from Instance 1 and $0.37 from Instance 2 simultaneously
	// Total requested: $0.74, Cap is $0.50.
	// EXACTLY ONE must succeed; the other MUST be rejected!
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	failures := 0

	runReservation := func(repo repository.Repository, execID string) {
		defer wg.Done()
		res := &domain.BudgetReservation{
			ID:             fmt.Sprintf("res-%s", uuid.New().String()[:8]),
			OrganizationID: "org-default",
			ExecutionID:    execID,
			StepKey:        "step_1",
			ReservedUSD:    0.37,
			Status:         "ACTIVE",
			ExpiresAt:      time.Now().Add(5 * time.Minute),
		}
		err := repo.AtomicReserveBudget(ctx, res, testEmpID, 0.37)
		mu.Lock()
		defer mu.Unlock()
		if err == nil {
			successes++
		} else {
			failures++
		}
	}

	wg.Add(2)
	go runReservation(repo1, exec1ID)
	go runReservation(repo2, exec2ID)
	wg.Wait()

	if successes != 1 || failures != 1 {
		t.Fatalf("Multi-instance budget cap failed! Expected 1 success, 1 failure, got %d successes, %d failures", successes, failures)
	}
}

func TestPostgres_MultiInstance_ConcurrentSpendCommit(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://abos:abos_secret@localhost:5432/abos_qa_20261027?sslmode=disable"
	}

	repo1, err := repository.NewPostgresRepo(dbURL)
	if err != nil {
		t.Skipf("Skipping live postgres concurrency test: %v", err)
		return
	}
	repo2, err := repository.NewPostgresRepo(dbURL)
	if err != nil {
		t.Skipf("Skipping live postgres concurrency test: %v", err)
		return
	}

	ctx := context.Background()
	testEmpID := fmt.Sprintf("qa-spend-%s", uuid.New().String()[:8])
	managerID := "h1"

	emp := &domain.Employee{
		ID:             testEmpID,
		OrganizationID: "org-default",
		DisplayName:    "QA Spend Agent",
		Role:           "QA Concurrency Tester",
		Kind:           "AI",
		ManagerID:      &managerID,
		BudgetUSD:      10.0,
		SpentUSD:       0.0,
		State:          domain.EmployeeStateActive,
	}
	if err := repo1.CreateEmployee(ctx, emp); err != nil {
		t.Fatalf("failed to create test employee: %v", err)
	}

	taskID := fmt.Sprintf("T-SP-%s", uuid.New().String()[:6])
	task := &domain.Task{
		ID:                 taskID,
		OrganizationID:     "org-default",
		Title:              "QA Concurrent Spend Task",
		AccountableOwnerID: testEmpID,
		Status:             domain.TaskStateInProgress,
		Priority:           "HIGH",
		CostUSD:            0.0,
		ExecutionsCount:    0,
	}
	if err := repo1.CreateTask(ctx, task); err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}

	// Create 2 executions and 2 reservations
	res1ID := fmt.Sprintf("res-c1-%s", uuid.New().String()[:6])
	res2ID := fmt.Sprintf("res-c2-%s", uuid.New().String()[:6])
	exec1ID := fmt.Sprintf("exec-c1-%s", uuid.New().String()[:6])
	exec2ID := fmt.Sprintf("exec-c2-%s", uuid.New().String()[:6])
	now := time.Now()

	_ = repo1.CreateExecution(ctx, &domain.Execution{
		ID: exec1ID, OrganizationID: "org-default", TaskID: taskID, ExecutingEmployeeID: testEmpID,
		Status: domain.ExecutionStateRunning, AttemptNo: 1, IdempotencyKey: uuid.New().String(), CreatedAt: now,
	})
	_ = repo2.CreateExecution(ctx, &domain.Execution{
		ID: exec2ID, OrganizationID: "org-default", TaskID: taskID, ExecutingEmployeeID: testEmpID,
		Status: domain.ExecutionStateRunning, AttemptNo: 2, IdempotencyKey: uuid.New().String(), CreatedAt: now,
	})

	_ = repo1.CreateBudgetReservation(ctx, &domain.BudgetReservation{
		ID: res1ID, OrganizationID: "org-default", ExecutionID: exec1ID, StepKey: "s1",
		ReservedUSD: 0.24, Status: "ACTIVE", ExpiresAt: now.Add(5 * time.Minute),
	})
	_ = repo2.CreateBudgetReservation(ctx, &domain.BudgetReservation{
		ID: res2ID, OrganizationID: "org-default", ExecutionID: exec2ID, StepKey: "s2",
		ReservedUSD: 0.35, Status: "ACTIVE", ExpiresAt: now.Add(5 * time.Minute),
	})

	// Concurrently commit execution 1 (cost $0.24) and execution 2 (cost $0.35)
	var wg sync.WaitGroup
	var err1, err2 error

	wg.Add(2)
	go func() {
		defer wg.Done()
		err1 = repo1.CommitExecutionSpend(ctx, res1ID, testEmpID, taskID, 0.24, domain.TaskStateInReview)
	}()
	go func() {
		defer wg.Done()
		err2 = repo2.CommitExecutionSpend(ctx, res2ID, testEmpID, taskID, 0.35, domain.TaskStateInReview)
	}()
	wg.Wait()

	if err1 != nil || err2 != nil {
		t.Fatalf("CommitExecutionSpend returned error: err1=%v, err2=%v", err1, err2)
	}

	// Verify that spend was atomically accumulated without overwriting
	savedEmp, err := repo1.GetEmployeeByID(ctx, testEmpID)
	if err != nil {
		t.Fatalf("failed to reload employee: %v", err)
	}
	expectedSpend := 0.59
	if diff := savedEmp.SpentUSD - expectedSpend; diff < -0.001 || diff > 0.001 {
		t.Fatalf("Employee spent_usd overwritten! Expected %.2f, got %.2f", expectedSpend, savedEmp.SpentUSD)
	}

	savedTask, err := repo1.GetTaskByID(ctx, taskID)
	if err != nil {
		t.Fatalf("failed to reload task: %v", err)
	}
	if diff := savedTask.CostUSD - expectedSpend; diff < -0.001 || diff > 0.001 {
		t.Fatalf("Task cost_usd overwritten! Expected %.2f, got %.2f", expectedSpend, savedTask.CostUSD)
	}
	if savedTask.ExecutionsCount != 2 {
		t.Fatalf("Task execs_count overwritten! Expected 2, got %d", savedTask.ExecutionsCount)
	}
}
