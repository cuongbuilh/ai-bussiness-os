package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cuongbuilh/ai-business-os/backend/internal/handler"
	"github.com/cuongbuilh/ai-business-os/backend/internal/repository"
	"github.com/cuongbuilh/ai-business-os/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

func qaRouter() http.Handler {
	svc := service.NewServices(repository.NewMemoryRepo(), service.NewEventBroker())
	h := handler.NewHandler(svc)
	r := chi.NewRouter()
	r.Get("/employees", h.ListEmployees)
	r.Post("/employees", h.CreateEmployee)
	r.Get("/employees/{employeeId}", h.GetEmployeeDetail)
	r.Get("/tasks", h.ListTasks)
	r.Post("/tasks", h.CreateTask)
	r.Get("/tasks/{taskId}", h.GetTaskDetail)
	r.Patch("/tasks/{taskId}", h.UpdateTaskStatus)
	r.Post("/tasks/{taskId}/accept", h.AcceptTask)
	r.Post("/tasks/{taskId}/execute", h.ExecuteTask)
	r.Get("/approvals", h.ListApprovals)
	r.Post("/approvals/{approvalId}/decide", h.DecideApproval)
	return r
}

func qaCallWithAuth(t *testing.T, r http.Handler, method, path string, input any, auth string) (int, map[string]any) {
	t.Helper()
	var body []byte
	if input != nil {
		body, _ = json.Marshal(input)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var data map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &data)
	return w.Code, data
}

func qaCall(t *testing.T, r http.Handler, method, path string, input any) (int, map[string]any) {
	// Anonymous tests explicitly test lack of credentials
	if path == "/employees" && method == "GET" {
		return qaCallWithAuth(t, r, method, path, input, "")
	}
	if len(path) >= 10 && path[:10] == "/approvals" && method == "POST" {
		return qaCallWithAuth(t, r, method, path, input, "")
	}
	if path == "/tasks/T-103/accept" && method == "POST" {
		return qaCallWithAuth(t, r, method, path, input, "")
	}
	// Normal UI/client requests pass authenticated founder token
	return qaCallWithAuth(t, r, method, path, input, "Bearer founder-h1-token")
}

func TestQA_API_01_AnonymousEmployeeList(t *testing.T) {
	code, _ := qaCall(t, qaRouter(), "GET", "/employees", nil)
	if code != 401 && code != 403 {
		t.Fatalf("expected 401/403, got %d", code)
	}
}

func TestQA_API_02_UIFormCreatesTask(t *testing.T) {
	code, data := qaCall(t, qaRouter(), "POST", "/tasks", map[string]any{"title": "QA UI form", "owner": "q1", "project": "p1", "priority": "HIGH", "criteria": "review", "due": "2026-11-01"})
	if code != 201 || data["id"] == nil {
		t.Fatalf("expected 201 + ID, got %d %v", code, data)
	}
}

func TestQA_API_03_RejectDirectAcceptedStatus(t *testing.T) {
	r := qaRouter()
	code, _ := qaCall(t, r, "PATCH", "/tasks/T-104", map[string]any{"status": "accepted"})
	_, task := qaCall(t, r, "GET", "/tasks/T-104", nil)
	if (code != 400 && code != 403 && code != 409) || task["status"] != "ready" {
		t.Fatalf("expected rejection and ready, got %d %v", code, task["status"])
	}
}

func TestQA_API_04_RejectBacklogAcceptance(t *testing.T) {
	r := qaRouter()
	code, _ := qaCall(t, r, "POST", "/tasks/T-108/accept", map[string]any{"comments": "QA"})
	_, task := qaCall(t, r, "GET", "/tasks/T-108", nil)
	if (code != 400 && code != 403 && code != 409) || task["status"] != "backlog" {
		t.Fatalf("expected rejection and backlog, got %d %v", code, task["status"])
	}
}

func TestQA_API_05_RejectToolOutsideAgentGrants(t *testing.T) {
	code, execution := qaCall(t, qaRouter(), "POST", "/tasks/T-104/execute", map[string]any{})
	if code == 400 || code == 403 {
		return
	}
	if steps, ok := execution["steps"].([]any); ok {
		for _, raw := range steps {
			tool := raw.(map[string]any)["toolName"]
			if tool == "task-mcp" || tool == "docs-mcp" {
				t.Fatalf("ungranted tool invoked: %v, HTTP %d", tool, code)
			}
		}
	}
}

func TestQA_API_06_RejectCallExceedingBudget(t *testing.T) {
	r := qaRouter()
	code, employee := qaCall(t, r, "POST", "/employees", map[string]any{"id": "qa-budget", "displayName": "QA Budget", "kind": "AI", "state": "ACTIVE", "title": "QA", "role": "QA Automation", "managerId": "h1", "model": "Model Y", "provider": "Provider A", "budget": 0.1, "spent": 0, "mcp": []string{}})
	if code != 201 {
		t.Fatalf("setup employee: %d %v", code, employee)
	}
	code, task := qaCall(t, r, "POST", "/tasks", map[string]any{"title": "QA budget", "owner": "qa-budget", "due": "2026-11-01"})
	if code != 201 {
		t.Fatalf("setup task: %d %v", code, task)
	}
	code, execution := qaCall(t, r, "POST", "/tasks/"+task["id"].(string)+"/execute", map[string]any{})
	_, employee = qaCall(t, r, "GET", "/employees/qa-budget", nil)
	if (code != 400 && code != 403) || employee["spent"].(float64) != 0 {
		t.Fatalf("expected deny + zero spend, got HTTP %d, spent=%v, execution=%v", code, employee["spent"], execution["totalCostUsd"])
	}
}

func TestQA_API_07_RejectUnauthenticatedApproval(t *testing.T) {
	code, _ := qaCall(t, qaRouter(), "POST", "/approvals/APR-12/decide", map[string]any{"decision": "APPROVE", "resolvedBy": "h1"})
	if code != 401 && code != 403 {
		t.Fatalf("expected 401/403, got %d", code)
	}
}

func TestQA_API_08_ProjectScope(t *testing.T) {
	r := qaRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/tasks?scope=project:p1", nil)
	req.Header.Set("Authorization", "Bearer founder-h1-token")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var tasks []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &tasks); err != nil {
		t.Fatal(err)
	}
	if len(tasks) == 0 {
		t.Fatal("empty project scope")
	}
	for _, task := range tasks {
		if task["project"] != "p1" {
			t.Fatalf("out-of-scope task %v", task["id"])
		}
	}
}

func TestQA_API_09_RejectUnknownTaskStatus(t *testing.T) {
	r := qaRouter()
	code, _ := qaCall(t, r, "PATCH", "/tasks/T-104", map[string]any{"status": "made_up_state"})
	_, task := qaCall(t, r, "GET", "/tasks/T-104", nil)
	if (code != 400 && code != 409) || task["status"] != "ready" {
		t.Fatalf("expected rejection and ready, got %d %v", code, task["status"])
	}
}

func TestQA_API_10_RejectUnknownManager(t *testing.T) {
	code, employee := qaCall(t, qaRouter(), "POST", "/employees", map[string]any{"id": "qa-orphan", "displayName": "QA Orphan", "managerId": "missing", "title": "QA", "role": "QA Automation"})
	if code != 400 && code != 409 {
		t.Fatalf("expected invalid manager rejected, got %d %v", code, employee)
	}
}
