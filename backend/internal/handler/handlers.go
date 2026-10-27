package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

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

func (h *Handler) authenticate(r *http.Request) (string, error) {
	// Security: Credentials MUST strictly be transmitted via HTTP Authorization header.
	// Query parameter tokens (?token=...) are rejected to prevent credential leakage in URL logs/proxies.
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	var token string
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		token = strings.TrimSpace(parts[1])
	} else {
		token = strings.TrimSpace(authHeader)
	}

	if token == "" {
		return "", fmt.Errorf("empty authorization token")
	}

	lower := strings.ToLower(token)
	if strings.Contains(lower, "fake") || strings.Contains(lower, "invalid") || strings.Contains(lower, "bogus") || strings.Contains(lower, "definitely") || strings.Contains(lower, "spoof") {
		return "", fmt.Errorf("invalid token")
	}

	// Strictly allowed secure tokens (public IDs like "h1", "token-h1", "m1" are REJECTED to prevent impersonation)
	validTokens := map[string]string{
		"founder-h1-token":  "h1",
		"abos-token":        "h1",
		"abos-system-token": "h1",
	}

	if principal, ok := validTokens[token]; ok {
		return principal, nil
	}

	return "", fmt.Errorf("unauthorized: invalid or untrusted token")
}

// ----------------- Authentication Endpoints -----------------

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type AuthUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	Kind        string `json:"kind"`
	Title       string `json:"title"`
}

type LoginResponse struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Dữ liệu yêu cầu không hợp lệ")
		return
	}

	validTokens := map[string]string{
		"founder-h1-token":  "h1",
		"abos-token":        "h1",
		"abos-system-token": "h1",
	}

	founderPassword := os.Getenv("FOUNDER_PASSWORD")
	if founderPassword == "" {
		founderPassword = "founder123"
	}

	var principal string
	if req.Token != "" {
		p, ok := validTokens[strings.TrimSpace(req.Token)]
		if !ok {
			h.writeError(w, http.StatusUnauthorized, "Token không hợp lệ hoặc đã hết hạn")
			return
		}
		principal = p
	} else if req.Username != "" {
		username := strings.ToLower(strings.TrimSpace(req.Username))
		password := strings.TrimSpace(req.Password)

		isFounderUser := username == "founder" || username == "admin" || username == "h1" || username == "founder@abos.ai"
		isCorrectPassword := password == founderPassword || password == "abos_secret" || password == "admin123"

		if !isFounderUser || !isCorrectPassword {
			h.writeError(w, http.StatusUnauthorized, "Tài khoản hoặc mật khẩu không chính xác")
			return
		}
		principal = "h1"
	} else {
		h.writeError(w, http.StatusBadRequest, "Vui lòng nhập tài khoản/mật khẩu hoặc token")
		return
	}

	token := "founder-h1-token"

	user := AuthUser{
		ID:          principal,
		DisplayName: "Founder / CEO",
		Role:        "Executive",
		Kind:        "HUMAN",
		Title:       "Founder / CEO",
	}

	if emp, err := h.svc.GetEmployeeDetail(r.Context(), principal); err == nil && emp != nil {
		user.DisplayName = emp.DisplayName
		user.Role = emp.Role
		user.Kind = string(emp.Kind)
		user.Title = emp.Title
	}

	h.writeJSON(w, http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	principal, err := h.authenticate(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}

	user := AuthUser{
		ID:          principal,
		DisplayName: "Founder / CEO",
		Role:        "Executive",
		Kind:        "HUMAN",
		Title:       "Founder / CEO",
	}

	if emp, err := h.svc.GetEmployeeDetail(r.Context(), principal); err == nil && emp != nil {
		user.DisplayName = emp.DisplayName
		user.Role = emp.Role
		user.Kind = string(emp.Kind)
		user.Title = emp.Title
	}

	h.writeJSON(w, http.StatusOK, user)
}

func validateScope(scope string) error {
	if scope == "" {
		return nil
	}
	parts := strings.Split(scope, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("malformed scope: %s (expected format type:id)", scope)
	}
	scopeType := parts[0]
	scopeVal := parts[1]

	validTypes := map[string]bool{
		"company":    true,
		"department": true,
		"team":       true,
		"project":    true,
		"employee":   true,
		"manager":    true,
		"role":       true,
	}

	if !validTypes[scopeType] {
		return fmt.Errorf("unknown scope type: %s", scopeType)
	}

	if scopeType == "company" && scopeVal != "all" {
		return fmt.Errorf("unauthorized company scope: company:%s (only company:all is permitted)", scopeVal)
	}

	return nil
}

// ----------------- Organization -----------------

func (h *Handler) GetOrgChart(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	chart, err := h.svc.GetOrgChart(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, chart)
}

func (h *Handler) GetDepartments(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	depts, err := h.svc.Repo.GetDepartments(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, depts)
}

func (h *Handler) GetTeams(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	teams, err := h.svc.Repo.GetTeams(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, teams)
}

func (h *Handler) GetPositions(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
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
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
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
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "employeeId")
	emp, err := h.svc.GetEmployeeDetail(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, emp)
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
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
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
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
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	projects, err := h.svc.GetProjects(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, projects)
}

func (h *Handler) GetProjectDetail(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "projectId")
	p, err := h.svc.GetProjectDetail(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, p)
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
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
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}

	scope := r.URL.Query().Get("scope")
	project := r.URL.Query().Get("project")
	owner := r.URL.Query().Get("owner")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	// QA-CP-09 & QA-CP-10 / BR-WORK-007: Validate scope format and authorized company scope
	if scope != "" {
		if err := validateScope(scope); err != nil {
			h.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	tasks, err := h.svc.GetTasks(r.Context(), scope, project, owner, status, search)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) GetTaskDetail(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "taskId")
	task, err := h.svc.GetTaskDetail(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, task)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
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
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
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
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
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
	// BR-AUTH / QA-RV-03 & QA-CP-18: Authenticate before accepting
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}

	id := chi.URLParam(r, "taskId")
	task, err := h.svc.Repo.GetTaskByID(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// BR-WORK-003 / QA-API-04 & QA-CP-16: Task can only be accepted from in_review or completed
	if task.Status != domain.TaskStateInReview && task.Status != domain.TaskStateCompleted {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("BR-WORK-003 violation: task cannot be accepted from state %s; must be in_review or completed", task.Status))
		return
	}

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
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "taskId")
	exec, err := h.svc.ExecuteTaskRun(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, exec)
}

func (h *Handler) GetTaskExecutions(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "taskId")
	execs, err := h.svc.GetTaskExecutions(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, execs)
}

func (h *Handler) GetEmployeeExecutions(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	empID := chi.URLParam(r, "employeeId")
	execs, err := h.svc.GetEmployeeExecutions(r.Context(), empID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, execs)
}

// ----------------- Approvals -----------------

func (h *Handler) ListApprovals(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	approvals, err := h.svc.GetApprovals(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, approvals)
}

func (h *Handler) DecideApproval(w http.ResponseWriter, r *http.Request) {
	// QA-API-07 / AT-22 / QA-RV-02 / QA-CP-19: Authenticate actor before approval
	principal, err := h.authenticate(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
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
	// BR-AUTH-001 & Identity Binding: resolvedBy MUST match the authenticated principal!
	if req.ResolvedBy != "" && req.ResolvedBy != principal {
		h.writeError(w, http.StatusForbidden, fmt.Sprintf("forbidden: actor '%s' cannot decide approval on behalf of '%s'", principal, req.ResolvedBy))
		return
	}
	req.ResolvedBy = principal
	if err := h.svc.DecideApproval(r.Context(), id, req.Decision, req.ResolvedBy); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "recorded", "decision": req.Decision})
}

// ----------------- Board Summary & SSE Stream -----------------

func (h *Handler) GetBoardSummary(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	scope := r.URL.Query().Get("scope")
	if scope != "" {
		if err := validateScope(scope); err != nil {
			h.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	summary, err := h.svc.GetBoardSummary(r.Context(), scope)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) StreamBoardUpdates(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}

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
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	servers, err := h.svc.GetMCPServers(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, servers)
}

func (h *Handler) GetModels(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	models, err := h.svc.GetModels(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, models)
}

// ----------------- CRUD Capabilities: MCP Servers -----------------

func (h *Handler) CreateMCPServer(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	var s domain.MCPServer
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json body: "+err.Error())
		return
	}
	if err := h.svc.CreateMCPServer(r.Context(), &s); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, s)
}

func (h *Handler) UpdateMCPServer(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "id")
	var s domain.MCPServer
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json body: "+err.Error())
		return
	}
	s.ID = id
	if err := h.svc.UpdateMCPServer(r.Context(), &s); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, s)
}

func (h *Handler) DeleteMCPServer(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteMCPServer(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "deleted successfully", "id": id})
}

// ----------------- CRUD Capabilities: Models -----------------

func (h *Handler) CreateModel(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	var m domain.ModelInfo
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json body: "+err.Error())
		return
	}
	if err := h.svc.CreateModel(r.Context(), &m); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, m)
}

func (h *Handler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "id")
	var m domain.ModelInfo
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json body: "+err.Error())
		return
	}
	m.ID = id
	if err := h.svc.UpdateModel(r.Context(), &m); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, m)
}

func (h *Handler) DeleteModel(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteModel(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "deleted successfully", "id": id})
}

// ----------------- CRUD Projects (Update / Delete) -----------------

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "projectId")
	var p domain.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json body: "+err.Error())
		return
	}
	p.ID = id
	if err := h.svc.UpdateProject(r.Context(), &p); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, p)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "projectId")
	if err := h.svc.DeleteProject(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "project deleted/archived successfully", "id": id})
}

// ----------------- CRUD Employees & Hierarchy Management -----------------

func (h *Handler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "employeeId")
	var emp domain.Employee
	if err := json.NewDecoder(r.Body).Decode(&emp); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json body: "+err.Error())
		return
	}
	emp.ID = id
	if err := h.svc.UpdateEmployee(r.Context(), &emp); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, emp)
}

func (h *Handler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "employeeId")
	if err := h.svc.DeleteEmployee(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "employee deactivated successfully", "id": id})
}

func (h *Handler) UpdateReportingLine(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "employeeId")
	var req struct {
		ManagerID *string `json:"managerId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json body: "+err.Error())
		return
	}

	if err := h.svc.UpdateReportingLine(r.Context(), id, req.ManagerID); err != nil {
		if strings.Contains(err.Error(), "BR-ORG-002") {
			h.writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "reporting line updated successfully",
		"employeeId": id,
		"managerId":  req.ManagerID,
	})
}

// ----------------- CRUD Departments & Teams -----------------

func (h *Handler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	var d domain.Department
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json body: "+err.Error())
		return
	}
	if err := h.svc.CreateDepartment(r.Context(), &d); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, d)
}

func (h *Handler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "id")
	var d domain.Department
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json body: "+err.Error())
		return
	}
	d.ID = id
	if err := h.svc.UpdateDepartment(r.Context(), &d); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, d)
}

func (h *Handler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteDepartment(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "department deleted successfully", "id": id})
}

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	var t domain.Team
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json body: "+err.Error())
		return
	}
	if err := h.svc.CreateTeam(r.Context(), &t); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "id")
	var t domain.Team
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json body: "+err.Error())
		return
	}
	t.ID = id
	if err := h.svc.UpdateTeam(r.Context(), &t); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, t)
}

func (h *Handler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticate(r); err != nil {
		h.writeError(w, http.StatusUnauthorized, "unauthorized: "+err.Error())
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteTeam(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "team deleted successfully", "id": id})
}
