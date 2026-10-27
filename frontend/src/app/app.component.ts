import { Component, OnInit, OnDestroy, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Subscription } from 'rxjs';
import { ApiService } from './services/api.service';
import {
  Department, Team, Employee, Project, Task,
  Execution, ApprovalRequest, MCPServer, ModelInfo, OrgNode, BoardSummary,
  AuthUser
} from './models/abos.models';

type PageTab = 'overview' | 'organization' | 'board' | 'projects' | 'employees' | 'capabilities' | 'approvals';
type BoardMode = 'kanban' | 'table';
type EmployeeTab = 'running' | 'overview' | 'tasks' | 'tools' | 'kpis';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent implements OnInit, OnDestroy {
  private api = inject(ApiService);
  private sseSub?: Subscription;

  // Active navigation & filters
  activePage: PageTab = 'overview';
  boardMode: BoardMode = 'kanban';
  scope = 'company:all';
  filterProject = '';
  filterOwner = '';
  filterStatus = '';
  filterText = '';

  // Data collections
  departments: Department[] = [];
  teams: Team[] = [];
  employees: Employee[] = [];
  projects: Project[] = [];
  tasks: Task[] = [];
  approvals: ApprovalRequest[] = [];
  mcpServers: MCPServer[] = [];
  models: ModelInfo[] = [];
  orgTree: OrgNode | null = null;
  boardSummary: BoardSummary = {
    totalTasks: 0,
    activeTasks: 0,
    acceptedTasks: 0,
    blockedTasks: 0,
    backlogTasks: 0,
    totalSpentUsd: 0,
    projectsCount: 0
  };

  // Status mapping
  kanbanColumns = [
    { code: 'backlog', label: 'Backlog' },
    { code: 'ready', label: 'Ready' },
    { code: 'in_progress', label: 'In progress' },
    { code: 'blocked', label: 'Blocked' },
    { code: 'in_review', label: 'In review' },
    { code: 'completed', label: 'Completed' },
    { code: 'accepted', label: 'Accepted' }
  ];

  // Selection states for Drawers/Modals
  selectedEmployee: Employee | null = null;
  selectedEmpTab: EmployeeTab = 'running';
  selectedEmpCaps: any = null;
  selectedEmpExecutions: Execution[] = [];

  selectedTask: Task | null = null;
  taskExecutions: Execution[] = [];
  isExecutingTask = false;

  selectedProject: Project | null = null;

  // Authentication state & user session
  authToken = this.api.getAuthToken();
  isAuthenticated = this.api.isAuthenticated();
  currentUser: AuthUser | null = this.api.getCurrentUser();
  showAuthModal = false;
  inputToken = '';

  // Login form state
  loginUsername = 'founder';
  loginPassword = 'founder123';
  loginTokenInput = '';
  loginMode: 'password' | 'token' = 'password';
  loginLoading = false;
  loginError = '';

  // Theme state: dark | light
  theme: 'dark' | 'light' = 'dark';

  doLogin(): void {
    this.loginError = '';
    this.loginLoading = true;

    const payload = this.loginMode === 'password'
      ? { username: this.loginUsername, password: this.loginPassword }
      : { token: this.loginTokenInput.trim() };

    this.api.login(payload).subscribe({
      next: (res) => {
        this.api.setAuthToken(res.token);
        this.api.setCurrentUser(res.user);
        this.currentUser = res.user;
        this.authToken = res.token;
        this.isAuthenticated = true;
        this.loginLoading = false;
        this.loadAllData();
        this.setupSSEListener();
        this.notify(`Đăng nhập thành công! Chào mừng ${res.user.displayName}`);
      },
      error: (err) => {
        this.loginLoading = false;
        this.loginError = err?.error?.error || 'Đăng nhập không thành công. Vui lòng kiểm tra lại thông tin.';
      }
    });
  }

  fillFounderDemo(): void {
    this.loginMode = 'password';
    this.loginUsername = 'founder';
    this.loginPassword = 'founder123';
    this.doLogin();
  }

  logout(): void {
    this.api.clearAuthToken();
    this.authToken = '';
    this.currentUser = null;
    this.isAuthenticated = false;
    this.sseSub?.unsubscribe();
    this.departments = [];
    this.teams = [];
    this.employees = [];
    this.projects = [];
    this.tasks = [];
    this.approvals = [];
    this.notify('Đã đăng xuất khỏi hệ thống điều hành.');
  }

  openAuthModal(): void {
    this.inputToken = this.authToken;
    this.showAuthModal = true;
  }

  saveToken(): void {
    const trimmed = this.inputToken.trim();
    if (trimmed) {
      this.api.setAuthToken(trimmed);
      this.authToken = trimmed;
      this.showAuthModal = false;
      this.loadAllData();
      this.notify('Đã cập nhật Bearer token thành công.');
    }
  }

  clearToken(): void {
    this.logout();
    this.showAuthModal = false;
  }

  // New Task Form Modal
  showCreateTaskModal = false;
  newTask: Partial<Task> = {
    title: '',
    description: '',
    owner: '',
    project: '',
    priority: 'HIGH',
    criteria: '',
    due: ''
  };

  // Toast Notification
  toastMessage = '';
  showToast = false;
  private toastTimer?: any;

  ngOnInit(): void {
    const savedTheme = localStorage.getItem('abos_theme') as 'dark' | 'light' | null;
    if (savedTheme === 'light' || savedTheme === 'dark') {
      this.theme = savedTheme;
    } else if (typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
      this.theme = 'light';
    }
    this.applyTheme(this.theme);

    if (this.isAuthenticated) {
      this.loadAllData();
      this.setupSSEListener();
      this.api.getMe().subscribe({
        next: (u) => {
          this.currentUser = u;
          this.api.setCurrentUser(u);
        },
        error: () => {
          // If token expired or invalid, reset session
          this.logout();
        }
      });
    }
  }

  toggleTheme(): void {
    this.theme = this.theme === 'dark' ? 'light' : 'dark';
    localStorage.setItem('abos_theme', this.theme);
    this.applyTheme(this.theme);
    this.notify(this.theme === 'dark' ? 'Đã bật chế độ giao diện Tối' : 'Đã bật chế độ giao diện Sáng');
  }

  private applyTheme(t: 'dark' | 'light'): void {
    if (typeof document !== 'undefined') {
      document.documentElement.setAttribute('data-theme', t);
      document.body.setAttribute('data-theme', t);
    }
  }

  ngOnDestroy(): void {
    this.sseSub?.unsubscribe();
  }

  loadAllData(): void {
    this.api.getDepartments().subscribe(d => this.departments = d);
    this.api.getTeams().subscribe(t => this.teams = t);
    this.api.getEmployees().subscribe(e => this.employees = e);
    this.api.getProjects().subscribe(p => this.projects = p);
    this.api.getApprovals().subscribe(a => this.approvals = a);
    this.api.getMCPServers().subscribe(m => this.mcpServers = m);
    this.api.getModels().subscribe(m => this.models = m);
    this.api.getOrgChart().subscribe(tree => this.orgTree = tree);
    this.loadTasks();
    this.loadSummary();
  }

  setupSSEListener(): void {
    this.sseSub = this.api.getEvents().subscribe(event => {
      if (event.type === 'system' && event.action === 'refetch_snapshot') {
        this.loadAllData();
        return;
      }
      this.notify(`Live update [${event.type}]: ${event.action}`);
      if (event.type === 'task' || event.type === 'execution') {
        this.loadTasks();
        this.loadSummary();
        if (this.selectedTask && event.payload?.taskId === this.selectedTask.id) {
          this.openTaskDetail(this.selectedTask.id);
        }
      } else if (event.type === 'approval') {
        this.api.getApprovals().subscribe(a => this.approvals = a);
      } else if (event.type === 'employee' || event.type === 'hierarchy') {
        this.api.getEmployees().subscribe(e => this.employees = e);
        this.api.getOrgChart().subscribe(tree => this.orgTree = tree);
      } else if (event.type === 'model') {
        this.api.getModels().subscribe(m => this.models = m);
      } else if (event.type === 'mcp_server') {
        this.api.getMCPServers().subscribe(s => this.mcpServers = s);
      } else if (event.type === 'project') {
        this.api.getProjects().subscribe(p => this.projects = p);
      } else if (event.type === 'department' || event.type === 'team') {
        this.api.getDepartments().subscribe(d => this.departments = d);
        this.api.getTeams().subscribe(t => this.teams = t);
      }
    });
  }

  loadTasks(): void {
    this.api.getTasks(this.scope, this.filterProject, this.filterOwner, this.filterStatus, this.filterText)
      .subscribe(tasks => {
        this.tasks = tasks;
      });
  }

  loadSummary(): void {
    this.api.getBoardSummary(this.scope).subscribe(sum => {
      this.boardSummary = sum;
    });
  }

  notify(msg: string): void {
    this.toastMessage = msg;
    this.showToast = true;
    clearTimeout(this.toastTimer);
    this.toastTimer = setTimeout(() => {
      this.showToast = false;
    }, 3500);
  }

  setPage(page: PageTab): void {
    this.activePage = page;
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }

  onScopeChange(): void {
    this.loadTasks();
    this.loadSummary();
  }

  resetFilters(): void {
    this.scope = 'company:all';
    this.filterProject = '';
    this.filterOwner = '';
    this.filterStatus = '';
    this.filterText = '';
    this.loadTasks();
    this.loadSummary();
  }

  // ---------------- Tasks & Kanban ----------------
  getTasksByColumn(status: string): Task[] {
    return this.tasks.filter(t => t.status === status);
  }

  draggedTask: Task | null = null;
  onTaskDragStart(event: DragEvent, task: Task): void {
    this.draggedTask = task;
    if (event.dataTransfer) {
      event.dataTransfer.setData('text/plain', task.id);
      event.dataTransfer.effectAllowed = 'move';
    }
  }

  onTaskDragOver(event: DragEvent): void {
    event.preventDefault();
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = 'move';
    }
  }

  onTaskDrop(event: DragEvent, newStatus: string): void {
    event.preventDefault();
    if (this.draggedTask && this.draggedTask.status !== newStatus) {
      this.updateTaskStatus(this.draggedTask, newStatus);
    }
    this.draggedTask = null;
  }

  updateTaskStatus(task: Task, newStatus: string): void {
    this.api.updateTaskStatus(task.id, newStatus).subscribe({
      next: () => {
        task.status = newStatus as any;
        this.notify(`Task ${task.id} chuyển sang ${newStatus}`);
        this.loadSummary();
      },
      error: (err) => this.notify(`Lỗi: ${err?.error?.error || err.message}`)
    });
  }

  openTaskDetail(taskId: string): void {
    this.api.getTaskDetail(taskId).subscribe(t => {
      this.selectedTask = t;
      this.api.getTaskExecutions(taskId).subscribe(execs => {
        this.taskExecutions = execs;
      });
    });
  }

  runTaskExecution(taskId: string): void {
    this.isExecutingTask = true;
    this.api.executeTask(taskId).subscribe({
      next: (exec) => {
        this.isExecutingTask = false;
        this.notify(`Agent ${exec.executingEmployeeName} đã chạy task #${exec.attemptNo} thành công!`);
        this.openTaskDetail(taskId);
        this.loadTasks();
        this.loadSummary();
      },
      error: (err) => {
        this.isExecutingTask = false;
        this.notify(`Lỗi thực thi: ${err?.error?.error || err.message}`);
      }
    });
  }

  acceptTask(taskId: string): void {
    this.api.acceptTask(taskId, 'Nghiệm thu đạt yêu cầu bởi Quản lý/Founder').subscribe({
      next: () => {
        this.notify(`Task ${taskId} đã được Nghiệm thu (ACCEPTED)!`);
        this.openTaskDetail(taskId);
        this.loadTasks();
        this.loadSummary();
      },
      error: (err) => this.notify(`Lỗi: ${err?.error?.error || err.message}`)
    });
  }

  // ---------------- Employee Drawer ----------------
  openEmployeeDetail(id: string, tab?: EmployeeTab): void {
    const defaultTab: EmployeeTab = tab || 'running';
    this.api.getEmployeeDetail(id).subscribe(emp => {
      this.selectedEmployee = emp;
      this.selectedEmpTab = defaultTab;
      this.api.getEmployeeCapabilities(id).subscribe(caps => {
        this.selectedEmpCaps = caps;
      });
      this.api.getEmployeeExecutions(id).subscribe(execs => {
        this.selectedEmpExecutions = execs;
      });
    });
  }

  setEmpTab(tab: EmployeeTab): void {
    this.selectedEmpTab = tab;
  }

  getEmployeeTasks(empId: string): Task[] {
    return this.tasks.filter(t => t.owner === empId || t.collab?.includes(empId));
  }

  getActiveTasksForEmployee(empId: string): Task[] {
    return this.tasks.filter(t => (t.owner === empId || t.collab?.includes(empId)) && ['in_progress', 'in_review', 'ready', 'blocked'].includes(t.status));
  }

  getRunningTasksCount(empId: string): number {
    return this.getActiveTasksForEmployee(empId).length;
  }

  runAgentTask(taskId: string): void {
    this.isExecutingTask = true;
    this.api.executeTask(taskId).subscribe({
      next: (exec) => {
        this.isExecutingTask = false;
        this.notify(`Agent ${exec.executingEmployeeName} đã chạy thành công task #${exec.attemptNo}!`);
        this.loadTasks();
        this.loadSummary();
        if (this.selectedEmployee) {
          this.api.getEmployeeExecutions(this.selectedEmployee.id).subscribe(execs => {
            this.selectedEmpExecutions = execs;
          });
        }
      },
      error: (err) => {
        this.isExecutingTask = false;
        this.notify(`Lỗi thực thi: ${err?.error?.error || err.message}`);
      }
    });
  }

  // ---------------- Project Modal ----------------
  openProjectDetail(id: string): void {
    this.api.getProjectDetail(id).subscribe(p => {
      this.selectedProject = p;
    });
  }

  // ---------------- Create Task Modal ----------------
  openCreateTaskModal(prefillOwner?: string, prefillProject?: string): void {
    this.newTask = {
      title: '',
      description: '',
      owner: prefillOwner || (this.employees.find(e => e.kind === 'AI')?.id || 'm2'),
      project: prefillProject || (this.projects[0]?.id || ''),
      priority: 'HIGH',
      criteria: 'Entity rõ và có tests đầy đủ',
      due: new Date(Date.now() + 7 * 86400000).toISOString().split('T')[0]
    };
    this.showCreateTaskModal = true;
  }

  submitCreateTask(): void {
    if (!this.newTask.title || !this.newTask.owner) {
      alert('Vui lòng nhập tiêu đề task và chọn người chịu trách nhiệm');
      return;
    }
    this.api.createTask(this.newTask).subscribe({
      next: (t) => {
        this.notify(`Đã tạo thành công task ${t.id}: ${t.title}`);
        this.showCreateTaskModal = false;
        this.loadTasks();
        this.loadSummary();
      },
      error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
    });
  }

  // ---------------- Approvals ----------------
  decideApproval(id: string, decision: 'APPROVE' | 'REJECT'): void {
    this.api.decideApproval(id, decision).subscribe({
      next: () => {
        this.notify(`Yêu cầu ${id} đã được ${decision === 'APPROVE' ? 'PHÊ DUYỆT' : 'TỪ CHỐI'}!`);
        this.api.getApprovals().subscribe(a => this.approvals = a);
        this.loadTasks();
        this.loadSummary();
      },
      error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
    });
  }

  // Helpers
  getStaff(id?: string): Employee | undefined {
    return this.employees.find(e => e.id === id);
  }

  getProject(id?: string): Project | undefined {
    return this.projects.find(p => p.id === id);
  }

  getDepartmentName(deptId?: string): string {
    return this.departments.find(d => d.id === deptId)?.name || 'Operations';
  }

  getPendingApprovalsCount(): number {
    return this.approvals.filter(a => a.status === 'PENDING').length;
  }

  getActiveAIEmployeesCount(): number {
    return this.employees.filter(e => e.kind === 'AI').length;
  }

  // ---------------- CRUD Capabilities: Models ----------------
  showModelModal = false;
  isEditingModel = false;
  modelForm: Partial<ModelInfo> & { featuresText?: string } = {
    provider: 'Anthropic',
    modelId: '',
    revision: '2024-10',
    status: 'ACTIVE',
    featuresText: 'Tool Calling, JSON Schema, Streaming'
  };

  openCreateModelModal(): void {
    this.isEditingModel = false;
    this.modelForm = {
      provider: 'Anthropic',
      modelId: '',
      revision: '2024-10',
      status: 'ACTIVE',
      featuresText: 'Tool Calling, JSON Schema, Streaming'
    };
    this.showModelModal = true;
  }

  openEditModelModal(m: ModelInfo, event?: Event): void {
    if (event) event.stopPropagation();
    this.isEditingModel = true;
    this.modelForm = {
      ...m,
      featuresText: (m.features || []).join(', ')
    };
    this.showModelModal = true;
  }

  saveModel(): void {
    if (!this.modelForm.provider || !this.modelForm.modelId) {
      alert('Vui lòng nhập Nhà cung cấp (Provider) và Mã Model ID');
      return;
    }
    const features = (this.modelForm.featuresText || '')
      .split(',')
      .map(s => s.trim())
      .filter(s => s.length > 0);

    const payload: Partial<ModelInfo> = {
      ...this.modelForm,
      features
    };

    if (this.isEditingModel && this.modelForm.id) {
      this.api.updateModel(this.modelForm.id, payload).subscribe({
        next: () => {
          this.notify(`Đã cập nhật mô hình ${this.modelForm.modelId}`);
          this.showModelModal = false;
          this.api.getModels().subscribe(m => this.models = m);
        },
        error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
      });
    } else {
      this.api.createModel(payload).subscribe({
        next: () => {
          this.notify(`Đã thêm mô hình mới ${this.modelForm.modelId}`);
          this.showModelModal = false;
          this.api.getModels().subscribe(m => this.models = m);
        },
        error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
      });
    }
  }

  deleteModel(id: string, event?: Event): void {
    if (event) event.stopPropagation();
    if (!confirm('Bạn có chắc chắn muốn gỡ bỏ mô hình này không?')) return;
    this.api.deleteModel(id).subscribe({
      next: () => {
        this.notify('Đã xóa mô hình thành công');
        this.api.getModels().subscribe(m => this.models = m);
      },
      error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
    });
  }

  // ---------------- CRUD Capabilities: MCP Servers ----------------
  showMCPModal = false;
  isEditingMCP = false;
  mcpForm: Partial<MCPServer> = {
    name: '',
    transport: 'stdio',
    description: '',
    risk: 'R1',
    enabled: true,
    reviewStatus: 'APPROVED'
  };

  openCreateMCPModal(): void {
    this.isEditingMCP = false;
    this.mcpForm = {
      name: '',
      transport: 'stdio',
      description: '',
      risk: 'R1',
      enabled: true,
      reviewStatus: 'APPROVED'
    };
    this.showMCPModal = true;
  }

  openEditMCPModal(s: MCPServer, event?: Event): void {
    if (event) event.stopPropagation();
    this.isEditingMCP = true;
    this.mcpForm = { ...s };
    this.showMCPModal = true;
  }

  saveMCP(): void {
    if (!this.mcpForm.name) {
      alert('Vui lòng nhập tên MCP Server');
      return;
    }
    if (this.isEditingMCP && this.mcpForm.id) {
      this.api.updateMCPServer(this.mcpForm.id, this.mcpForm).subscribe({
        next: () => {
          this.notify(`Đã cập nhật MCP Server ${this.mcpForm.name}`);
          this.showMCPModal = false;
          this.api.getMCPServers().subscribe(s => this.mcpServers = s);
        },
        error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
      });
    } else {
      this.api.createMCPServer(this.mcpForm).subscribe({
        next: () => {
          this.notify(`Đã đăng ký MCP Server ${this.mcpForm.name}`);
          this.showMCPModal = false;
          this.api.getMCPServers().subscribe(s => this.mcpServers = s);
        },
        error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
      });
    }
  }

  deleteMCP(id: string, event?: Event): void {
    if (event) event.stopPropagation();
    if (!confirm('Bạn có chắc chắn muốn gỡ bỏ MCP Server này?')) return;
    this.api.deleteMCPServer(id).subscribe({
      next: () => {
        this.notify('Đã xóa MCP Server thành công');
        this.api.getMCPServers().subscribe(s => this.mcpServers = s);
      },
      error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
    });
  }

  // ---------------- CRUD Projects ----------------
  showProjectModal = false;
  isEditingProject = false;
  projectForm: Partial<Project> & { teamsText?: string } = {
    code: '',
    name: '',
    description: '',
    owner: '',
    status: 'ACTIVE',
    budget: 50,
    due: '',
    teamsText: 'platform, quality'
  };

  openCreateProjectModal(): void {
    this.isEditingProject = false;
    this.projectForm = {
      code: '',
      name: '',
      description: '',
      owner: this.employees.find(e => e.role === 'AI Manager' || e.role === 'Tech Lead')?.id || 'm2',
      status: 'ACTIVE',
      budget: 60,
      due: new Date(Date.now() + 60 * 86400000).toISOString().split('T')[0],
      teamsText: 'platform, quality'
    };
    this.showProjectModal = true;
  }

  openEditProjectModal(p: Project, event?: Event): void {
    if (event) event.stopPropagation();
    this.isEditingProject = true;
    this.projectForm = {
      ...p,
      teamsText: (p.teams || []).join(', ')
    };
    this.showProjectModal = true;
  }

  saveProject(): void {
    if (!this.projectForm.code || !this.projectForm.name || !this.projectForm.owner) {
      alert('Vui lòng nhập Mã dự án, Tên dự án và Chọn người chịu trách nhiệm (Owner)');
      return;
    }
    const teams = (this.projectForm.teamsText || '')
      .split(',')
      .map(s => s.trim())
      .filter(s => s.length > 0);

    const payload: Partial<Project> = {
      ...this.projectForm,
      teams
    };

    if (this.isEditingProject && this.projectForm.id) {
      this.api.updateProject(this.projectForm.id, payload).subscribe({
        next: () => {
          this.notify(`Đã cập nhật dự án ${this.projectForm.code}`);
          this.showProjectModal = false;
          this.api.getProjects().subscribe(p => this.projects = p);
          if (this.selectedProject && this.selectedProject.id === this.projectForm.id) {
            this.openProjectDetail(this.projectForm.id);
          }
        },
        error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
      });
    } else {
      this.api.createProject(payload).subscribe({
        next: () => {
          this.notify(`Đã tạo dự án mới ${this.projectForm.code}`);
          this.showProjectModal = false;
          this.api.getProjects().subscribe(p => this.projects = p);
        },
        error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
      });
    }
  }

  deleteProject(id: string, event?: Event): void {
    if (event) event.stopPropagation();
    if (!confirm('Bạn có chắc chắn muốn xóa/lưu trữ dự án này?')) return;
    this.api.deleteProject(id).subscribe({
      next: () => {
        this.notify('Đã xóa / lưu trữ dự án thành công');
        this.selectedProject = null;
        this.api.getProjects().subscribe(p => this.projects = p);
      },
      error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
    });
  }

  // ---------------- Org Hierarchy Reporting Line Reassignment ----------------
  showReassignModal = false;
  reassignEmployeeId = '';
  reassignManagerId: string | null = null;
  reassignError = '';

  openReassignModal(employeeId?: string, event?: Event): void {
    if (event) event.stopPropagation();
    this.reassignError = '';
    this.reassignEmployeeId = employeeId || (this.employees.find(e => e.id !== 'h1')?.id || '');
    const currentEmp = this.employees.find(e => e.id === this.reassignEmployeeId);
    this.reassignManagerId = currentEmp?.managerId || null;
    this.showReassignModal = true;
  }

  onReassignEmpChange(): void {
    this.reassignError = '';
    const currentEmp = this.employees.find(e => e.id === this.reassignEmployeeId);
    this.reassignManagerId = currentEmp?.managerId || null;
  }

  getValidManagersFor(empId: string): Employee[] {
    if (!empId) return this.employees;
    // Find all subordinates recursively
    const subordinates = new Set<string>();
    subordinates.add(empId);
    let added = true;
    while (added) {
      added = false;
      for (const e of this.employees) {
        if (e.managerId && subordinates.has(e.managerId) && !subordinates.has(e.id)) {
          subordinates.add(e.id);
          added = true;
        }
      }
    }
    // Valid managers are employees not in subordinates set
    return this.employees.filter(e => !subordinates.has(e.id));
  }

  submitReassign(): void {
    this.reassignError = '';
    if (!this.reassignEmployeeId) {
      this.reassignError = 'Vui lòng chọn nhân sự cần điều chuyển';
      return;
    }
    this.api.updateReportingLine(this.reassignEmployeeId, this.reassignManagerId).subscribe({
      next: () => {
        this.notify('Đã cập nhật tuyến báo cáo phân cấp thành công!');
        this.showReassignModal = false;
        this.api.getOrgChart().subscribe(tree => this.orgTree = tree);
        this.api.getEmployees().subscribe(e => this.employees = e);
      },
      error: (err) => {
        this.reassignError = err?.error?.error || err.message;
      }
    });
  }

  // ---------------- Departments & Teams Management ----------------
  showDeptTeamModal = false;
  activeDeptTeamTab: 'depts' | 'teams' = 'depts';
  isEditingDept = false;
  isEditingTeam = false;
  deptForm: Partial<Department> = { name: '' };
  teamForm: Partial<Team> = { name: '', departmentId: '' };

  openDeptTeamModal(): void {
    this.activeDeptTeamTab = 'depts';
    this.isEditingDept = false;
    this.isEditingTeam = false;
    this.deptForm = { name: '' };
    this.teamForm = { name: '', departmentId: this.departments[0]?.id || '' };
    this.showDeptTeamModal = true;
  }

  cancelEditDept(): void {
    this.isEditingDept = false;
    this.deptForm = { name: '' };
  }

  cancelEditTeam(): void {
    this.isEditingTeam = false;
    this.teamForm = { name: '', departmentId: this.departments[0]?.id || '' };
  }

  saveDepartment(): void {
    if (!this.deptForm.name) {
      alert('Vui lòng nhập tên phòng ban');
      return;
    }
    if (this.isEditingDept && this.deptForm.id) {
      this.api.updateDepartment(this.deptForm.id, this.deptForm).subscribe({
        next: () => {
          this.notify('Đã cập nhật phòng ban');
          this.isEditingDept = false;
          this.deptForm = { name: '' };
          this.api.getDepartments().subscribe(d => this.departments = d);
        },
        error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
      });
    } else {
      this.api.createDepartment(this.deptForm).subscribe({
        next: () => {
          this.notify('Đã tạo phòng ban mới');
          this.deptForm = { name: '' };
          this.api.getDepartments().subscribe(d => this.departments = d);
        },
        error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
      });
    }
  }

  deleteDepartment(id: string): void {
    if (!confirm('Bạn có chắc chắn muốn xóa phòng ban này?')) return;
    this.api.deleteDepartment(id).subscribe({
      next: () => {
        this.notify('Đã xóa phòng ban thành công');
        this.api.getDepartments().subscribe(d => this.departments = d);
      },
      error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
    });
  }

  saveTeam(): void {
    if (!this.teamForm.name || !this.teamForm.departmentId) {
      alert('Vui lòng nhập tên team và chọn phòng ban trực thuộc');
      return;
    }
    if (this.isEditingTeam && this.teamForm.id) {
      this.api.updateTeam(this.teamForm.id, this.teamForm).subscribe({
        next: () => {
          this.notify('Đã cập nhật team');
          this.isEditingTeam = false;
          this.teamForm = { name: '', departmentId: this.departments[0]?.id || '' };
          this.api.getTeams().subscribe(t => this.teams = t);
        },
        error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
      });
    } else {
      this.api.createTeam(this.teamForm).subscribe({
        next: () => {
          this.notify('Đã tạo team mới');
          this.teamForm = { name: '', departmentId: this.departments[0]?.id || '' };
          this.api.getTeams().subscribe(t => this.teams = t);
        },
        error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
      });
    }
  }

  deleteTeam(id: string): void {
    if (!confirm('Bạn có chắc chắn muốn xóa team này?')) return;
    this.api.deleteTeam(id).subscribe({
      next: () => {
        this.notify('Đã xóa team thành công');
        this.api.getTeams().subscribe(t => this.teams = t);
      },
      error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
    });
  }

  // ---------------- Workforce (Employees) Management ----------------
  showEmployeeModal = false;
  isEditingEmployee = false;
  employeeModalTab: 'basic' | 'model_budget' | 'mcp_skills' = 'basic';
  employeeForm: Partial<Employee> & { skillsText?: string } = {
    displayName: '',
    kind: 'AI',
    title: '',
    role: '',
    departmentId: '',
    managerId: '',
    model: 'Model X',
    provider: 'Provider A',
    budget: 30,
    skillsText: 'Automation, Code Review',
    mcp: ['task-mcp'],
    teams: ['platform']
  };

  openCreateEmployeeModal(): void {
    this.isEditingEmployee = false;
    this.employeeModalTab = 'basic';
    this.employeeForm = {
      displayName: '',
      kind: 'AI',
      title: 'AI Specialist',
      role: 'Engineer',
      departmentId: this.departments[0]?.id || 'tech',
      managerId: 'm2',
      model: 'Model X',
      provider: 'Provider A',
      budget: 30,
      skillsText: 'Automation, Analysis',
      mcp: ['task-mcp', 'git-mcp'],
      teams: ['platform']
    };
    this.showEmployeeModal = true;
  }

  openEditEmployeeModal(e: Employee, event?: Event): void {
    if (event) event.stopPropagation();
    this.isEditingEmployee = true;
    this.employeeModalTab = 'basic';
    this.employeeForm = {
      ...e,
      skillsText: (e.skills || []).join(', ')
    };
    this.showEmployeeModal = true;
  }

  saveEmployee(): void {
    if (!this.employeeForm.displayName || !this.employeeForm.title) {
      alert('Vui lòng nhập Họ tên và Chức danh');
      return;
    }
    const skills = (this.employeeForm.skillsText || '')
      .split(',')
      .map(s => s.trim())
      .filter(s => s.length > 0);

    const payload: Partial<Employee> = {
      ...this.employeeForm,
      name: this.employeeForm.displayName,
      skills
    };

    if (this.isEditingEmployee && this.employeeForm.id) {
      this.api.updateEmployee(this.employeeForm.id, payload).subscribe({
        next: () => {
          this.notify(`Đã cập nhật nhân sự ${this.employeeForm.displayName}`);
          this.showEmployeeModal = false;
          this.api.getEmployees().subscribe(e => this.employees = e);
          this.api.getOrgChart().subscribe(tree => this.orgTree = tree);
          if (this.selectedEmployee && this.selectedEmployee.id === this.employeeForm.id) {
            this.openEmployeeDetail(this.employeeForm.id);
          }
        },
        error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
      });
    } else {
      this.api.createEmployee(payload).subscribe({
        next: () => {
          this.notify(`Đã thêm nhân sự mới ${this.employeeForm.displayName}`);
          this.showEmployeeModal = false;
          this.api.getEmployees().subscribe(e => this.employees = e);
          this.api.getOrgChart().subscribe(tree => this.orgTree = tree);
        },
        error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
      });
    }
  }

  deactivateEmployee(id: string, event?: Event): void {
    if (event) event.stopPropagation();
    if (!confirm('Bạn có chắc chắn muốn vô hiệu hóa nhân sự này?')) return;
    this.api.deleteEmployee(id).subscribe({
      next: () => {
        this.notify('Đã vô hiệu hóa nhân sự thành công');
        this.api.getEmployees().subscribe(e => this.employees = e);
        this.api.getOrgChart().subscribe(tree => this.orgTree = tree);
      },
      error: (err) => alert(`Lỗi: ${err?.error?.error || err.message}`)
    });
  }

  toggleMCPGrant(mcpName: string): void {
    if (!this.employeeForm.mcp) {
      this.employeeForm.mcp = [];
    }
    const idx = this.employeeForm.mcp.indexOf(mcpName);
    if (idx > -1) {
      this.employeeForm.mcp.splice(idx, 1);
    } else {
      this.employeeForm.mcp.push(mcpName);
    }
  }

  isMCPGranted(mcpName: string): boolean {
    return !!this.employeeForm.mcp?.includes(mcpName);
  }

  closeOverlays(): void {
    this.selectedEmployee = null;
    this.selectedTask = null;
    this.selectedProject = null;
    this.showCreateTaskModal = false;
    this.showAuthModal = false;
    this.showModelModal = false;
    this.showMCPModal = false;
    this.showProjectModal = false;
    this.showReassignModal = false;
    this.showDeptTeamModal = false;
    this.showEmployeeModal = false;
  }
}
