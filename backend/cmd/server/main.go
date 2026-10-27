package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cuongbuilh/ai-business-os/backend/internal/handler"
	"github.com/cuongbuilh/ai-business-os/backend/internal/repository"
	"github.com/cuongbuilh/ai-business-os/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	var repo repository.Repository

	if databaseURL != "" {
		log.Printf("Connecting to PostgreSQL at %s...", databaseURL)
		pgRepo, err := repository.NewPostgresRepo(databaseURL)
		if err != nil {
			log.Printf("Warning: Failed to initialize PostgreSQL (%v). Falling back to In-Memory repository.", err)
			repo = repository.NewMemoryRepo()
		} else {
			log.Printf("Connected to PostgreSQL successfully.")
			repo = pgRepo
		}
	} else {
		log.Printf("DATABASE_URL not set. Using In-Memory repository with canonical seed data.")
		repo = repository.NewMemoryRepo()
	}

	broker := service.NewEventBroker()
	svc := service.NewServices(repo, broker)
	h := handler.NewHandler(svc)

	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	// Sanitize sensitive query parameters (token, secret, key, password) before logger to prevent log leakage
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if q := r.URL.Query(); len(q) > 0 {
				sensitiveKeys := []string{"token", "secret", "key", "password", "auth", "authorization"}
				redacted := false
				for _, k := range sensitiveKeys {
					if q.Has(k) {
						q.Set(k, "[REDACTED]")
						redacted = true
					}
				}
				if redacted {
					r.URL.RawQuery = q.Encode()
					r.RequestURI = r.URL.RequestURI()
				}
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration for local Angular dev server and Docker Compose
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "If-Match"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.HealthCheck)

		// Organization & Org Chart
		r.Get("/organization/org-chart", h.GetOrgChart)
		r.Get("/departments", h.GetDepartments)
		r.Get("/teams", h.GetTeams)
		r.Get("/positions", h.GetPositions)

		// Employees
		r.Get("/employees", h.ListEmployees)
		r.Post("/employees", h.CreateEmployee)
		r.Get("/employees/{employeeId}", h.GetEmployeeDetail)
		r.Get("/employees/{employeeId}/capabilities", h.GetEmployeeCapabilities)

		// Projects
		r.Get("/projects", h.ListProjects)
		r.Post("/projects", h.CreateProject)
		r.Get("/projects/{projectId}", h.GetProjectDetail)

		// Tasks
		r.Get("/tasks", h.ListTasks)
		r.Post("/tasks", h.CreateTask)
		r.Get("/tasks/{taskId}", h.GetTaskDetail)
		r.Patch("/tasks/{taskId}", h.UpdateTaskStatus)
		r.Post("/tasks/{taskId}/assign", h.AssignTask)
		r.Post("/tasks/{taskId}/accept", h.AcceptTask)
		r.Get("/tasks/{taskId}/executions", h.GetTaskExecutions)
		r.Post("/tasks/{taskId}/execute", h.ExecuteTask)

		// Boards & Realtime
		r.Get("/boards/summary", h.GetBoardSummary)
		r.Get("/boards/stream", h.StreamBoardUpdates)

		// Approvals
		r.Get("/approvals", h.ListApprovals)
		r.Post("/approvals/{approvalId}/decide", h.DecideApproval)

		// Capabilities
		r.Get("/capabilities/mcp-servers", h.GetMCPServers)
		r.Get("/capabilities/models", h.GetModels)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
	}

	go func() {
		log.Printf("AI Business OS Backend running on http://0.0.0.0:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server listen error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down AI Business OS backend server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully.")
}
