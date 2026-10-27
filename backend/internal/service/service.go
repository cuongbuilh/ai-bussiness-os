package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cuongbuilh/ai-business-os/backend/internal/domain"
	"github.com/cuongbuilh/ai-business-os/backend/internal/repository"
	"github.com/google/uuid"
)

type Services struct {
	Repo   repository.Repository
	Broker *EventBroker
}

func NewServices(repo repository.Repository, broker *EventBroker) *Services {
	return &Services{
		Repo:   repo,
		Broker: broker,
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
	if emp.ManagerID != nil && *emp.ManagerID == emp.ID {
		return fmt.Errorf("BR-ORG-002 violation: employee cannot manage themselves")
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

	if err := s.Repo.CreateTask(ctx, task); err != nil {
		return err
	}

	s.Broker.Publish("task", "created", task)
	return nil
}

func (s *Services) UpdateTaskStatus(ctx context.Context, taskID string, newStatus domain.TaskState) error {
	task, err := s.Repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
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

	// BR-BUD-003: Check budget limit
	if emp.BudgetUSD > 0 && emp.SpentUSD >= emp.BudgetUSD {
		return nil, fmt.Errorf("BR-BUD-003 violation: monthly budget cap reached ($%.2f / $%.2f)", emp.SpentUSD, emp.BudgetUSD)
	}

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

	// Update task status to in_progress
	task.Status = domain.TaskStateInProgress
	task.ExecutionsCount++
	_ = s.Repo.UpdateTask(ctx, task)
	s.Broker.Publish("task", "status_changed", map[string]interface{}{"taskId": task.ID, "newStatus": task.Status})

	// Generate step simulation according to agent skills & MCP
	steps := []struct {
		name   string
		tool   string
		input  string
		output string
		cost   float64
		tokens int64
	}{
		{
			name:   "Inspect context & criteria",
			tool:   "task-mcp",
			input:  fmt.Sprintf(`{"taskId":"%s","criteria":"%s"}`, task.ID, task.AcceptanceCriteria),
			output: `{"status":"valid","criteriaMatched":true}`,
			cost:   0.08,
			tokens: 420,
		},
		{
			name:   "Execute core skill procedure",
			tool:   emp.Model,
			input:  fmt.Sprintf(`{"prompt":"Solve %s for project %s"}`, task.Title, task.ProjectCode),
			output: `{"result":"Solution formulated with comprehensive validation"}`,
			cost:   0.24,
			tokens: 1250,
		},
		{
			name:   "Record evidence & artifact",
			tool:   "docs-mcp",
			input:  `{"format":"markdown","status":"verified"}`,
			output: `{"docId":"doc-evidence-final","published":true}`,
			cost:   0.05,
			tokens: 280,
		},
	}

	var totalCost float64
	var totalTokens int64

	for i, stepInfo := range steps {
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

	// Settle ledger & update employee spend and task cost
	emp.SpentUSD += totalCost
	_ = s.Repo.UpdateEmployee(ctx, emp)

	task.CostUSD += totalCost
	task.Status = domain.TaskStateInReview // Ready for review / completion!
	_ = s.Repo.UpdateTask(ctx, task)

	_ = s.Repo.CreateExecution(ctx, exec)

	s.Broker.Publish("execution", "completed", exec)
	s.Broker.Publish("task", "status_changed", map[string]interface{}{"taskId": task.ID, "newStatus": task.Status})

	return exec, nil
}

func (s *Services) GetTaskExecutions(ctx context.Context, taskID string) ([]domain.Execution, error) {
	return s.Repo.GetExecutionsByTaskID(ctx, taskID)
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
