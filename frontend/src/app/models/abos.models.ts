export type EmployeeKind = 'HUMAN' | 'AI';
export type EmployeeState = 'DRAFT' | 'ACTIVE' | 'SUSPENDED' | 'RETIRED';
export type TaskState = 'DRAFT' | 'backlog' | 'ready' | 'in_progress' | 'blocked' | 'in_review' | 'completed' | 'accepted' | 'rejected' | 'cancelled';
export type ExecutionState = 'CREATED' | 'QUEUED' | 'RUNNING' | 'WAITING_APPROVAL' | 'WAITING_TOOL' | 'SUCCEEDED' | 'FAILED' | 'TIMED_OUT' | 'CANCELLED';

export interface Department {
  id: string;
  organizationId: string;
  name: string;
  parentDepartmentId?: string;
  createdAt: string;
}

export interface Team {
  id: string;
  organizationId: string;
  departmentId: string;
  name: string;
  leadEmployeeId?: string;
  createdAt: string;
}

export interface Position {
  id: string;
  organizationId: string;
  departmentId?: string;
  title: string;
  roleCode: string;
  description: string;
  capMcps?: string[];
}

export interface Employee {
  id: string;
  organizationId: string;
  departmentId?: string;
  positionId?: string;
  displayName: string;
  name: string;
  kind: EmployeeKind;
  state: EmployeeState;
  title: string;
  role: string;
  departmentName?: string;
  managerId?: string;
  managerName?: string;
  model: string;
  provider: string;
  skills: string[];
  mcp: string[];
  teams: string[];
  kpi: number;
  budget: number;
  spent: number;
  version: number;
  createdAt: string;
}

export interface Milestone {
  name: string;
  done: boolean;
}

export interface Project {
  id: string;
  organizationId: string;
  code: string;
  name: string;
  description: string;
  owner: string;
  ownerName?: string;
  status: string;
  budget: number;
  spent: number;
  due: string;
  members: string[];
  teams: string[];
  milestones: Milestone[];
  version: number;
}

export interface Task {
  id: string;
  organizationId: string;
  project?: string;
  projectCode?: string;
  projectName?: string;
  parentTaskId?: string;
  title: string;
  description: string;
  owner: string;
  ownerName?: string;
  collab?: string[];
  status: TaskState;
  priority: string;
  criteria: string;
  due: string;
  acceptedAt?: string;
  cost: number;
  execs: number;
  originType?: string;
  originRef?: string;
  createdAt: string;
  updatedAt: string;
  version: number;
}

export interface ExecutionStep {
  stepNo: number;
  stepName: string;
  toolName?: string;
  input?: string;
  output?: string;
  tokensUsed: number;
  costUsd: number;
  status: string;
  timestamp: string;
}

export interface Execution {
  id: string;
  organizationId: string;
  taskId: string;
  executingEmployeeId: string;
  executingEmployeeName?: string;
  status: ExecutionState;
  attemptNo: number;
  idempotencyKey: string;
  logs: string[];
  steps: ExecutionStep[];
  totalTokens: number;
  totalCostUsd: number;
  createdAt: string;
  startedAt?: string;
  endedAt?: string;
}

export interface ApprovalRequest {
  id: string;
  organizationId: string;
  taskId?: string;
  taskTitle?: string;
  requestedBy: string;
  requesterName?: string;
  type: string;
  payloadJson?: any;
  risk: string;
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'EXPIRED';
  title: string;
  expiresAt?: string;
  resolvedBy?: string;
  resolvedAt?: string;
  createdAt: string;
}

export interface MCPServer {
  id: string;
  name: string;
  transport: string;
  description: string;
  risk: string;
  enabled: boolean;
  reviewStatus: string;
}

export interface ModelInfo {
  id: string;
  provider: string;
  modelId: string;
  revision: string;
  features: string[];
  status: string;
}

export interface OrgNode {
  id: string;
  name: string;
  title: string;
  role: string;
  kind: string;
  model: string;
  provider: string;
  department: string;
  children?: OrgNode[];
}

export interface BoardSummary {
  totalTasks: number;
  activeTasks: number;
  acceptedTasks: number;
  blockedTasks: number;
  backlogTasks: number;
  totalSpentUsd: number;
  projectsCount: number;
}

export interface AuthUser {
  id: string;
  displayName: string;
  role: string;
  kind: string;
  title: string;
}

export interface LoginResponse {
  token: string;
  user: AuthUser;
}
