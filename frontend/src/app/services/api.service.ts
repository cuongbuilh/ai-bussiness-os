import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';
import {
  Department, Team, Position, Employee, Project, Task,
  Execution, ApprovalRequest, MCPServer, ModelInfo, OrgNode, BoardSummary
} from '../models/abos.models';

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private http = inject(HttpClient);
  // Default to relative /api/v1 (proxied by Nginx in Docker or Angular CLI dev proxy)
  private baseUrl = '/api/v1';

  private sseEvent$ = new Subject<{ type: string; action: string; payload: any }>();
  private eventSource: EventSource | null = null;

  constructor() {
    this.initSSE();
  }

  getEvents(): Observable<{ type: string; action: string; payload: any }> {
    return this.sseEvent$.asObservable();
  }

  private initSSE(): void {
    if (typeof EventSource === 'undefined') return;

    try {
      this.eventSource = new EventSource(`${this.baseUrl}/boards/stream`);
      
      this.eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.sseEvent$.next(data);
        } catch {
          // ignore ping
        }
      };

      const eventTypes = ['task', 'employee', 'project', 'approval', 'execution'];
      for (const t of eventTypes) {
        this.eventSource.addEventListener(t, (e: MessageEvent) => {
          try {
            const data = JSON.parse(e.data);
            this.sseEvent$.next(data);
          } catch (err) {
            console.error('SSE parse error:', err);
          }
        });
      }

      this.eventSource.onerror = () => {
        // Retry after delay
        if (this.eventSource) {
          this.eventSource.close();
          setTimeout(() => this.initSSE(), 5000);
        }
      };
    } catch (e) {
      console.warn('Could not connect SSE stream:', e);
    }
  }

  // ---------------- Organization ----------------
  getOrgChart(): Observable<OrgNode> {
    return this.http.get<OrgNode>(`${this.baseUrl}/organization/org-chart`);
  }

  getDepartments(): Observable<Department[]> {
    return this.http.get<Department[]>(`${this.baseUrl}/departments`);
  }

  getTeams(): Observable<Team[]> {
    return this.http.get<Team[]>(`${this.baseUrl}/teams`);
  }

  getPositions(): Observable<Position[]> {
    return this.http.get<Position[]>(`${this.baseUrl}/positions`);
  }

  // ---------------- Employees ----------------
  getEmployees(): Observable<Employee[]> {
    return this.http.get<Employee[]>(`${this.baseUrl}/employees`);
  }

  getEmployeeDetail(id: string): Observable<Employee> {
    return this.http.get<Employee>(`${this.baseUrl}/employees/${id}`);
  }

  createEmployee(emp: Partial<Employee>): Observable<Employee> {
    return this.http.post<Employee>(`${this.baseUrl}/employees`, emp);
  }

  getEmployeeCapabilities(id: string): Observable<any> {
    return this.http.get<any>(`${this.baseUrl}/employees/${id}/capabilities`);
  }

  // ---------------- Projects ----------------
  getProjects(): Observable<Project[]> {
    return this.http.get<Project[]>(`${this.baseUrl}/projects`);
  }

  getProjectDetail(id: string): Observable<Project> {
    return this.http.get<Project>(`${this.baseUrl}/projects/${id}`);
  }

  createProject(proj: Partial<Project>): Observable<Project> {
    return this.http.post<Project>(`${this.baseUrl}/projects`, proj);
  }

  // ---------------- Tasks ----------------
  getTasks(scope = '', project = '', owner = '', status = '', search = ''): Observable<Task[]> {
    let params = new HttpParams();
    if (scope) params = params.set('scope', scope);
    if (project) params = params.set('project', project);
    if (owner) params = params.set('owner', owner);
    if (status) params = params.set('status', status);
    if (search) params = params.set('search', search);

    return this.http.get<Task[]>(`${this.baseUrl}/tasks`, { params });
  }

  getTaskDetail(id: string): Observable<Task> {
    return this.http.get<Task>(`${this.baseUrl}/tasks/${id}`);
  }

  createTask(task: Partial<Task>): Observable<Task> {
    return this.http.post<Task>(`${this.baseUrl}/tasks`, task);
  }

  updateTaskStatus(taskId: string, status: string): Observable<any> {
    return this.http.patch(`${this.baseUrl}/tasks/${taskId}`, { status });
  }

  assignTask(taskId: string, ownerId: string, collaborators: string[] = []): Observable<any> {
    return this.http.post(`${this.baseUrl}/tasks/${taskId}/assign`, { ownerId, collaborators });
  }

  acceptTask(taskId: string, comments = ''): Observable<any> {
    return this.http.post(`${this.baseUrl}/tasks/${taskId}/accept`, { comments });
  }

  // ---------------- Executions ----------------
  executeTask(taskId: string): Observable<Execution> {
    return this.http.post<Execution>(`${this.baseUrl}/tasks/${taskId}/execute`, {});
  }

  getTaskExecutions(taskId: string): Observable<Execution[]> {
    return this.http.get<Execution[]>(`${this.baseUrl}/tasks/${taskId}/executions`);
  }

  // ---------------- Approvals ----------------
  getApprovals(): Observable<ApprovalRequest[]> {
    return this.http.get<ApprovalRequest[]>(`${this.baseUrl}/approvals`);
  }

  decideApproval(approvalId: string, decision: 'APPROVE' | 'REJECT', resolvedBy = 'h1'): Observable<any> {
    return this.http.post(`${this.baseUrl}/approvals/${approvalId}/decide`, { decision, resolvedBy });
  }

  // ---------------- Board Summary & Capabilities ----------------
  getBoardSummary(scope = ''): Observable<BoardSummary> {
    let params = new HttpParams();
    if (scope) params = params.set('scope', scope);
    return this.http.get<BoardSummary>(`${this.baseUrl}/boards/summary`, { params });
  }

  getMCPServers(): Observable<MCPServer[]> {
    return this.http.get<MCPServer[]>(`${this.baseUrl}/capabilities/mcp-servers`);
  }

  getModels(): Observable<ModelInfo[]> {
    return this.http.get<ModelInfo[]>(`${this.baseUrl}/capabilities/models`);
  }
}
