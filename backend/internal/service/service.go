package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cuongbuilh/ai-business-os/backend/internal/domain"
	"github.com/cuongbuilh/ai-business-os/backend/internal/repository"
	"github.com/google/uuid"
)

type Services struct {
	Repo               repository.Repository
	Broker             *EventBroker
	budgetMu           sync.Mutex
	activeReservations map[string]float64
	committedSpend     map[string]float64
}

func NewServices(repo repository.Repository, broker *EventBroker) *Services {
	return &Services{
		Repo:               repo,
		Broker:             broker,
		activeReservations: make(map[string]float64),
		committedSpend:     make(map[string]float64),
	}
}

// ----------------- Organization & Employee -----------------

func (s *Services) GetOrgChart(ctx context.Context) (*domain.OrgNode, error) {
	return s.Repo.GetOrgTree(ctx)
}

func (s *Services) GetEmployees(ctx context.Context) ([]domain.Employee, error) {
	return s.Repo.GetEmployees(ctx)
}

func (s *Services) GetEmployeeDetail(ctx context.Context, id string) (*domain.Employee, error) {
	return s.Repo.GetEmployeeByID(ctx, id)
}

func (s *Services) CreateEmployee(ctx context.Context, emp *domain.Employee) error {
	if emp.ID == "" {
		emp.ID = fmt.Sprintf("e-%s", strings.ToLower(uuid.New().String()[:8]))
	}

	// BR-ORG-002: Check cycle and self-reporting
	if emp.ManagerID != nil && *emp.ManagerID != "" {
		if *emp.ManagerID == emp.ID {
			return fmt.Errorf("BR-ORG-002 violation: employee cannot manage themselves")
		}

		// BR-ORG-001: Manager must exist in the organization
		mgr, err := s.Repo.GetEmployeeByID(ctx, *emp.ManagerID)
		if err != nil || mgr == nil {
			return fmt.Errorf("BR-ORG-001 violation: manager %s does not exist", *emp.ManagerID)
		}
	}

	if emp.DisplayName == "" {
		emp.DisplayName = emp.Name
	}
	if emp.Name == "" {
		emp.Name = emp.DisplayName
	}
	if emp.Kind == "" {
		emp.Kind = domain.EmployeeKindAI
	}
	if emp.State == "" {
		emp.State = domain.EmployeeStateActive
	}
	if emp.KPI == 0 {
		emp.KPI = 90
	}

	if err := s.Repo.CreateEmployee(ctx, emp); err != nil {
		return err
	}

	s.Broker.Publish("employee", "created", emp)
	return nil
}

func (s *Services) GetEmployeeCapabilities(ctx context.Context, id string) (map[string]interface{}, error) {
	emp, err := s.Repo.GetEmployeeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	positions, _ := s.Repo.GetPositions(ctx)
	var posCap []string
	for _, p := range positions {
		if p.RoleCode == emp.Role {
			posCap = p.CapMCPs
			break
		}
	}

	// Calculate effective tool grants (Position Ceiling ∩ Agent Grant)
	effective := make([]string, 0)
	capMap := make(map[string]bool)
	for _, c := range posCap {
		capMap[c] = true
	}
	for _, m := range emp.MCP {
		if capMap[m] {
			effective = append(effective, m)
		}
	}

	return map[string]interface{}{
		"employeeId":          emp.ID,
		"employeeName":        emp.DisplayName,
		"role":                emp.Role,
		"positionCeilingMCPs": posCap,
		"agentGrantedMCPs":    emp.MCP,
		"effectiveMCPs":       effective,
		"skills":              emp.Skills,
		"primaryModel":        emp.Model,
		"provider":            emp.Provider,
	}, nil
}

// ----------------- Projects -----------------

func (s *Services) GetProjects(ctx context.Context) ([]domain.Project, error) {
	return s.Repo.GetProjects(ctx)
}

func (s *Services) GetProjectDetail(ctx context.Context, id string) (*domain.Project, error) {
	return s.Repo.GetProjectByID(ctx, id)
}

func (s *Services) CreateProject(ctx context.Context, p *domain.Project) error {
	if p.ID == "" {
		p.ID = fmt.Sprintf("p-%s", strings.ToLower(uuid.New().String()[:8]))
	}
	if p.AccountableOwnerID == "" {
		return fmt.Errorf("BR-WORK-001 violation: project must have an accountable owner")
	}

	if err := s.Repo.CreateProject(ctx, p); err != nil {
		return err
	}

	s.Broker.Publish("project", "created", p)
	return nil
}

// ----------------- Tasks -----------------

func (s *Services) GetTasks(ctx context.Context, scope, projectFilter, ownerFilter, statusFilter, search string) ([]domain.Task, error) {
	return s.Repo.GetTasks(ctx, scope, projectFilter, ownerFilter, statusFilter, search)
}

func (s *Services) GetTaskDetail(ctx context.Context, id string) (*domain.Task, error) {
	return s.Repo.GetTaskByID(ctx, id)
}

func (s *Services) CreateTask(ctx context.Context, task *domain.Task) error {
	// BR-WORK-002: Task must have an accountable owner before ready
	if task.AccountableOwnerID == "" {
		return fmt.Errorf("BR-WORK-002 violation: task must have an accountable owner")
	}

	if task.Status == "" {
		task.Status = domain.TaskStateReady
	}
	if task.Priority == "" {
		task.Priority = "MEDIUM"
	}
	if task.DueAt == "" {
		task.DueAt = time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02")
	}
	if task.OriginType == "" {
		task.OriginType = "HUMAN"
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

	if err := s.Repo.CreateTask(ctx, task); err != nil {
		return err
	}

	s.Broker.Publish("task", "created", task)
	return nil
}

func (s *Services) UpdateTaskStatus(ctx context.Context, taskID string, newStatus domain.TaskState) error {
	// BR-WORK-003: Direct transition to accepted via generic patch status is strictly forbidden
	if newStatus == domain.TaskStateAccepted {
		return fmt.Errorf("BR-WORK-003 violation: direct transition to accepted status is forbidden; tasks must be accepted through formal acceptance workflow")
	}

	task, err := s.Repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}

	// QA-RV-04 & QA-CP-17: State machine transitions
	// DRAFT -> BACKLOG -> READY -> IN_PROGRESS -> (BLOCKED | IN_REVIEW) -> COMPLETED -> ACCEPTED
	allowedTransitions := map[domain.TaskState][]domain.TaskState{
		domain.TaskStateDraft:      {domain.TaskStateBacklog, domain.TaskStateReady, domain.TaskStateCancelled},
		domain.TaskStateBacklog:    {domain.TaskStateReady, domain.TaskStateDraft, domain.TaskStateCancelled},
		domain.TaskStateReady:      {domain.TaskStateInProgress, domain.TaskStateBlocked, domain.TaskStateBacklog, domain.TaskStateCancelled},
		domain.TaskStateInProgress: {domain.TaskStateInReview, domain.TaskStateBlocked, domain.TaskStateReady, domain.TaskStateCancelled, domain.TaskStateCompleted},
		domain.TaskStateBlocked:    {domain.TaskStateReady, domain.TaskStateInProgress, domain.TaskStateCancelled},
		domain.TaskStateInReview:   {domain.TaskStateCompleted, domain.TaskStateInProgress, domain.TaskStateRejected, domain.TaskStateCancelled},
		domain.TaskStateCompleted:  {domain.TaskStateInProgress, domain.TaskStateInReview, domain.TaskStateCancelled},
		domain.TaskStateRejected:   {domain.TaskStateInProgress, domain.TaskStateReady, domain.TaskStateCancelled},
	}

	valid := false
	for _, next := range allowedTransitions[task.Status] {
		if next == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("BR-WORK-006 violation: invalid task status transition from %s to %s", task.Status, newStatus)
	}

	oldStatus := task.Status
	task.Status = newStatus

	if err := s.Repo.UpdateTask(ctx, task); err != nil {
		return err
	}

	s.Broker.Publish("task", "status_changed", map[string]interface{}{
		"taskId":    task.ID,
		"oldStatus": oldStatus,
		"newStatus": newStatus,
		"task":      task,
	})
	return nil
}

func (s *Services) AssignTask(ctx context.Context, taskID, newOwnerID string, collab []string) error {
	task, err := s.Repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}

	emp, err := s.Repo.GetEmployeeByID(ctx, newOwnerID)
	if err != nil {
		return fmt.Errorf("new owner employee not found: %w", err)
	}

	task.AccountableOwnerID = emp.ID
	task.OwnerName = emp.DisplayName
	if collab != nil {
		task.Collaborators = collab
	}

	if err := s.Repo.UpdateTask(ctx, task); err != nil {
		return err
	}

	s.Broker.Publish("task", "assigned", task)
	return nil
}

func (s *Services) AcceptTask(ctx context.Context, taskID string, comments string) error {
	task, err := s.Repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}

	// BR-WORK-003: Task can only be accepted if it has completed work (in_review or completed)
	if task.Status != domain.TaskStateInReview && task.Status != domain.TaskStateCompleted {
		return fmt.Errorf("BR-WORK-003 violation: task cannot be accepted from state %s; must be in_review or completed", task.Status)
	}

	now := time.Now()
	task.Status = domain.TaskStateAccepted
	task.AcceptedAt = &now

	if err := s.Repo.UpdateTask(ctx, task); err != nil {
		return err
	}

	s.Broker.Publish("task", "accepted", map[string]interface{}{
		"taskId":     task.ID,
		"acceptedAt": now,
		"comments":   comments,
		"task":       task,
	})
	return nil
}

// ----------------- Task Execution Engine -----------------

func (s *Services) ExecuteTaskRun(ctx context.Context, taskID string) (*domain.Execution, error) {
	task, err := s.Repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	emp, err := s.Repo.GetEmployeeByID(ctx, task.AccountableOwnerID)
	if err != nil {
		return nil, fmt.Errorf("executing agent employee not found: %w", err)
	}

	// Calculate effective tool grants (Position Ceiling ∩ Agent Grant)
	caps, _ := s.GetEmployeeCapabilities(ctx, emp.ID)
	var effectiveMCPs []string
	if caps != nil {
		if rawList, ok := caps["effectiveMCPs"].([]string); ok {
			effectiveMCPs = rawList
		}
	}
	effectiveMap := make(map[string]bool)
	for _, m := range effectiveMCPs {
		effectiveMap[m] = true
	}

	estimatedRunCost := 0.37
	now := time.Now()
	execID := fmt.Sprintf("exec-%s", uuid.New().String()[:8])
	exec := &domain.Execution{
		ID:                  execID,
		OrganizationID:      task.OrganizationID,
		TaskID:              task.ID,
		ExecutingEmployeeID: emp.ID,
		ExecutingEmpName:    emp.DisplayName,
		Status:              domain.ExecutionStateRunning,
		AttemptNo:           task.ExecutionsCount + 1,
		IdempotencyKey:      uuid.New().String(),
		CreatedAt:           now,
		StartedAt:           &now,
		Logs:                []string{fmt.Sprintf("[%s] Agent %s initialized task run #%d", now.Format("15:04:05"), emp.DisplayName, task.ExecutionsCount+1)},
		Steps:               make([]domain.ExecutionStep, 0),
	}
	if err := s.Repo.CreateExecution(ctx, exec); err != nil {
		return nil, fmt.Errorf("failed to initialize execution: %w", err)
	}

	// BR-BUD-001 & BR-BUD-003 / QA-RV-05: Durable atomic budget reservation across multiple instances
	resID := fmt.Sprintf("res-%s", uuid.New().String()[:8])
	res := &domain.BudgetReservation{
		ID:             resID,
		OrganizationID: task.OrganizationID,
		ExecutionID:    exec.ID,
		StepKey:        "task_run_reserve",
		ReservedUSD:    estimatedRunCost,
		Status:         "ACTIVE",
		ExpiresAt:      now.Add(10 * time.Minute),
	}

	if err := s.Repo.AtomicReserveBudget(ctx, res, emp.ID, estimatedRunCost); err != nil {
		return nil, err
	}

	s.budgetMu.Lock()
	s.activeReservations[emp.ID] += estimatedRunCost
	s.budgetMu.Unlock()

	hasReserved := true
	defer func() {
		if hasReserved {
			s.budgetMu.Lock()
			_ = s.Repo.UpdateBudgetReservationStatus(ctx, resID, "RELEASED")
			s.activeReservations[emp.ID] -= estimatedRunCost
			if s.activeReservations[emp.ID] < 0 {
				s.activeReservations[emp.ID] = 0
			}
			s.budgetMu.Unlock()
		}
	}()

	// Select execution steps strictly respecting agent's effective grants
	type plannedStep struct {
		name   string
		tool   string
		input  string
		output string
		cost   float64
		tokens int64
	}
	var plannedSteps []plannedStep

	// Choose tools based ONLY on agent's granted capabilities
	if effectiveMap["test-mcp"] {
		plannedSteps = append(plannedSteps, plannedStep{
			name:   "Execute test verification suite",
			tool:   "test-mcp",
			input:  fmt.Sprintf(`{"testScope":"%s","criteria":"%s"}`, task.Title, task.AcceptanceCriteria),
			output: `{"passed":true,"coverage":96.4}`,
			cost:   0.08,
			tokens: 380,
		})
	} else if effectiveMap["task-mcp"] {
		plannedSteps = append(plannedSteps, plannedStep{
			name:   "Inspect context & criteria",
			tool:   "task-mcp",
			input:  fmt.Sprintf(`{"taskId":"%s","criteria":"%s"}`, task.ID, task.AcceptanceCriteria),
			output: `{"status":"valid","criteriaMatched":true}`,
			cost:   0.08,
			tokens: 420,
		})
	}

	// LLM Reasoning step using assigned model
	modelName := emp.Model
	if modelName == "" || modelName == "Human" {
		modelName = "Model Y"
	}
	plannedSteps = append(plannedSteps, plannedStep{
		name:   "Execute core skill reasoning",
		tool:   modelName,
		input:  fmt.Sprintf(`{"prompt":"Solve %s for project %s"}`, task.Title, task.ProjectCode),
		output: `{"result":"Solution formulated with comprehensive validation"}`,
		cost:   0.24,
		tokens: 1250,
	})

	// Final verification/artifact step with allowed tool
	if effectiveMap["git-mcp"] {
		plannedSteps = append(plannedSteps, plannedStep{
			name:   "Commit results to repository",
			tool:   "git-mcp",
			input:  fmt.Sprintf(`{"branch":"feature/%s","message":"Implement %s"}`, strings.ToLower(task.ID), task.Title),
			output: `{"commitSha":"a1b2c3d4","status":"pushed"}`,
			cost:   0.05,
			tokens: 220,
		})
	} else if effectiveMap["docs-mcp"] {
		plannedSteps = append(plannedSteps, plannedStep{
			name:   "Record evidence & artifact",
			tool:   "docs-mcp",
			input:  `{"format":"markdown","status":"verified"}`,
			output: `{"docId":"doc-evidence-final","published":true}`,
			cost:   0.05,
			tokens: 280,
		})
	}

	var totalCost float64
	var totalTokens int64

	for i, stepInfo := range plannedSteps {
		stepTime := time.Now()
		step := domain.ExecutionStep{
			StepNo:     i + 1,
			StepName:   stepInfo.name,
			ToolName:   stepInfo.tool,
			Input:      stepInfo.input,
			Output:     stepInfo.output,
			TokensUsed: stepInfo.tokens,
			CostUSD:    stepInfo.cost,
			Status:     "SUCCEEDED",
			Timestamp:  stepTime,
		}
		totalCost += stepInfo.cost
		totalTokens += stepInfo.tokens
		exec.Steps = append(exec.Steps, step)
		exec.Logs = append(exec.Logs, fmt.Sprintf("[%s] Step %d completed: %s via %s (cost: $%.2f)",
			stepTime.Format("15:04:05"), i+1, stepInfo.name, stepInfo.tool, stepInfo.cost))
	}

	endTime := time.Now()
	exec.Status = domain.ExecutionStateSucceeded
	exec.EndedAt = &endTime
	exec.TotalCostUSD = totalCost
	exec.TotalTokens = totalTokens

	// Atomic delta commit: marks reservation committed and applies delta spend to employee and task
	if err := s.Repo.CommitExecutionSpend(ctx, resID, emp.ID, task.ID, totalCost, domain.TaskStateInReview); err != nil {
		return nil, fmt.Errorf("failed to commit execution spend: %w", err)
	}

	s.budgetMu.Lock()
	s.committedSpend[emp.ID] += totalCost
	s.activeReservations[emp.ID] -= estimatedRunCost
	if s.activeReservations[emp.ID] < 0 {
		s.activeReservations[emp.ID] = 0
	}
	hasReserved = false
	s.budgetMu.Unlock()

	emp.SpentUSD += totalCost
	task.CostUSD += totalCost
	task.ExecutionsCount++
	task.Status = domain.TaskStateInReview

	if err := s.Repo.UpdateExecution(ctx, exec); err != nil {
		return nil, fmt.Errorf("failed to persist execution: %w", err)
	}

	// Record immutable steps into usage_ledger
	for i, st := range exec.Steps {
		entry := &domain.UsageLedgerEntry{
			ID:              fmt.Sprintf("led-%s-%d", exec.ID, i+1),
			OrganizationID:  task.OrganizationID,
			ExecutionID:     exec.ID,
			TaskID:          task.ID,
			ProjectID:       task.ProjectID,
			EmployeeID:      emp.ID,
			ChargeReference: fmt.Sprintf("step-%d-%s", st.StepNo, st.ToolName),
			EntryType:       "INFERENCE_STEP",
			PostedAt:        st.Timestamp,
			InputTokens:     st.TokensUsed,
			OutputTokens:    0,
			CostUSD:         st.CostUSD,
		}
		if err := s.Repo.RecordUsageLedger(ctx, entry); err != nil {
			return nil, fmt.Errorf("failed to record usage ledger: %w", err)
		}
	}

	s.Broker.Publish("execution", "completed", exec)
	s.Broker.Publish("task", "status_changed", map[string]interface{}{"taskId": task.ID, "newStatus": task.Status})

	return exec, nil
}

func (s *Services) GetTaskExecutions(ctx context.Context, taskID string) ([]domain.Execution, error) {
	return s.Repo.GetExecutionsByTaskID(ctx, taskID)
}

func (s *Services) GetEmployeeExecutions(ctx context.Context, employeeID string) ([]domain.Execution, error) {
	return s.Repo.GetExecutionsByEmployeeID(ctx, employeeID)
}

// ----------------- Approvals -----------------

func (s *Services) GetApprovals(ctx context.Context) ([]domain.ApprovalRequest, error) {
	return s.Repo.GetApprovals(ctx)
}

func (s *Services) DecideApproval(ctx context.Context, approvalID, decision, resolvedBy string) error {
	app, err := s.Repo.GetApprovalByID(ctx, approvalID)
	if err != nil {
		return err
	}

	// BR-AUTH-001: No self-approval!
	if resolvedBy != "" && resolvedBy == app.RequestedByEmployeeID {
		return fmt.Errorf("BR-AUTH-001 violation: requester cannot self-approve escalation")
	}

	now := time.Now()
	if strings.ToUpper(decision) == "APPROVE" || strings.ToUpper(decision) == "APPROVED" {
		app.Status = domain.ApprovalStateApproved
	} else {
		app.Status = domain.ApprovalStateRejected
	}
	app.ResolvedByEmployeeID = &resolvedBy
	app.ResolvedAt = &now

	if err := s.Repo.UpdateApproval(ctx, app); err != nil {
		return err
	}

	// If linked to a blocked task, unblock it when approved!
	if app.TaskID != nil && *app.TaskID != "" && app.Status == domain.ApprovalStateApproved {
		task, err := s.Repo.GetTaskByID(ctx, *app.TaskID)
		if err == nil && task.Status == domain.TaskStateBlocked {
			task.Status = domain.TaskStateReady
			_ = s.Repo.UpdateTask(ctx, task)
			s.Broker.Publish("task", "status_changed", map[string]interface{}{"taskId": task.ID, "newStatus": task.Status})
		}
	}

	s.Broker.Publish("approval", "decided", app)
	return nil
}

// ----------------- Capabilities & Board -----------------

func (s *Services) GetMCPServers(ctx context.Context) ([]domain.MCPServer, error) {
	return s.Repo.GetMCPServers(ctx)
}

func (s *Services) GetModels(ctx context.Context) ([]domain.ModelInfo, error) {
	return s.Repo.GetModels(ctx)
}

func (s *Services) GetBoardSummary(ctx context.Context, scope string) (*domain.BoardSummary, error) {
	return s.Repo.GetBoardSummary(ctx, scope)
}

// ----------------- Organization, Department & Team Management -----------------

func (s *Services) CreateDepartment(ctx context.Context, d *domain.Department) error {
	if d.Name == "" {
		return fmt.Errorf("department name is required")
	}
	if err := s.Repo.CreateDepartment(ctx, d); err != nil {
		return err
	}
	s.Broker.Publish("department", "created", d)
	return nil
}

func (s *Services) UpdateDepartment(ctx context.Context, d *domain.Department) error {
	if d.ID == "" {
		return fmt.Errorf("department ID is required")
	}
	if err := s.Repo.UpdateDepartment(ctx, d); err != nil {
		return err
	}
	s.Broker.Publish("department", "updated", d)
	return nil
}

func (s *Services) DeleteDepartment(ctx context.Context, id string) error {
	if err := s.Repo.DeleteDepartment(ctx, id); err != nil {
		return err
	}
	s.Broker.Publish("department", "deleted", map[string]string{"id": id})
	return nil
}

func (s *Services) CreateTeam(ctx context.Context, t *domain.Team) error {
	if t.Name == "" {
		return fmt.Errorf("team name is required")
	}
	if err := s.Repo.CreateTeam(ctx, t); err != nil {
		return err
	}
	s.Broker.Publish("team", "created", t)
	return nil
}

func (s *Services) UpdateTeam(ctx context.Context, t *domain.Team) error {
	if t.ID == "" {
		return fmt.Errorf("team ID is required")
	}
	if err := s.Repo.UpdateTeam(ctx, t); err != nil {
		return err
	}
	s.Broker.Publish("team", "updated", t)
	return nil
}

func (s *Services) DeleteTeam(ctx context.Context, id string) error {
	if err := s.Repo.DeleteTeam(ctx, id); err != nil {
		return err
	}
	s.Broker.Publish("team", "deleted", map[string]string{"id": id})
	return nil
}

// ----------------- Employee & Hierarchy Management -----------------

func (s *Services) UpdateEmployee(ctx context.Context, emp *domain.Employee) error {
	if emp.ID == "" {
		return fmt.Errorf("employee ID is required")
	}
	if err := s.Repo.UpdateEmployee(ctx, emp); err != nil {
		return err
	}
	s.Broker.Publish("employee", "updated", emp)
	return nil
}

func (s *Services) DeleteEmployee(ctx context.Context, id string) error {
	if err := s.Repo.DeleteEmployee(ctx, id); err != nil {
		return err
	}
	s.Broker.Publish("employee", "deleted", map[string]string{"id": id})
	return nil
}

func (s *Services) UpdateReportingLine(ctx context.Context, employeeID string, managerID *string) error {
	if employeeID == "" {
		return fmt.Errorf("employeeId cannot be empty")
	}

	emp, err := s.Repo.GetEmployeeByID(ctx, employeeID)
	if err != nil || emp == nil {
		return fmt.Errorf("employee %s not found", employeeID)
	}

	if managerID == nil || *managerID == "" {
		// Clearing manager (e.g. root founder)
		if err := s.Repo.UpdateReportingLine(ctx, employeeID, nil); err != nil {
			return err
		}
		s.Broker.Publish("hierarchy", "reassigned", map[string]interface{}{
			"employeeId": employeeID,
			"managerId":  nil,
		})
		return nil
	}

	targetMgrID := *managerID
	// BR-ORG-002: cannot manage themselves
	if targetMgrID == employeeID {
		return fmt.Errorf("BR-ORG-002 violation: employee cannot manage themselves")
	}

	// BR-ORG-001: manager must exist
	targetMgr, err := s.Repo.GetEmployeeByID(ctx, targetMgrID)
	if err != nil || targetMgr == nil {
		return fmt.Errorf("BR-ORG-001 violation: manager %s does not exist", targetMgrID)
	}

	// Cycle detection: traverse up from targetMgrID to see if employeeID is an ancestor
	employees, err := s.Repo.GetEmployees(ctx)
	if err != nil {
		return err
	}
	empMap := make(map[string]domain.Employee)
	for _, e := range employees {
		empMap[e.ID] = e
	}

	curr := targetMgrID
	visited := make(map[string]bool)
	for curr != "" {
		if curr == employeeID {
			return fmt.Errorf("BR-ORG-002 violation: circular reporting chain detected (%s is an ancestor of %s)", employeeID, targetMgrID)
		}
		if visited[curr] {
			break
		}
		visited[curr] = true

		m, exists := empMap[curr]
		if !exists || m.ManagerID == nil {
			break
		}
		curr = *m.ManagerID
	}

	if err := s.Repo.UpdateReportingLine(ctx, employeeID, managerID); err != nil {
		return err
	}

	s.Broker.Publish("hierarchy", "reassigned", map[string]interface{}{
		"employeeId": employeeID,
		"managerId":  targetMgrID,
	})
	return nil
}

// ----------------- Project Management -----------------

func (s *Services) UpdateProject(ctx context.Context, p *domain.Project) error {
	if p.ID == "" {
		return fmt.Errorf("project ID is required")
	}
	if err := s.Repo.UpdateProject(ctx, p); err != nil {
		return err
	}
	s.Broker.Publish("project", "updated", p)
	return nil
}

func (s *Services) DeleteProject(ctx context.Context, id string) error {
	if err := s.Repo.DeleteProject(ctx, id); err != nil {
		return err
	}
	s.Broker.Publish("project", "deleted", map[string]string{"id": id})
	return nil
}

// ----------------- Capabilities (MCP & Models) Management -----------------

func (s *Services) CreateMCPServer(ctx context.Context, server *domain.MCPServer) error {
	if server.Name == "" {
		return fmt.Errorf("mcp server name is required")
	}
	if err := s.Repo.CreateMCPServer(ctx, server); err != nil {
		return err
	}
	s.Broker.Publish("mcp_server", "created", server)
	return nil
}

func (s *Services) UpdateMCPServer(ctx context.Context, server *domain.MCPServer) error {
	if server.ID == "" {
		return fmt.Errorf("mcp server ID is required")
	}
	if err := s.Repo.UpdateMCPServer(ctx, server); err != nil {
		return err
	}
	s.Broker.Publish("mcp_server", "updated", server)
	return nil
}

func (s *Services) DeleteMCPServer(ctx context.Context, id string) error {
	if err := s.Repo.DeleteMCPServer(ctx, id); err != nil {
		return err
	}
	s.Broker.Publish("mcp_server", "deleted", map[string]string{"id": id})
	return nil
}

func (s *Services) CreateModel(ctx context.Context, model *domain.ModelInfo) error {
	if model.ProviderCode == "" || model.ModelID == "" {
		return fmt.Errorf("provider and modelId are required")
	}
	if err := s.Repo.CreateModel(ctx, model); err != nil {
		return err
	}
	s.Broker.Publish("model", "created", model)
	return nil
}

func (s *Services) UpdateModel(ctx context.Context, model *domain.ModelInfo) error {
	if model.ID == "" {
		return fmt.Errorf("model ID is required")
	}
	if err := s.Repo.UpdateModel(ctx, model); err != nil {
		return err
	}
	s.Broker.Publish("model", "updated", model)
	return nil
}

func (s *Services) DeleteModel(ctx context.Context, id string) error {
	if err := s.Repo.DeleteModel(ctx, id); err != nil {
		return err
	}
	s.Broker.Publish("model", "deleted", map[string]string{"id": id})
	return nil
}
