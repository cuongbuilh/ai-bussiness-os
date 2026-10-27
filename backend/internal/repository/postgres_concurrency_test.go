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
