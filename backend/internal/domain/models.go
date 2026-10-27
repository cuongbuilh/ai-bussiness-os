package domain

import (
	"encoding/json"
	"time"
)

type EmployeeKind string

const (
	EmployeeKindHuman EmployeeKind = "HUMAN"
	EmployeeKindAI    EmployeeKind = "AI"
)

type EmployeeState string

const (
	EmployeeStateDraft     EmployeeState = "DRAFT"
	EmployeeStateActive    EmployeeState = "ACTIVE"
	EmployeeStateSuspended EmployeeState = "SUSPENDED"
	EmployeeStateRetired   EmployeeState = "RETIRED"
)

type TaskState string

const (
	TaskStateDraft      TaskState = "DRAFT"
	TaskStateBacklog    TaskState = "backlog"
	TaskStateReady      TaskState = "ready"
	TaskStateInProgress TaskState = "in_progress"
	TaskStateBlocked    TaskState = "blocked"
	TaskStateInReview   TaskState = "in_review"
	TaskStateCompleted  TaskState = "completed"
	TaskStateAccepted   TaskState = "accepted"
	TaskStateRejected   TaskState = "rejected"
	TaskStateCancelled  TaskState = "cancelled"
)

type ExecutionState string

const (
	ExecutionStateCreated         ExecutionState = "CREATED"
	ExecutionStateQueued          ExecutionState = "QUEUED"
	ExecutionStateRunning         ExecutionState = "RUNNING"
	ExecutionStateWaitingApproval ExecutionState = "WAITING_APPROVAL"
	ExecutionStateWaitingTool     ExecutionState = "WAITING_TOOL"
	ExecutionStateSucceeded       ExecutionState = "SUCCEEDED"
	ExecutionStateFailed          ExecutionState = "FAILED"
	ExecutionStateTimedOut        ExecutionState = "TIMED_OUT"
	ExecutionStateCancelled       ExecutionState = "CANCELLED"
)

type ApprovalState string

const (
	ApprovalStatePending  ApprovalState = "PENDING"
	ApprovalStateApproved ApprovalState = "APPROVED"
	ApprovalStateRejected ApprovalState = "REJECTED"
	ApprovalStateExpired  ApprovalState = "EXPIRED"
)

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Department struct {
	ID                 string    `json:"id"`
	OrganizationID     string    `json:"organizationId"`
	Name               string    `json:"name"`
	ParentDepartmentID *string   `json:"parentDepartmentId,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
}

type Team struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	DepartmentID   string    `json:"departmentId"`
	Name           string    `json:"name"`
	LeadEmployeeID *string   `json:"leadEmployeeId,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Position struct {
	ID             string   `json:"id"`
	OrganizationID string   `json:"organizationId"`
	DepartmentID   *string  `json:"departmentId,omitempty"`
	Title          string   `json:"title"`
	RoleCode       string   `json:"roleCode"`
	Description    string   `json:"description"`
	CapMCPs        []string `json:"capMcps,omitempty"`
}

type Employee struct {
	ID             string        `json:"id"`
	OrganizationID string        `json:"organizationId"`
	DepartmentID   *string       `json:"departmentId,omitempty"`
	PositionID     *string       `json:"positionId,omitempty"`
	DisplayName    string        `json:"displayName"`
	Name           string        `json:"name"`
	Kind           EmployeeKind  `json:"kind"`
	State          EmployeeState `json:"state"`
	Title          string        `json:"title"`
	Role           string        `json:"role"`
	DepartmentName string        `json:"departmentName,omitempty"`
	ManagerID      *string       `json:"managerId,omitempty"`
	ManagerName    string        `json:"managerName,omitempty"`
	Model          string        `json:"model"`
	Provider       string        `json:"provider"`
	Skills         []string      `json:"skills"`
	MCP            []string      `json:"mcp"`
	Teams          []string      `json:"teams"`
	KPI            int           `json:"kpi"`
	BudgetUSD      float64       `json:"budget"`
	SpentUSD       float64       `json:"spent"`
	Version        int64         `json:"version"`
	CreatedAt      time.Time     `json:"createdAt"`
}

type ReportingLine struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organizationId"`
	SubordinateID  string     `json:"subordinateId"`
	ManagerID      string     `json:"managerId"`
	ValidFrom      time.Time  `json:"validFrom"`
	ValidTo        *time.Time `json:"validTo,omitempty"`
}

type Milestone struct {
	Name string `json:"name"`
	Done bool   `json:"done"`
}

type Project struct {
	ID                 string      `json:"id"`
	OrganizationID     string      `json:"organizationId"`
	Code               string      `json:"code"`
	Name               string      `json:"name"`
	Description        string      `json:"description"`
	AccountableOwnerID string      `json:"owner"`
	OwnerName          string      `json:"ownerName,omitempty"`
	Status             string      `json:"status"`
	BudgetUSD          float64     `json:"budget"`
	SpentUSD           float64     `json:"spent"`
	DueAt              string      `json:"due"`
	Members            []string    `json:"members"`
	Teams              []string    `json:"teams"`
	Milestones         []Milestone `json:"milestones"`
	CreatedAt          time.Time   `json:"createdAt"`
	Version            int64       `json:"version"`
}

type Task struct {
	ID                   string     `json:"id"`
	OrganizationID       string     `json:"organizationId"`
	ProjectID            *string    `json:"project"`
	ProjectCode          string     `json:"projectCode,omitempty"`
	ProjectName          string     `json:"projectName,omitempty"`
	ParentTaskID         *string    `json:"parentTaskId,omitempty"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	AccountableOwnerID   string     `json:"owner"`
	OwnerName            string     `json:"ownerName,omitempty"`
	Collaborators        []string   `json:"collab"`
	Status               TaskState  `json:"status"`
	Priority             string     `json:"priority"`
	AcceptanceCriteria   string     `json:"criteria"`
	DueAt                string     `json:"due"`
	AcceptedAt           *time.Time `json:"acceptedAt,omitempty"`
	CostUSD              float64    `json:"cost"`
	ExecutionsCount      int        `json:"execs"`
	OriginType           string     `json:"originType"`
	OriginRef            string     `json:"originRef,omitempty"`
	InitiatedByEmpID     *string    `json:"initiatedByEmployeeId,omitempty"`
	TriggerOccurrenceID  *string    `json:"triggerOccurrenceId,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	Version              int64      `json:"version"`
}

type ExecutionStep struct {
	StepNo     int       `json:"stepNo"`
	StepName   string    `json:"stepName"`
	ToolName   string    `json:"toolName,omitempty"`
	Input      string    `json:"input,omitempty"`
	Output     string    `json:"output,omitempty"`
	TokensUsed int64     `json:"tokensUsed"`
	CostUSD    float64   `json:"costUsd"`
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
}

type Execution struct {
	ID                  string          `json:"id"`
	OrganizationID      string          `json:"organizationId"`
	TaskID              string          `json:"taskId"`
	ExecutingEmployeeID string          `json:"executingEmployeeId"`
	ExecutingEmpName    string          `json:"executingEmployeeName,omitempty"`
	ModelAssignmentID   *string         `json:"modelAssignmentId,omitempty"`
	Status              ExecutionState  `json:"status"`
	AttemptNo           int             `json:"attemptNo"`
	IdempotencyKey      string          `json:"idempotencyKey"`
	Logs                []string        `json:"logs"`
	Steps               []ExecutionStep `json:"steps"`
	TotalTokens         int64           `json:"totalTokens"`
	TotalCostUSD        float64         `json:"totalCostUsd"`
	CreatedAt           time.Time       `json:"createdAt"`
	StartedAt           *time.Time      `json:"startedAt,omitempty"`
	EndedAt             *time.Time      `json:"endedAt,omitempty"`
}

type ApprovalRequest struct {
	ID                    string          `json:"id"`
	OrganizationID        string          `json:"organizationId"`
	TaskID                *string         `json:"taskId,omitempty"`
	TaskTitle             string          `json:"taskTitle,omitempty"`
	RequestedByEmployeeID string          `json:"requestedBy"`
	RequesterName         string          `json:"requesterName,omitempty"`
	ActionKind            string          `json:"type"`
	PayloadJSON           json.RawMessage `json:"payloadJson"`
	RiskTier              string          `json:"risk"`
	Status                ApprovalState   `json:"status"`
	Title                 string          `json:"title"`
	ExpiresAt             *time.Time      `json:"expiresAt,omitempty"`
	ResolvedByEmployeeID  *string         `json:"resolvedBy,omitempty"`
	ResolvedAt            *time.Time      `json:"resolvedAt,omitempty"`
	CreatedAt             time.Time       `json:"createdAt"`
}

type MCPServer struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	Name           string    `json:"name"`
	Transport      string    `json:"transport"`
	Description    string    `json:"description"`
	RiskTier       string    `json:"risk"`
	Enabled        bool      `json:"enabled"`
	ReviewStatus   string    `json:"reviewStatus"`
	CreatedAt      time.Time `json:"createdAt"`
}

type ModelInfo struct {
	ID           string   `json:"id"`
	ProviderCode string   `json:"provider"`
	ModelID      string   `json:"modelId"`
	Revision     string   `json:"revision"`
	Features     []string `json:"features"`
	Status       string   `json:"status"`
}

type OrgNode struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Title      string     `json:"title"`
	Role       string     `json:"role"`
	Kind       string     `json:"kind"`
	Model      string     `json:"model"`
	Provider   string     `json:"provider"`
	Department string     `json:"department"`
	DirectSub  []*OrgNode `json:"children,omitempty"`
}

type BoardSummary struct {
	TotalTasks    int     `json:"totalTasks"`
	ActiveTasks   int     `json:"activeTasks"`
	AcceptedTasks int     `json:"acceptedTasks"`
	BlockedTasks  int     `json:"blockedTasks"`
	BacklogTasks  int     `json:"backlogTasks"`
	TotalSpentUSD float64 `json:"totalSpentUsd"`
	ProjectsCount int     `json:"projectsCount"`
}

type EventMessage struct {
	Type      string      `json:"type"`
	Action    string      `json:"action"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}
