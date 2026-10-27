import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';
import {
  Department, Team, Position, Employee, Project, Task,
  Execution, ApprovalRequest, MCPServer, ModelInfo, OrgNode, BoardSummary,
  AuthUser, LoginResponse
} from '../models/abos.models';

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private http = inject(HttpClient);
  // Default to relative /api/v1 (proxied by Nginx in Docker or Angular CLI dev proxy)
  private baseUrl = '/api/v1';
  private tokenKey = 'abos_token';
  private userKey = 'abos_user';

  getAuthToken(): string {
    if (typeof localStorage === 'undefined') return '';
    return localStorage.getItem(this.tokenKey) || '';
  }

  setAuthToken(token: string): void {
    if (typeof localStorage !== 'undefined') {
      if (token) {
        localStorage.setItem(this.tokenKey, token.trim());
      } else {
        localStorage.removeItem(this.tokenKey);
      }
    }
    this.restartSSE();
  }

  getCurrentUser(): AuthUser | null {
    if (typeof localStorage === 'undefined') return null;
    const raw = localStorage.getItem(this.userKey);
    if (!raw) return null;
    try {
      return JSON.parse(raw);
    } catch {
      return null;
    }
  }

  setCurrentUser(user: AuthUser | null): void {
    if (typeof localStorage !== 'undefined') {
      if (user) {
        localStorage.setItem(this.userKey, JSON.stringify(user));
      } else {
        localStorage.removeItem(this.userKey);
      }
    }
  }

  isAuthenticated(): boolean {
    return !!this.getAuthToken();
  }

  clearAuthToken(): void {
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem(this.tokenKey);
      localStorage.removeItem(this.userKey);
    }
    this.restartSSE();
  }

  // ---------------- Authentication ----------------
  login(credentials: { username?: string; password?: string; token?: string }): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(`${this.baseUrl}/auth/login`, credentials);
  }

  getMe(): Observable<AuthUser> {
    return this.http.get<AuthUser>(`${this.baseUrl}/auth/me`, { headers: this.headers });
  }

  private get headers(): HttpHeaders {
    const token = this.getAuthToken();
    if (token) {
      return new HttpHeaders({
        'Authorization': `Bearer ${token}`
      });
    }
    return new HttpHeaders();
  }

  private sseEvent$ = new Subject<{ type: string; action: string; payload: any }>();
  private abortController: AbortController | null = null;
  private sseRetryTimer: any = null;

  constructor() {
    this.initSSE();
  }

  getEvents(): Observable<{ type: string; action: string; payload: any }> {
    return this.sseEvent$.asObservable();
  }

  restartSSE(): void {
    this.initSSE();
  }

  private async initSSE(): Promise<void> {
    if (typeof window === 'undefined' || typeof fetch === 'undefined') return;

    if (this.abortController) {
      this.abortController.abort();
      this.abortController = null;
    }
    if (this.sseRetryTimer) {
      clearTimeout(this.sseRetryTimer);
      this.sseRetryTimer = null;
    }

    const token = this.getAuthToken();
    if (!token) {
      return;
    }

    this.abortController = new AbortController();

    try {
      // Secure SSE connection passing token strictly via HTTP Authorization header (never in URL)
      const response = await fetch(`${this.baseUrl}/boards/stream`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Accept': 'text/event-stream'
        },
        signal: this.abortController.signal
      });

      if (!response.ok || !response.body) {
        this.sseRetryTimer = setTimeout(() => this.initSSE(), 5000);
        return;
      }

      // F-08: On initial connect or reconnect, emit refetch event to refresh snapshot
      this.sseEvent$.next({ type: 'system', action: 'refetch_snapshot', payload: null });

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { value, done } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        const blocks = buffer.split('\n\n');
        buffer = blocks.pop() || '';

        for (const block of blocks) {
          const match = block.match(/data:\s*(.+)/);
          if (match) {
            try {
              const data = JSON.parse(match[1]);
              this.sseEvent$.next(data);
            } catch {
              // ignore ping or non-json message
            }
          }
        }
      }
    } catch (err: any) {
      if (err?.name !== 'AbortError') {
        this.sseRetryTimer = setTimeout(() => this.initSSE(), 5000);
      }
    }
  }

  // ---------------- Organization ----------------
  getOrgChart(): Observable<OrgNode> {
    return this.http.get<OrgNode>(`${this.baseUrl}/organization/org-chart`, { headers: this.headers });
  }

  getDepartments(): Observable<Department[]> {
    return this.http.get<Department[]>(`${this.baseUrl}/departments`, { headers: this.headers });
  }

  getTeams(): Observable<Team[]> {
    return this.http.get<Team[]>(`${this.baseUrl}/teams`, { headers: this.headers });
  }

  getPositions(): Observable<Position[]> {
    return this.http.get<Position[]>(`${this.baseUrl}/positions`, { headers: this.headers });
  }

  // ---------------- Employees ----------------
  getEmployees(): Observable<Employee[]> {
    return this.http.get<Employee[]>(`${this.baseUrl}/employees`, { headers: this.headers });
  }

  getEmployeeDetail(id: string): Observable<Employee> {
    return this.http.get<Employee>(`${this.baseUrl}/employees/${id}`, { headers: this.headers });
  }

  createEmployee(emp: Partial<Employee>): Observable<Employee> {
    return this.http.post<Employee>(`${this.baseUrl}/employees`, emp, { headers: this.headers });
  }

  getEmployeeCapabilities(id: string): Observable<any> {
    return this.http.get<any>(`${this.baseUrl}/employees/${id}/capabilities`, { headers: this.headers });
  }

  // ---------------- Projects ----------------
  getProjects(): Observable<Project[]> {
    return this.http.get<Project[]>(`${this.baseUrl}/projects`, { headers: this.headers });
  }

  getProjectDetail(id: string): Observable<Project> {
    return this.http.get<Project>(`${this.baseUrl}/projects/${id}`, { headers: this.headers });
  }

  createProject(proj: Partial<Project>): Observable<Project> {
    return this.http.post<Project>(`${this.baseUrl}/projects`, proj, { headers: this.headers });
  }

  // ---------------- Tasks ----------------
  getTasks(scope = '', project = '', owner = '', status = '', search = ''): Observable<Task[]> {
    let params = new HttpParams();
    if (scope) params = params.set('scope', scope);
    if (project) params = params.set('project', project);
    if (owner) params = params.set('owner', owner);
    if (status) params = params.set('status', status);
    if (search) params = params.set('search', search);

    return this.http.get<Task[]>(`${this.baseUrl}/tasks`, { headers: this.headers, params });
  }

  getTaskDetail(id: string): Observable<Task> {
    return this.http.get<Task>(`${this.baseUrl}/tasks/${id}`, { headers: this.headers });
  }

  createTask(task: Partial<Task>): Observable<Task> {
    return this.http.post<Task>(`${this.baseUrl}/tasks`, task, { headers: this.headers });
  }

  updateTaskStatus(taskId: string, status: string): Observable<any> {
    return this.http.patch(`${this.baseUrl}/tasks/${taskId}`, { status }, { headers: this.headers });
  }

  assignTask(taskId: string, ownerId: string, collaborators: string[] = []): Observable<any> {
    return this.http.post(`${this.baseUrl}/tasks/${taskId}/assign`, { ownerId, collaborators }, { headers: this.headers });
  }

  acceptTask(taskId: string, comments = ''): Observable<any> {
    return this.http.post(`${this.baseUrl}/tasks/${taskId}/accept`, { comments }, { headers: this.headers });
  }

  // ---------------- Executions ----------------
  executeTask(taskId: string): Observable<Execution> {
    return this.http.post<Execution>(`${this.baseUrl}/tasks/${taskId}/execute`, {}, { headers: this.headers });
  }

  getTaskExecutions(taskId: string): Observable<Execution[]> {
    return this.http.get<Execution[]>(`${this.baseUrl}/tasks/${taskId}/executions`, { headers: this.headers });
  }

  getEmployeeExecutions(employeeId: string): Observable<Execution[]> {
    return this.http.get<Execution[]>(`${this.baseUrl}/employees/${employeeId}/executions`, { headers: this.headers });
  }

  // ---------------- Approvals ----------------
  getApprovals(): Observable<ApprovalRequest[]> {
    return this.http.get<ApprovalRequest[]>(`${this.baseUrl}/approvals`, { headers: this.headers });
  }

  decideApproval(approvalId: string, decision: 'APPROVE' | 'REJECT', resolvedBy = 'h1'): Observable<any> {
    return this.http.post(`${this.baseUrl}/approvals/${approvalId}/decide`, { decision, resolvedBy }, { headers: this.headers });
  }

  // ---------------- Board Summary & Capabilities ----------------
  getBoardSummary(scope = ''): Observable<BoardSummary> {
    let params = new HttpParams();
    if (scope) params = params.set('scope', scope);
    return this.http.get<BoardSummary>(`${this.baseUrl}/boards/summary`, { headers: this.headers, params });
  }

  getMCPServers(): Observable<MCPServer[]> {
    return this.http.get<MCPServer[]>(`${this.baseUrl}/capabilities/mcp-servers`, { headers: this.headers });
  }

  getModels(): Observable<ModelInfo[]> {
    return this.http.get<ModelInfo[]>(`${this.baseUrl}/capabilities/models`, { headers: this.headers });
  }
}
