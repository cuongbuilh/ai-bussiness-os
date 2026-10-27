package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/cuongbuilh/ai-business-os/backend/internal/domain"
	"github.com/cuongbuilh/ai-business-os/backend/internal/repository"
	"github.com/cuongbuilh/ai-business-os/backend/internal/service"
)

type budgetBarrierRepo struct {
	repository.Repository
	mu      sync.Mutex
	reads   int
	release chan struct{}
}

func (r *budgetBarrierRepo) GetEmployeeByID(ctx context.Context, id string) (*domain.Employee, error) {
	emp, err := r.Repository.GetEmployeeByID(ctx, id)
	if id != "qa-race" || err != nil {
		return emp, err
	}
	r.mu.Lock()
	r.reads++
	wait := r.reads <= 2
	if r.reads == 2 {
		close(r.release)
	}
	r.mu.Unlock()
	if wait {
		<-r.release
	}
	return emp, nil
}

func qaCallWithAuthorization(t *testing.T, r http.Handler, method, path string, input any, authorization string) (int, map[string]any) {
	t.Helper()
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var data map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &data)
	return w.Code, data
}

func TestQA_RV_01_RejectForgedBearerToken(t *testing.T) {
	code, _ := qaCallWithAuthorization(t, qaRouter(), "GET", "/employees", nil, "Bearer definitely-invalid")
	if code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Fatalf("forged bearer token accepted: HTTP %d", code)
	}
}

func TestQA_RV_02_RejectSpoofedApprovalActor(t *testing.T) {
	code, _ := qaCallWithAuthorization(t, qaRouter(), "POST", "/approvals/APR-12/decide", map[string]any{"decision": "APPROVE", "resolvedBy": "h1"}, "Bearer definitely-invalid")
	if code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Fatalf("forged token approved as founder: HTTP %d", code)
	}
}

func TestQA_RV_03_RejectAnonymousInReviewAcceptance(t *testing.T) {
	code, _ := qaCall(t, qaRouter(), "POST", "/tasks/T-103/accept", map[string]any{"comments": ""})
	if code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Fatalf("anonymous acceptance succeeded: HTTP %d", code)
	}
}

func TestQA_RV_04_RejectReadyToCompletedJump(t *testing.T) {
	r := qaRouter()
	code, _ := qaCall(t, r, "PATCH", "/tasks/T-104", map[string]any{"status": "completed"})
	_, task := qaCall(t, r, "GET", "/tasks/T-104", nil)
	if code < 400 || task["status"] != "ready" {
		t.Fatalf("READY to COMPLETED jump accepted: HTTP %d, status=%v", code, task["status"])
	}
}

func TestQA_RV_05_ConcurrentBudgetNeverExceedsCap(t *testing.T) {
	repo := &budgetBarrierRepo{Repository: repository.NewMemoryRepo(), release: make(chan struct{})}
	svc := service.NewServices(repo, service.NewEventBroker())
	ctx := context.Background()
	manager := "h1"
	if err := svc.CreateEmployee(ctx, &domain.Employee{ID: "qa-race", DisplayName: "QA Race", ManagerID: &manager, Model: "Model Y", Role: "QA Automation", BudgetUSD: 0.37}); err != nil {
		t.Fatal(err)
	}
	task := &domain.Task{Title: "QA budget race", AccountableOwnerID: "qa-race"}
	if err := svc.CreateTask(ctx, task); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.ExecuteTaskRun(ctx, task.ID)
			if err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if successes > 1 {
		t.Fatalf("cap $0.37 admitted %d concurrent runs", successes)
	}
}
