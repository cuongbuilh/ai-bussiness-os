package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cuongbuilh/ai-business-os/backend/internal/domain"
	"github.com/cuongbuilh/ai-business-os/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *service.Services
}

func NewHandler(svc *service.Services) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "ai-business-os-control-plane",
		"version": "1.1.0",
	})
}

// ----------------- Organization -----------------

func (h *Handler) GetOrgChart(w http.ResponseWriter, r *http.Request) {
	chart, err := h.svc.GetOrgChart(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, chart)
}

func (h *Handler) GetDepartments(w http.ResponseWriter, r *http.Request) {
	depts, err := h.svc.Repo.GetDepartments(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, depts)
}

func (h *Handler) GetTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := h.svc.Repo.GetTeams(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, teams)
}

func (h *Handler) GetPositions(w http.ResponseWriter, r *http.Request) {
	positions, err := h.svc.Repo.GetPositions(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, positions)
}

// ----------------- Employees -----------------

func (h *Handler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	// BR-AUTH / AT-17 / QA-API-01: Reject unauthenticated anonymous employee list
	if r.Header.Get("Authorization") == "" {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: missing authorization header")
		return
	}

	employees, err := h.svc.GetEmployees(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, employees)
}

func (h *Handler) GetEmployeeDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "employeeId")
	emp, err := h.svc.GetEmployeeDetail(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, emp)
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var emp domain.Employee
	if err := json.NewDecoder(r.Body).Decode(&emp); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.CreateEmployee(r.Context(), &emp); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, emp)
}

func (h *Handler) GetEmployeeCapabilities(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "employeeId")
	caps, err := h.svc.GetEmployeeCapabilities(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, caps)
}

// ----------------- Projects -----------------

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.svc.GetProjects(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, projects)
}

func (h *Handler) GetProjectDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectId")
	p, err := h.svc.GetProjectDetail(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, p)
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var p domain.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.CreateProject(r.Context(), &p); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, p)
}

// ----------------- Tasks -----------------

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	project := r.URL.Query().Get("project")
	owner := r.URL.Query().Get("owner")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	tasks, err := h.svc.GetTasks(r.Context(), scope, project, owner, status, search)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) GetTaskDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "taskId")
	task, err := h.svc.GetTaskDetail(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, task)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task domain.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.CreateTask(r.Context(), &task); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, task)
}

func (h *Handler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "taskId")
	var req struct {
		Status domain.TaskState `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// QA-API-03 & QA-API-09: Validate task status state machine
	validStatuses := map[domain.TaskState]bool{
		domain.TaskStateDraft:      true,
		domain.TaskStateBacklog:    true,
		domain.TaskStateReady:      true,
		domain.TaskStateInProgress: true,
		domain.TaskStateBlocked:    true,
		domain.TaskStateInReview:   true,
		domain.TaskStateCompleted:  true,
		domain.TaskStateRejected:   true,
		domain.TaskStateCancelled:  true,
	}

	if req.Status == domain.TaskStateAccepted {
		h.writeError(w, http.StatusBadRequest, "direct transition to accepted status is forbidden; tasks must be accepted through formal acceptance workflow")
		return
	}

	if !validStatuses[req.Status] {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid task status: %s", req.Status))
		return
	}

	if err := h.svc.UpdateTaskStatus(r.Context(), id, req.Status); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "updated", "taskId": id, "newStatus": req.Status})
}

func (h *Handler) AssignTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "taskId")
	var req struct {
		OwnerID       string   `json:"ownerId"`
		Collaborators []string `json:"collaborators"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.AssignTask(r.Context(), id, req.OwnerID, req.Collaborators); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "assigned"})
}

func (h *Handler) AcceptTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "taskId")
	var req struct {
		Comments string `json:"comments"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := h.svc.AcceptTask(r.Context(), id, req.Comments); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}

// ----------------- Executions -----------------

func (h *Handler) ExecuteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "taskId")
	exec, err := h.svc.ExecuteTaskRun(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, exec)
}

func (h *Handler) GetTaskExecutions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "taskId")
	execs, err := h.svc.GetTaskExecutions(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, execs)
}

// ----------------- Approvals -----------------

func (h *Handler) ListApprovals(w http.ResponseWriter, r *http.Request) {
	approvals, err := h.svc.GetApprovals(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, approvals)
}

func (h *Handler) DecideApproval(w http.ResponseWriter, r *http.Request) {
	// QA-API-07 / AT-22: Reject unauthenticated approval decision
	if r.Header.Get("Authorization") == "" {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: missing authorization header")
		return
	}

	id := chi.URLParam(r, "approvalId")
	var req struct {
		Decision   string `json:"decision"`
		ResolvedBy string `json:"resolvedBy"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ResolvedBy == "" {
		req.ResolvedBy = "h1" // Default to Founder in Web Console
	}
	if err := h.svc.DecideApproval(r.Context(), id, req.Decision, req.ResolvedBy); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "recorded", "decision": req.Decision})
}

// ----------------- Board Summary & SSE Stream -----------------

func (h *Handler) GetBoardSummary(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	summary, err := h.svc.GetBoardSummary(r.Context(), scope)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) StreamBoardUpdates(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := h.svc.Broker.Subscribe()
	defer h.svc.Broker.Unsubscribe(ch)

	// Send initial ping
	fmt.Fprintf(w, "event: ping\ndata: {\"connected\":true}\n\n")
	flusher.Flush()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprint(w, msg)
			flusher.Flush()
		}
	}
}

// ----------------- Capabilities & Models -----------------

func (h *Handler) GetMCPServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.svc.GetMCPServers(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, servers)
}

func (h *Handler) GetModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.svc.GetModels(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, models)
}
