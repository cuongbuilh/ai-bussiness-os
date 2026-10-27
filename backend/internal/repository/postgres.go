package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cuongbuilh/ai-business-os/backend/internal/domain"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(databaseURL string) (*PostgresRepo, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// Retry connection for up to 30 seconds (helps Docker compose startup order)
	var pingErr error
	for i := 0; i < 15; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if pingErr != nil {
		return nil, fmt.Errorf("failed to ping postgres after retries: %w", pingErr)
	}

	repo := &PostgresRepo{db: db}
	if err := repo.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return repo, nil
}

func (r *PostgresRepo) initSchema() error {
	initFiles := []string{
		"migrations/001_init.sql",
		"migrations/002_seed.sql",
	}

	for _, file := range initFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			// Fallback check if running from another directory
			content, err = os.ReadFile("../" + file)
			if err != nil {
				continue
			}
		}
		if _, err := r.db.Exec(string(content)); err != nil {
			// Ignore seed conflict or proceed
			fmt.Printf("Notice during migration (%s): %v\n", file, err)
		}
	}
	return nil
}

func (r *PostgresRepo) GetDepartments(ctx context.Context) ([]domain.Department, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, organization_id, name, parent_department_id, created_at FROM departments")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Department
	for rows.Next() {
		var d domain.Department
		if err := rows.Scan(&d.ID, &d.OrganizationID, &d.Name, &d.ParentDepartmentID, &d.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, nil
}

func (r *PostgresRepo) GetTeams(ctx context.Context) ([]domain.Team, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, organization_id, department_id, name, lead_employee_id, created_at FROM teams")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Team
	for rows.Next() {
		var t domain.Team
		if err := rows.Scan(&t.ID, &t.OrganizationID, &t.DepartmentID, &t.Name, &t.LeadEmployeeID, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *PostgresRepo) GetPositions(ctx context.Context) ([]domain.Position, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, organization_id, department_id, title, role_code, description, cap_mcps FROM positions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Position
	for rows.Next() {
		var p domain.Position
		var capBytes []byte
		if err := rows.Scan(&p.ID, &p.OrganizationID, &p.DepartmentID, &p.Title, &p.RoleCode, &p.Description, &capBytes); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(capBytes, &p.CapMCPs)
		list = append(list, p)
	}
	return list, nil
}

func (r *PostgresRepo) GetEmployees(ctx context.Context) ([]domain.Employee, error) {
	query := `
		SELECT e.id, e.organization_id, e.department_id, e.position_id, e.display_name, e.kind, e.state,
		       e.title, e.role_code, e.manager_id, COALESCE(m.display_name, 'None / Founder'),
		       e.model_name, e.provider_name, e.skills, e.mcp_grants, e.team_ids, e.kpi_score,
		       e.budget_usd, e.spent_usd, e.version, e.created_at, COALESCE(d.name, '')
		FROM employees e
		LEFT JOIN employees m ON e.manager_id = m.id
		LEFT JOIN departments d ON e.department_id = d.id
		ORDER BY e.created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Employee
	for rows.Next() {
		var e domain.Employee
		var skillsBytes, mcpBytes, teamBytes []byte
		err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.DepartmentID, &e.PositionID, &e.DisplayName, &e.Kind, &e.State,
			&e.Title, &e.Role, &e.ManagerID, &e.ManagerName,
			&e.Model, &e.Provider, &skillsBytes, &mcpBytes, &teamBytes, &e.KPI,
			&e.BudgetUSD, &e.SpentUSD, &e.Version, &e.CreatedAt, &e.DepartmentName,
		)
		if err != nil {
			return nil, err
		}
		e.Name = e.DisplayName
		_ = json.Unmarshal(skillsBytes, &e.Skills)
		_ = json.Unmarshal(mcpBytes, &e.MCP)
		_ = json.Unmarshal(teamBytes, &e.Teams)
		list = append(list, e)
	}
	return list, nil
}

func (r *PostgresRepo) GetEmployeeByID(ctx context.Context, id string) (*domain.Employee, error) {
	query := `
		SELECT e.id, e.organization_id, e.department_id, e.position_id, e.display_name, e.kind, e.state,
		       e.title, e.role_code, e.manager_id, COALESCE(m.display_name, 'None / Founder'),
		       e.model_name, e.provider_name, e.skills, e.mcp_grants, e.team_ids, e.kpi_score,
		       e.budget_usd, e.spent_usd, e.version, e.created_at, COALESCE(d.name, '')
		FROM employees e
		LEFT JOIN employees m ON e.manager_id = m.id
		LEFT JOIN departments d ON e.department_id = d.id
		WHERE e.id = $1
	`
	var e domain.Employee
	var skillsBytes, mcpBytes, teamBytes []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID, &e.OrganizationID, &e.DepartmentID, &e.PositionID, &e.DisplayName, &e.Kind, &e.State,
		&e.Title, &e.Role, &e.ManagerID, &e.ManagerName,
		&e.Model, &e.Provider, &skillsBytes, &mcpBytes, &teamBytes, &e.KPI,
		&e.BudgetUSD, &e.SpentUSD, &e.Version, &e.CreatedAt, &e.DepartmentName,
	)
	if err != nil {
		return nil, err
	}
	e.Name = e.DisplayName
	_ = json.Unmarshal(skillsBytes, &e.Skills)
	_ = json.Unmarshal(mcpBytes, &e.MCP)
	_ = json.Unmarshal(teamBytes, &e.Teams)
	return &e, nil
}

func (r *PostgresRepo) CreateEmployee(ctx context.Context, emp *domain.Employee) error {
	skillsJSON, _ := json.Marshal(emp.Skills)
	mcpJSON, _ := json.Marshal(emp.MCP)
	teamJSON, _ := json.Marshal(emp.Teams)

	query := `
		INSERT INTO employees (
			id, organization_id, department_id, position_id, display_name, kind, state,
			title, role_code, manager_id, model_name, provider_name, skills, mcp_grants,
			team_ids, kpi_score, budget_usd, spent_usd, version, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, now())
	`
	_, err := r.db.ExecContext(ctx, query,
		emp.ID, "org-default", emp.DepartmentID, emp.PositionID, emp.DisplayName, emp.Kind, emp.State,
		emp.Title, emp.Role, emp.ManagerID, emp.Model, emp.Provider, skillsJSON, mcpJSON,
		teamJSON, emp.KPI, emp.BudgetUSD, emp.SpentUSD, 1,
	)
	return err
}

func (r *PostgresRepo) UpdateEmployee(ctx context.Context, emp *domain.Employee) error {
	skillsJSON, _ := json.Marshal(emp.Skills)
	mcpJSON, _ := json.Marshal(emp.MCP)
	teamJSON, _ := json.Marshal(emp.Teams)

	query := `
		UPDATE employees SET
			display_name = $1, title = $2, role_code = $3, manager_id = $4,
			model_name = $5, provider_name = $6, skills = $7, mcp_grants = $8,
			team_ids = $9, kpi_score = $10, budget_usd = $11, spent_usd = $12,
			version = version + 1
		WHERE id = $13
	`
	_, err := r.db.ExecContext(ctx, query,
		emp.DisplayName, emp.Title, emp.Role, emp.ManagerID,
		emp.Model, emp.Provider, skillsJSON, mcpJSON,
		teamJSON, emp.KPI, emp.BudgetUSD, emp.SpentUSD, emp.ID,
	)
	return err
}

func (r *PostgresRepo) GetOrgTree(ctx context.Context) (*domain.OrgNode, error) {
	employees, err := r.GetEmployees(ctx)
	if err != nil {
		return nil, err
	}

	var root *domain.Employee
	for _, e := range employees {
		if e.ManagerID == nil {
			copy := e
			root = &copy
			break
		}
	}
	if root == nil && len(employees) > 0 {
		copy := employees[0]
		root = &copy
	}
	if root == nil {
		return nil, fmt.Errorf("no employees found")
	}

	var buildNode func(e *domain.Employee) *domain.OrgNode
	buildNode = func(e *domain.Employee) *domain.OrgNode {
		node := &domain.OrgNode{
			ID:         e.ID,
			Name:       e.DisplayName,
			Title:      e.Title,
			Role:       e.Role,
			Kind:       string(e.Kind),
			Model:      e.Model,
			Provider:   e.Provider,
			Department: e.DepartmentName,
			DirectSub:  make([]*domain.OrgNode, 0),
		}
		for _, child := range employees {
			if child.ManagerID != nil && *child.ManagerID == e.ID {
				childCopy := child
				node.DirectSub = append(node.DirectSub, buildNode(&childCopy))
			}
		}
		return node
	}
	return buildNode(root), nil
}

func (r *PostgresRepo) GetProjects(ctx context.Context) ([]domain.Project, error) {
	query := `
		SELECT p.id, p.organization_id, p.code, p.name, p.description, p.accountable_owner_id,
		       COALESCE(e.display_name, ''), p.status, p.budget_usd, p.spent_usd, p.due_at,
		       p.members, p.teams, p.milestones, p.created_at, p.version
		FROM projects p
		LEFT JOIN employees e ON p.accountable_owner_id = e.id
		ORDER BY p.created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Project
	for rows.Next() {
		var p domain.Project
		var memBytes, teamBytes, mileBytes []byte
		err := rows.Scan(
			&p.ID, &p.OrganizationID, &p.Code, &p.Name, &p.Description, &p.AccountableOwnerID,
			&p.OwnerName, &p.Status, &p.BudgetUSD, &p.SpentUSD, &p.DueAt,
			&memBytes, &teamBytes, &mileBytes, &p.CreatedAt, &p.Version,
		)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(memBytes, &p.Members)
		_ = json.Unmarshal(teamBytes, &p.Teams)
		_ = json.Unmarshal(mileBytes, &p.Milestones)
		list = append(list, p)
	}
	return list, nil
}

func (r *PostgresRepo) GetProjectByID(ctx context.Context, id string) (*domain.Project, error) {
	query := `
		SELECT p.id, p.organization_id, p.code, p.name, p.description, p.accountable_owner_id,
		       COALESCE(e.display_name, ''), p.status, p.budget_usd, p.spent_usd, p.due_at,
		       p.members, p.teams, p.milestones, p.created_at, p.version
		FROM projects p
		LEFT JOIN employees e ON p.accountable_owner_id = e.id
		WHERE p.id = $1
	`
	var p domain.Project
	var memBytes, teamBytes, mileBytes []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.OrganizationID, &p.Code, &p.Name, &p.Description, &p.AccountableOwnerID,
		&p.OwnerName, &p.Status, &p.BudgetUSD, &p.SpentUSD, &p.DueAt,
		&memBytes, &teamBytes, &mileBytes, &p.CreatedAt, &p.Version,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(memBytes, &p.Members)
	_ = json.Unmarshal(teamBytes, &p.Teams)
	_ = json.Unmarshal(mileBytes, &p.Milestones)
	return &p, nil
}

func (r *PostgresRepo) CreateProject(ctx context.Context, p *domain.Project) error {
	memJSON, _ := json.Marshal(p.Members)
	teamJSON, _ := json.Marshal(p.Teams)
	mileJSON, _ := json.Marshal(p.Milestones)

	query := `
		INSERT INTO projects (
			id, organization_id, code, name, description, accountable_owner_id,
			status, budget_usd, spent_usd, due_at, members, teams, milestones, created_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, now(), 1)
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, "org-default", p.Code, p.Name, p.Description, p.AccountableOwnerID,
		p.Status, p.BudgetUSD, p.SpentUSD, p.DueAt, memJSON, teamJSON, mileJSON,
	)
	return err
}

func (r *PostgresRepo) UpdateProject(ctx context.Context, p *domain.Project) error {
	memJSON, _ := json.Marshal(p.Members)
	teamJSON, _ := json.Marshal(p.Teams)
	mileJSON, _ := json.Marshal(p.Milestones)

	query := `
		UPDATE projects SET
			code = $1, name = $2, description = $3, accountable_owner_id = $4,
			status = $5, budget_usd = $6, spent_usd = $7, due_at = $8,
			members = $9, teams = $10, milestones = $11, version = version + 1
		WHERE id = $12
	`
	_, err := r.db.ExecContext(ctx, query,
		p.Code, p.Name, p.Description, p.AccountableOwnerID,
		p.Status, p.BudgetUSD, p.SpentUSD, p.DueAt,
		memJSON, teamJSON, mileJSON, p.ID,
	)
	return err
}

func (r *PostgresRepo) GetTasks(ctx context.Context, scope, projectFilter, ownerFilter, statusFilter, search string) ([]domain.Task, error) {
	query := `
		SELECT t.id, t.organization_id, t.project_id, COALESCE(p.code, 'Operations'), COALESCE(p.name, ''),
		       t.parent_task_id, t.title, t.description, t.accountable_owner_id, COALESCE(e.display_name, ''),
		       t.collaborators, t.status, t.priority, t.acceptance_criteria, t.due_at, t.accepted_at,
		       t.cost_usd, t.execs_count, t.origin_type, t.origin_ref, t.created_at, t.updated_at, t.version
		FROM tasks t
		LEFT JOIN projects p ON t.project_id = p.id
		LEFT JOIN employees e ON t.accountable_owner_id = e.id
		ORDER BY t.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allTasks []domain.Task
	for rows.Next() {
		var t domain.Task
		var collabBytes []byte
		err := rows.Scan(
			&t.ID, &t.OrganizationID, &t.ProjectID, &t.ProjectCode, &t.ProjectName,
			&t.ParentTaskID, &t.Title, &t.Description, &t.AccountableOwnerID, &t.OwnerName,
			&collabBytes, &t.Status, &t.Priority, &t.AcceptanceCriteria, &t.DueAt, &t.AcceptedAt,
			&t.CostUSD, &t.ExecutionsCount, &t.OriginType, &t.OriginRef, &t.CreatedAt, &t.UpdatedAt, &t.Version,
		)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(collabBytes, &t.Collaborators)
		allTasks = append(allTasks, t)
	}

	// Apply scope and filters
	employees, _ := r.GetEmployees(ctx)
	searchLower := strings.ToLower(search)
	var filtered []domain.Task
	for _, t := range allTasks {
		if !r.inScope(ctx, &t, scope, employees) {
			continue
		}
		if projectFilter != "" && (t.ProjectID == nil || *t.ProjectID != projectFilter) {
			continue
		}
		if ownerFilter != "" {
			matched := t.AccountableOwnerID == ownerFilter
			if !matched {
				for _, c := range t.Collaborators {
					if c == ownerFilter {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
		}
		if statusFilter != "" && string(t.Status) != statusFilter {
			continue
		}
		if searchLower != "" {
			txt := strings.ToLower(fmt.Sprintf("%s %s %s %s", t.ID, t.Title, t.ProjectCode, t.OwnerName))
			if !strings.Contains(txt, searchLower) {
				continue
			}
		}
		filtered = append(filtered, t)
	}
	return filtered, nil
}

func (r *PostgresRepo) GetTaskByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `
		SELECT t.id, t.organization_id, t.project_id, COALESCE(p.code, 'Operations'), COALESCE(p.name, ''),
		       t.parent_task_id, t.title, t.description, t.accountable_owner_id, COALESCE(e.display_name, ''),
		       t.collaborators, t.status, t.priority, t.acceptance_criteria, t.due_at, t.accepted_at,
		       t.cost_usd, t.execs_count, t.origin_type, t.origin_ref, t.created_at, t.updated_at, t.version
		FROM tasks t
		LEFT JOIN projects p ON t.project_id = p.id
		LEFT JOIN employees e ON t.accountable_owner_id = e.id
		WHERE t.id = $1
	`
	var t domain.Task
	var collabBytes []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.OrganizationID, &t.ProjectID, &t.ProjectCode, &t.ProjectName,
		&t.ParentTaskID, &t.Title, &t.Description, &t.AccountableOwnerID, &t.OwnerName,
		&collabBytes, &t.Status, &t.Priority, &t.AcceptanceCriteria, &t.DueAt, &t.AcceptedAt,
		&t.CostUSD, &t.ExecutionsCount, &t.OriginType, &t.OriginRef, &t.CreatedAt, &t.UpdatedAt, &t.Version,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(collabBytes, &t.Collaborators)
	return &t, nil
}

func (r *PostgresRepo) CreateTask(ctx context.Context, task *domain.Task) error {
	collabJSON, _ := json.Marshal(task.Collaborators)
	if task.ID == "" {
		task.ID = fmt.Sprintf("T-%s", strings.ToUpper(uuid.New().String()[:8]))
	}
	if task.OrganizationID == "" {
		task.OrganizationID = "org-default"
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = now
	}
	if task.Version == 0 {
		task.Version = 1
	}

	query := `
		INSERT INTO tasks (
			id, organization_id, project_id, parent_task_id, title, description,
			accountable_owner_id, collaborators, status, priority, acceptance_criteria,
			due_at, cost_usd, execs_count, origin_type, origin_ref, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`
	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.OrganizationID, task.ProjectID, task.ParentTaskID, task.Title, task.Description,
		task.AccountableOwnerID, collabJSON, task.Status, task.Priority, task.AcceptanceCriteria,
		task.DueAt, task.CostUSD, task.ExecutionsCount, task.OriginType, task.OriginRef,
		task.CreatedAt, task.UpdatedAt, task.Version,
	)
	return err
}

func (r *PostgresRepo) UpdateTask(ctx context.Context, task *domain.Task) error {
	collabJSON, _ := json.Marshal(task.Collaborators)
	query := `
		UPDATE tasks SET
			project_id = $1, title = $2, description = $3, accountable_owner_id = $4,
			collaborators = $5, status = $6, priority = $7, acceptance_criteria = $8,
			due_at = $9, accepted_at = $10, cost_usd = $11, execs_count = $12,
			updated_at = now(), version = version + 1
		WHERE id = $13
	`
	_, err := r.db.ExecContext(ctx, query,
		task.ProjectID, task.Title, task.Description, task.AccountableOwnerID,
		collabJSON, task.Status, task.Priority, task.AcceptanceCriteria,
		task.DueAt, task.AcceptedAt, task.CostUSD, task.ExecutionsCount, task.ID,
	)
	return err
}

func (r *PostgresRepo) GetExecutionsByTaskID(ctx context.Context, taskID string) ([]domain.Execution, error) {
	query := `
		SELECT x.id, x.organization_id, x.task_id, x.executing_employee_id, COALESCE(e.display_name, ''),
		       x.status, x.attempt_no, x.idempotency_key, x.logs, x.steps, x.total_tokens, x.total_cost_usd,
		       x.created_at, x.started_at, x.ended_at
		FROM executions x
		LEFT JOIN employees e ON x.executing_employee_id = e.id
		WHERE x.task_id = $1
		ORDER BY x.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]domain.Execution, 0)
	for rows.Next() {
		var x domain.Execution
		var logsBytes, stepsBytes []byte
		err := rows.Scan(
			&x.ID, &x.OrganizationID, &x.TaskID, &x.ExecutingEmployeeID, &x.ExecutingEmpName,
			&x.Status, &x.AttemptNo, &x.IdempotencyKey, &logsBytes, &stepsBytes,
			&x.TotalTokens, &x.TotalCostUSD, &x.CreatedAt, &x.StartedAt, &x.EndedAt,
		)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(logsBytes, &x.Logs)
		_ = json.Unmarshal(stepsBytes, &x.Steps)
		list = append(list, x)
	}
	return list, nil
}

func (r *PostgresRepo) GetExecutionsByEmployeeID(ctx context.Context, empID string) ([]domain.Execution, error) {
	query := `
		SELECT x.id, x.organization_id, x.task_id, x.executing_employee_id, COALESCE(e.display_name, ''),
		       x.status, x.attempt_no, x.idempotency_key, x.logs, x.steps, x.total_tokens, x.total_cost_usd,
		       x.created_at, x.started_at, x.ended_at
		FROM executions x
		LEFT JOIN employees e ON x.executing_employee_id = e.id
		WHERE x.executing_employee_id = $1
		ORDER BY x.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, empID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]domain.Execution, 0)
	for rows.Next() {
		var x domain.Execution
		var logsBytes, stepsBytes []byte
		err := rows.Scan(
			&x.ID, &x.OrganizationID, &x.TaskID, &x.ExecutingEmployeeID, &x.ExecutingEmpName,
			&x.Status, &x.AttemptNo, &x.IdempotencyKey, &logsBytes, &stepsBytes,
			&x.TotalTokens, &x.TotalCostUSD, &x.CreatedAt, &x.StartedAt, &x.EndedAt,
		)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(logsBytes, &x.Logs)
		_ = json.Unmarshal(stepsBytes, &x.Steps)
		list = append(list, x)
	}
	return list, nil
}

func (r *PostgresRepo) CreateExecution(ctx context.Context, exec *domain.Execution) error {
	logsJSON, _ := json.Marshal(exec.Logs)
	stepsJSON, _ := json.Marshal(exec.Steps)

	query := `
		INSERT INTO executions (
			id, organization_id, task_id, executing_employee_id, status,
			attempt_no, idempotency_key, logs, steps, total_tokens, total_cost_usd,
			created_at, started_at, ended_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now(), $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		exec.ID, "org-default", exec.TaskID, exec.ExecutingEmployeeID, exec.Status,
		exec.AttemptNo, exec.IdempotencyKey, logsJSON, stepsJSON, exec.TotalTokens, exec.TotalCostUSD,
		exec.StartedAt, exec.EndedAt,
	)
	return err
}

func (r *PostgresRepo) UpdateExecution(ctx context.Context, exec *domain.Execution) error {
	logsJSON, _ := json.Marshal(exec.Logs)
	stepsJSON, _ := json.Marshal(exec.Steps)

	query := `
		UPDATE executions SET
			status = $1, logs = $2, steps = $3, total_tokens = $4,
			total_cost_usd = $5, ended_at = $6
		WHERE id = $7
	`
	_, err := r.db.ExecContext(ctx, query,
		exec.Status, logsJSON, stepsJSON, exec.TotalTokens,
		exec.TotalCostUSD, exec.EndedAt, exec.ID,
	)
	return err
}

func (r *PostgresRepo) GetApprovals(ctx context.Context) ([]domain.ApprovalRequest, error) {
	query := `
		SELECT a.id, a.organization_id, a.task_id, COALESCE(t.title, ''), a.requested_by,
		       COALESCE(e.display_name, ''), a.action_kind, a.payload_json, a.risk_tier,
		       a.status, a.title, a.expires_at, a.resolved_by, a.resolved_at, a.created_at
		FROM approval_requests a
		LEFT JOIN tasks t ON a.task_id = t.id
		LEFT JOIN employees e ON a.requested_by = e.id
		ORDER BY a.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.ApprovalRequest
	for rows.Next() {
		var a domain.ApprovalRequest
		err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.TaskID, &a.TaskTitle, &a.RequestedByEmployeeID,
			&a.RequesterName, &a.ActionKind, &a.PayloadJSON, &a.RiskTier,
			&a.Status, &a.Title, &a.ExpiresAt, &a.ResolvedByEmployeeID, &a.ResolvedAt, &a.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (r *PostgresRepo) GetApprovalByID(ctx context.Context, id string) (*domain.ApprovalRequest, error) {
	query := `
		SELECT a.id, a.organization_id, a.task_id, COALESCE(t.title, ''), a.requested_by,
		       COALESCE(e.display_name, ''), a.action_kind, a.payload_json, a.risk_tier,
		       a.status, a.title, a.expires_at, a.resolved_by, a.resolved_at, a.created_at
		FROM approval_requests a
		LEFT JOIN tasks t ON a.task_id = t.id
		LEFT JOIN employees e ON a.requested_by = e.id
		WHERE a.id = $1
	`
	var a domain.ApprovalRequest
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.OrganizationID, &a.TaskID, &a.TaskTitle, &a.RequestedByEmployeeID,
		&a.RequesterName, &a.ActionKind, &a.PayloadJSON, &a.RiskTier,
		&a.Status, &a.Title, &a.ExpiresAt, &a.ResolvedByEmployeeID, &a.ResolvedAt, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *PostgresRepo) UpdateApproval(ctx context.Context, a *domain.ApprovalRequest) error {
	query := `
		UPDATE approval_requests SET
			status = $1, resolved_by = $2, resolved_at = $3
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, a.Status, a.ResolvedByEmployeeID, a.ResolvedAt, a.ID)
	return err
}

func (r *PostgresRepo) GetMCPServers(ctx context.Context) ([]domain.MCPServer, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, organization_id, name, transport, description, risk_tier, enabled, review_status, created_at FROM mcp_servers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.MCPServer
	for rows.Next() {
		var m domain.MCPServer
		if err := rows.Scan(&m.ID, &m.OrganizationID, &m.Name, &m.Transport, &m.Description, &m.RiskTier, &m.Enabled, &m.ReviewStatus, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *PostgresRepo) GetModels(ctx context.Context) ([]domain.ModelInfo, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, provider_code, model_id, revision, features, status FROM model_versions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.ModelInfo
	for rows.Next() {
		var m domain.ModelInfo
		var featBytes []byte
		if err := rows.Scan(&m.ID, &m.ProviderCode, &m.ModelID, &m.Revision, &featBytes, &m.Status); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(featBytes, &m.Features)
		list = append(list, m)
	}
	return list, nil
}

func (r *PostgresRepo) GetBoardSummary(ctx context.Context, scope string) (*domain.BoardSummary, error) {
	tasks, err := r.GetTasks(ctx, scope, "", "", "", "")
	if err != nil {
		return nil, err
	}
	summary := &domain.BoardSummary{
		TotalTasks: len(tasks),
	}
	projects := make(map[string]bool)
	for _, t := range tasks {
		if t.Status == domain.TaskStateAccepted {
			summary.AcceptedTasks++
		}
		if t.Status == domain.TaskStateInProgress || t.Status == domain.TaskStateInReview {
			summary.ActiveTasks++
		}
		if t.Status == domain.TaskStateBlocked {
			summary.BlockedTasks++
		}
		if t.Status == domain.TaskStateBacklog {
			summary.BacklogTasks++
		}
		summary.TotalSpentUSD += t.CostUSD
		if t.ProjectID != nil && *t.ProjectID != "" {
			projects[*t.ProjectID] = true
		}
	}
	summary.ProjectsCount = len(projects)
	return summary, nil
}

func (r *PostgresRepo) inScope(ctx context.Context, t *domain.Task, scope string, employees []domain.Employee) bool {
	if scope == "" || scope == "company:all" {
		return true
	}
	parts := strings.Split(scope, ":")
	if len(parts) != 2 {
		return true
	}
	scopeType, scopeID := parts[0], parts[1]
	related := append([]string{t.AccountableOwnerID}, t.Collaborators...)

	empMap := make(map[string]domain.Employee)
	for _, e := range employees {
		empMap[e.ID] = e
	}

	switch scopeType {
	case "company":
		return true
	case "project":
		return t.ProjectID != nil && *t.ProjectID == scopeID
	case "department":
		for _, id := range related {
			if emp, ok := empMap[id]; ok && emp.DepartmentID != nil && *emp.DepartmentID == scopeID {
				return true
			}
		}
	case "team":
		for _, id := range related {
			if emp, ok := empMap[id]; ok {
				for _, tm := range emp.Teams {
					if tm == scopeID {
						return true
					}
				}
			}
		}
	case "employee":
		for _, id := range related {
			if id == scopeID {
				return true
			}
		}
	case "manager":
		sub := make(map[string]bool)
		sub[scopeID] = true
		changed := true
		for changed {
			changed = false
			for _, e := range employees {
				if e.ManagerID != nil && sub[*e.ManagerID] && !sub[e.ID] {
					sub[e.ID] = true
					changed = true
				}
			}
		}
		for _, id := range related {
			if sub[id] {
				return true
			}
		}
	case "role":
		for _, id := range related {
			if emp, ok := empMap[id]; ok && emp.Role == scopeID {
				return true
			}
		}
	}
	return false
}

func (r *PostgresRepo) AtomicReserveBudget(ctx context.Context, res *domain.BudgetReservation, employeeID string, estimatedCost float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin budget reservation transaction: %w", err)
	}
	defer tx.Rollback()

	// Row-level lock on employee to serialize budget reservation across multiple instances
	var budgetUSD, spentUSD float64
	err = tx.QueryRowContext(ctx, "SELECT budget_usd, spent_usd FROM employees WHERE id = $1 FOR UPDATE", employeeID).Scan(&budgetUSD, &spentUSD)
	if err != nil {
		return fmt.Errorf("failed to lock employee for budget check: %w", err)
	}

	// Calculate currently active reservations for this employee
	var activeReserved float64
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(br.reserved_usd), 0)
		FROM budget_reservations br
		JOIN executions e ON br.execution_id = e.id
		WHERE e.executing_employee_id = $1
		  AND br.status = 'ACTIVE'
		  AND br.expires_at > now()
	`, employeeID).Scan(&activeReserved)
	if err != nil {
		return fmt.Errorf("failed to query active reservations: %w", err)
	}

	if budgetUSD > 0 && (spentUSD+activeReserved+estimatedCost) > budgetUSD {
		return fmt.Errorf("BR-BUD-003 violation: planned run charge ($%.2f) exceeds remaining budget cap ($%.2f remaining of $%.2f cap)",
			estimatedCost, budgetUSD-(spentUSD+activeReserved), budgetUSD)
	}

	// Insert reservation within same transaction
	query := `
		INSERT INTO budget_reservations (
			id, organization_id, execution_id, step_key, reserved_usd, reserved_tokens, status, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = tx.ExecContext(ctx, query,
		res.ID, res.OrganizationID, res.ExecutionID, res.StepKey,
		res.ReservedUSD, res.ReservedTokens, res.Status, res.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert budget reservation: %w", err)
	}

	return tx.Commit()
}

func (r *PostgresRepo) CommitExecutionSpend(ctx context.Context, resID, employeeID, taskID string, costUSD float64, taskStatus domain.TaskState) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin commit tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Mark budget reservation as COMMITTED
	_, err = tx.ExecContext(ctx, "UPDATE budget_reservations SET status = 'COMMITTED' WHERE id = $1", resID)
	if err != nil {
		return fmt.Errorf("failed to commit budget reservation: %w", err)
	}

	// 2. Atomic delta increment on employee spent_usd (never overwrites concurrent spend)
	_, err = tx.ExecContext(ctx, "UPDATE employees SET spent_usd = spent_usd + $1, version = version + 1 WHERE id = $2", costUSD, employeeID)
	if err != nil {
		return fmt.Errorf("failed to increment employee spent: %w", err)
	}

	// 3. Atomic delta increment on task cost_usd and execs_count
	_, err = tx.ExecContext(ctx, `
		UPDATE tasks SET
			cost_usd = cost_usd + $1,
			execs_count = execs_count + 1,
			status = $2,
			updated_at = now(),
			version = version + 1
		WHERE id = $3
	`, costUSD, taskStatus, taskID)
	if err != nil {
		return fmt.Errorf("failed to increment task cost: %w", err)
	}

	return tx.Commit()
}

func (r *PostgresRepo) CreateBudgetReservation(ctx context.Context, res *domain.BudgetReservation) error {
	query := `
		INSERT INTO budget_reservations (
			id, organization_id, execution_id, step_key, reserved_usd, reserved_tokens, status, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		res.ID, res.OrganizationID, res.ExecutionID, res.StepKey,
		res.ReservedUSD, res.ReservedTokens, res.Status, res.ExpiresAt,
	)
	return err
}

func (r *PostgresRepo) UpdateBudgetReservationStatus(ctx context.Context, id, status string) error {
	query := `UPDATE budget_reservations SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *PostgresRepo) GetActiveBudgetReservations(ctx context.Context, employeeID string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(br.reserved_usd), 0)
		FROM budget_reservations br
		JOIN executions e ON br.execution_id = e.id
		WHERE e.executing_employee_id = $1
		  AND br.status = 'ACTIVE'
		  AND br.expires_at > now()
	`
	var total float64
	err := r.db.QueryRowContext(ctx, query, employeeID).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PostgresRepo) RecordUsageLedger(ctx context.Context, entry *domain.UsageLedgerEntry) error {
	query := `
		INSERT INTO usage_ledger (
			id, organization_id, execution_id, task_id, project_id, employee_id,
			model_version_id, charge_reference, entry_type, posted_at,
			input_tokens, output_tokens, cache_read_tokens, cost_usd
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.OrganizationID, entry.ExecutionID, entry.TaskID, entry.ProjectID, entry.EmployeeID,
		entry.ModelVersionID, entry.ChargeReference, entry.EntryType, entry.PostedAt,
		entry.InputTokens, entry.OutputTokens, entry.CacheReadTokens, entry.CostUSD,
	)
	return err
}

// ----------------- Organization, Department & Team Management -----------------

func (r *PostgresRepo) CreateDepartment(ctx context.Context, d *domain.Department) error {
	if d.ID == "" {
		d.ID = strings.ToLower(strings.ReplaceAll(d.Name, " ", "-"))
	}
	if d.OrganizationID == "" {
		d.OrganizationID = "org-default"
	}
	query := `INSERT INTO departments (id, organization_id, name, parent_department_id, created_at)
	          VALUES ($1, $2, $3, $4, now())`
	_, err := r.db.ExecContext(ctx, query, d.ID, d.OrganizationID, d.Name, d.ParentDepartmentID)
	return err
}

func (r *PostgresRepo) UpdateDepartment(ctx context.Context, d *domain.Department) error {
	query := `UPDATE departments SET name = $1, parent_department_id = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, d.Name, d.ParentDepartmentID, d.ID)
	return err
}

func (r *PostgresRepo) DeleteDepartment(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM departments WHERE id = $1`, id)
	return err
}

func (r *PostgresRepo) CreateTeam(ctx context.Context, t *domain.Team) error {
	if t.ID == "" {
		t.ID = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
	}
	if t.OrganizationID == "" {
		t.OrganizationID = "org-default"
	}
	query := `INSERT INTO teams (id, organization_id, department_id, name, lead_employee_id, created_at)
	          VALUES ($1, $2, $3, $4, $5, now())`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.OrganizationID, t.DepartmentID, t.Name, t.LeadEmployeeID)
	return err
}

func (r *PostgresRepo) UpdateTeam(ctx context.Context, t *domain.Team) error {
	query := `UPDATE teams SET name = $1, department_id = $2, lead_employee_id = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, t.Name, t.DepartmentID, t.LeadEmployeeID, t.ID)
	return err
}

func (r *PostgresRepo) DeleteTeam(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM teams WHERE id = $1`, id)
	return err
}

// ----------------- Employee & Hierarchy Management -----------------

func (r *PostgresRepo) DeleteEmployee(ctx context.Context, id string) error {
	// Soft delete/retire to preserve audit trail and references
	query := `UPDATE employees SET state = 'RETIRED' WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresRepo) UpdateReportingLine(ctx context.Context, employeeID string, managerID *string) error {
	query := `UPDATE employees SET manager_id = $1, version = version + 1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, managerID, employeeID)
	return err
}

// ----------------- Project Management -----------------

func (r *PostgresRepo) DeleteProject(ctx context.Context, id string) error {
	var taskCount int
	_ = r.db.QueryRowContext(ctx, `SELECT count(*) FROM tasks WHERE project_id = $1`, id).Scan(&taskCount)
	if taskCount > 0 {
		_, err := r.db.ExecContext(ctx, `UPDATE projects SET status = 'ARCHIVED' WHERE id = $1`, id)
		return err
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = $1`, id)
	return err
}

// ----------------- Capabilities (MCP & Models) Management -----------------

func (r *PostgresRepo) CreateMCPServer(ctx context.Context, s *domain.MCPServer) error {
	if s.ID == "" {
		s.ID = "mcp-" + strings.ToLower(s.Name)
	}
	if s.OrganizationID == "" {
		s.OrganizationID = "org-default"
	}
	if s.RiskTier == "" {
		s.RiskTier = "R1"
	}
	if s.ReviewStatus == "" {
		s.ReviewStatus = "APPROVED"
	}
	query := `INSERT INTO mcp_servers (id, organization_id, name, transport, description, risk_tier, enabled, review_status, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())`
	_, err := r.db.ExecContext(ctx, query, s.ID, s.OrganizationID, s.Name, s.Transport, s.Description, s.RiskTier, s.Enabled, s.ReviewStatus)
	return err
}

func (r *PostgresRepo) UpdateMCPServer(ctx context.Context, s *domain.MCPServer) error {
	query := `UPDATE mcp_servers SET name = $1, transport = $2, description = $3, risk_tier = $4, enabled = $5, review_status = $6 WHERE id = $7`
	_, err := r.db.ExecContext(ctx, query, s.Name, s.Transport, s.Description, s.RiskTier, s.Enabled, s.ReviewStatus, s.ID)
	return err
}

func (r *PostgresRepo) DeleteMCPServer(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM mcp_servers WHERE id = $1`, id)
	return err
}

func (r *PostgresRepo) CreateModel(ctx context.Context, m *domain.ModelInfo) error {
	if m.ID == "" {
		m.ID = fmt.Sprintf("model-%s-%s", strings.ToLower(m.ProviderCode), strings.ToLower(m.ModelID))
	}
	featJSON, _ := json.Marshal(m.Features)
	if m.Revision == "" {
		m.Revision = "default"
	}
	if m.Status == "" {
		m.Status = "ACTIVE"
	}
	query := `INSERT INTO model_versions (id, provider_code, model_id, revision, features, status)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, m.ID, m.ProviderCode, m.ModelID, m.Revision, featJSON, m.Status)
	return err
}

func (r *PostgresRepo) UpdateModel(ctx context.Context, m *domain.ModelInfo) error {
	featJSON, _ := json.Marshal(m.Features)
	query := `UPDATE model_versions SET provider_code = $1, model_id = $2, revision = $3, features = $4, status = $5 WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, m.ProviderCode, m.ModelID, m.Revision, featJSON, m.Status, m.ID)
	return err
}

func (r *PostgresRepo) DeleteModel(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM model_versions WHERE id = $1`, id)
	return err
}

