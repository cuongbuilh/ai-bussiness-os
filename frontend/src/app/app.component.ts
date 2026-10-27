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
type EmployeeTab = 'overview' | 'tasks' | 'tools' | 'kpis';

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
  selectedEmpTab: EmployeeTab = 'overview';
  selectedEmpCaps: any = null;

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
      } else if (event.type === 'employee') {
        this.api.getEmployees().subscribe(e => this.employees = e);
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
  openEmployeeDetail(id: string, tab: EmployeeTab = 'overview'): void {
    this.api.getEmployeeDetail(id).subscribe(emp => {
      this.selectedEmployee = emp;
      this.selectedEmpTab = tab;
      this.api.getEmployeeCapabilities(id).subscribe(caps => {
        this.selectedEmpCaps = caps;
      });
    });
  }

  setEmpTab(tab: EmployeeTab): void {
    this.selectedEmpTab = tab;
  }

  getEmployeeTasks(empId: string): Task[] {
    return this.tasks.filter(t => t.owner === empId || t.collab?.includes(empId));
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

  closeOverlays(): void {
    this.selectedEmployee = null;
    this.selectedTask = null;
    this.selectedProject = null;
    this.showCreateTaskModal = false;
    this.showAuthModal = false;
  }
}
