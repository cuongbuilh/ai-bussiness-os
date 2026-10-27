package repository

import (
	"context"
	"github.com/cuongbuilh/ai-business-os/backend/internal/domain"
)

type Repository interface {
	// Organization & Departments & Teams
	GetDepartments(ctx context.Context) ([]domain.Department, error)
	GetTeams(ctx context.Context) ([]domain.Team, error)
	GetPositions(ctx context.Context) ([]domain.Position, error)
	GetOrgTree(ctx context.Context) (*domain.OrgNode, error)

	// Employees
	GetEmployees(ctx context.Context) ([]domain.Employee, error)
	GetEmployeeByID(ctx context.Context, id string) (*domain.Employee, error)
	CreateEmployee(ctx context.Context, emp *domain.Employee) error
	UpdateEmployee(ctx context.Context, emp *domain.Employee) error

	// Projects
	GetProjects(ctx context.Context) ([]domain.Project, error)
	GetProjectByID(ctx context.Context, id string) (*domain.Project, error)
	CreateProject(ctx context.Context, p *domain.Project) error
	UpdateProject(ctx context.Context, p *domain.Project) error

	// Tasks
	GetTasks(ctx context.Context, scope, projectFilter, ownerFilter, statusFilter, search string) ([]domain.Task, error)
	GetTaskByID(ctx context.Context, id string) (*domain.Task, error)
	CreateTask(ctx context.Context, task *domain.Task) error
	UpdateTask(ctx context.Context, task *domain.Task) error

	// Executions
	GetExecutionsByTaskID(ctx context.Context, taskID string) ([]domain.Execution, error)
	CreateExecution(ctx context.Context, exec *domain.Execution) error
	UpdateExecution(ctx context.Context, exec *domain.Execution) error

	// Approvals
	GetApprovals(ctx context.Context) ([]domain.ApprovalRequest, error)
	GetApprovalByID(ctx context.Context, id string) (*domain.ApprovalRequest, error)
	UpdateApproval(ctx context.Context, approval *domain.ApprovalRequest) error

	// Capabilities & Models
	GetMCPServers(ctx context.Context) ([]domain.MCPServer, error)
	GetModels(ctx context.Context) ([]domain.ModelInfo, error)

	// Metrics
	GetBoardSummary(ctx context.Context, scope string) (*domain.BoardSummary, error)

	// Budget & Ledger
	AtomicReserveBudget(ctx context.Context, res *domain.BudgetReservation, employeeID string, estimatedCost float64) error
	CommitExecutionSpend(ctx context.Context, resID, employeeID, taskID string, costUSD float64, taskStatus domain.TaskState) error
	CreateBudgetReservation(ctx context.Context, res *domain.BudgetReservation) error
	UpdateBudgetReservationStatus(ctx context.Context, id, status string) error
	GetActiveBudgetReservations(ctx context.Context, employeeID string) (float64, error)
	RecordUsageLedger(ctx context.Context, entry *domain.UsageLedgerEntry) error
}
