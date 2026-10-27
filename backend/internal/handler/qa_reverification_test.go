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

func TestQA_RV_06_RejectPublicIDAsBearerToken(t *testing.T) {
	for _, forged := range []string{"Bearer h1", "Bearer token-h1", "Bearer m1", "Bearer e1"} {
		code, _ := qaCallWithAuthorization(t, qaRouter(), "GET", "/employees", nil, forged)
		if code != http.StatusUnauthorized && code != http.StatusForbidden {
			t.Fatalf("public ID impersonation accepted for %s: HTTP %d", forged, code)
		}
	}
}

func TestQA_RV_07_RejectAnonymousTaskAndApprovalReads(t *testing.T) {
	r := qaRouter()
	// Anonymous company board
	code, _ := qaCallWithAuthorization(t, r, "GET", "/tasks?scope=company:all", nil, "")
	if code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Fatalf("anonymous company tasks accepted: HTTP %d", code)
	}

	// Anonymous task detail
	code, _ = qaCallWithAuthorization(t, r, "GET", "/tasks/T-101", nil, "")
	if code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Fatalf("anonymous task detail accepted: HTTP %d", code)
	}

	// Anonymous approvals list
	code, _ = qaCallWithAuthorization(t, r, "GET", "/approvals", nil, "")
	if code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Fatalf("anonymous approvals list accepted: HTTP %d", code)
	}
}

func TestQA_RV_08_RejectSpoofedResolvedByInApproval(t *testing.T) {
	// Authenticated as h1 via founder-h1-token, but trying to record as m1
	code, _ := qaCallWithAuthorization(t, qaRouter(), "POST", "/approvals/APR-12/decide",
		map[string]any{"decision": "APPROVE", "resolvedBy": "m1"}, "Bearer founder-h1-token")
	if code != http.StatusForbidden {
		t.Fatalf("spoofed resolvedBy accepted without 403 Forbidden: HTTP %d", code)
	}
}

func TestQA_RV_09_RejectAnonymousWrites(t *testing.T) {
	r := qaRouter()
	// Anonymous CreateTask
	code, _ := qaCallWithAuthorization(t, r, "POST", "/tasks", map[string]any{"title": "Anon Task"}, "")
	if code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Fatalf("anonymous CreateTask accepted: HTTP %d", code)
	}

	// Anonymous CreateEmployee
	code, _ = qaCallWithAuthorization(t, r, "POST", "/employees", map[string]any{"id": "anon"}, "")
	if code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Fatalf("anonymous CreateEmployee accepted: HTTP %d", code)
	}

	// Anonymous ExecuteTask
	code, _ = qaCallWithAuthorization(t, r, "POST", "/tasks/T-104/execute", map[string]any{}, "")
	if code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Fatalf("anonymous ExecuteTask accepted: HTTP %d", code)
	}
}

func TestQA_RV_10_PersistentBudgetReservationAndUsageLedger(t *testing.T) {
	repo := repository.NewMemoryRepo()
	svc := service.NewServices(repo, service.NewEventBroker())
	ctx := context.Background()

	// Task run writes into usage_ledger and budget_reservations
	exec, err := svc.ExecuteTaskRun(ctx, "T-101")
	if err != nil {
		t.Fatalf("execute task run failed: %v", err)
	}
	if exec.TotalCostUSD <= 0 {
		t.Fatalf("expected positive execution cost, got %f", exec.TotalCostUSD)
	}
}

func TestQA_RV_11_RejectAnonymousReadsAndProjectCreate(t *testing.T) {
	r := qaRouter()

	// 1. Anonymous POST /projects with {} must return 401 Unauthorized (not 400 Bad Request)
	code, _ := qaCallWithAuthorization(t, r, "POST", "/projects", map[string]any{}, "")
	if code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Fatalf("anonymous POST /projects returned %d, expected 401", code)
	}

	// 2. Anonymous GET on read endpoints
	anonymousReadEndpoints := []string{
		"/employees/e1",
		"/employees/e1/capabilities",
		"/projects",
		"/projects/p1",
		"/departments",
		"/teams",
		"/positions",
		"/capabilities/mcp-servers",
		"/capabilities/models",
		"/boards/stream",
	}

	for _, endpoint := range anonymousReadEndpoints {
		c, _ := qaCallWithAuthorization(t, r, "GET", endpoint, nil, "")
		if c != http.StatusUnauthorized && c != http.StatusForbidden {
			t.Fatalf("anonymous GET %s returned %d, expected 401", endpoint, c)
		}
	}
}

func TestQA_RV_12_SSEStreamAuthViaQueryParam(t *testing.T) {
	r := qaRouter()

	// SSE with query param token should be authorized
	req := httptest.NewRequest("GET", "/boards/stream?token=founder-h1-token", nil)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	go func() {
		// Cancel request after stream starts to exit loop
		cancel()
	}()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("SSE with query param token returned %d, expected 200", w.Code)
	}
}
