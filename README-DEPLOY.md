# AI Business OS — Hướng dẫn Triển khai & Vận hành (Golang + Angular + Docker Compose)

Hệ thống **AI Business OS** đã được triển khai đầy đủ mã nguồn bao gồm:
1. **Backend (Golang)**: Control Plane & Agent Execution Engine (`/backend`)
2. **Frontend (Angular 18)**: Web Console điều hành doanh nghiệp AI (`/frontend`)
3. **Database (PostgreSQL 16)**: Lưu trữ dữ liệu quan hệ, di trú lược đồ (`/backend/migrations`)
4. **Orchestration (Docker Compose)**: Cấu hình container hóa toàn bộ hệ thống (`docker-compose.yml`)

---

## 1. Cấu trúc mã nguồn

```
ai-business-os-spec-v1.1/
├── backend/
│   ├── cmd/server/main.go            # Entrypoint REST API & SSE Server
│   ├── internal/
│   │   ├── domain/models.go          # Entity, Enum, và cấu trúc dữ liệu domain
│   │   ├── repository/               # PostgreSQL & In-Memory fallback repository
│   │   ├── service/                  # Business logic (Org, Task, Approval, Execution, EventBroker)
│   │   └── handler/handlers.go       # HTTP Handlers (Chuẩn OpenAPI /api/v1/...)
│   ├── migrations/
│   │   ├── 001_init.sql              # Khởi tạo bảng, enum, ràng buộc
│   │   └── 002_seed.sql              # Dữ liệu mẫu canonical ban đầu
│   ├── Dockerfile                    # Multi-stage Golang build (alpine)
│   ├── go.mod & go.sum
│   └── ...
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   │   ├── models/               # TypeScript interfaces
│   │   │   ├── services/api.service.ts # HTTP Client & Realtime SSE Stream
│   │   │   ├── app.component.ts      # Component điều phối chính
│   │   │   └── app.component.html    # Giao diện Web Console tương tác
│   │   ├── styles.scss               # Cyber-Executive Dark Theme
│   │   └── index.html
│   ├── nginx.conf                    # Cấu hình Nginx reverse-proxy /api & SSE
│   ├── Dockerfile                    # Multi-stage Angular build + Nginx Alpine
│   ├── proxy.conf.json               # Cấu hình proxy cho local development
│   └── package.json
├── docker-compose.yml                # Docker compose orchestrating all 3 services
├── .env.example                      # Mẫu biến môi trường
└── README-DEPLOY.md                  # Tài liệu hướng dẫn này
```

---

## 2. Khởi chạy nhanh bằng Docker Compose

Yêu cầu: Đã cài đặt Docker và Docker Compose.

```bash
# 1. Khởi chạy toàn bộ hệ thống (PostgreSQL, Golang Backend, Angular Frontend)
docker compose up --build -d

# 2. Kiểm tra trạng thái các container
docker compose ps
```

Khi chạy thành công:
- **Web Console (Angular)**: [http://localhost:3000](http://localhost:3000)
- **Control Plane API (Golang)**: [http://localhost:8080/api/v1/health](http://localhost:8080/api/v1/health)
- **PostgreSQL Database**: `localhost:5432` (User: `abos`, Database: `abos_db`)

---

## 3. Xác thực & Đăng nhập Hệ thống (Authentication)

Hệ thống bảo vệ toàn bộ các endpoint nghiệp vụ bằng cơ chế **Bearer Token Authentication** và cung cấp trang đăng nhập trực quan trên Web Console:

### 3.1. Đăng nhập qua Web Console ([http://localhost:3000](http://localhost:3000))
Khi truy cập Web Console lần đầu, hệ thống hiển thị màn hình đăng nhập bảo mật:
- **Tài khoản mặc định (Founder / CEO):**
  - Tên đăng nhập: `founder` (hoặc `admin`, `h1`)
  - Mật khẩu: `founder123`
  - Hoặc bấm nút **⚡ Đăng nhập nhanh tài khoản Founder / CEO** để tự động xác thực.
- **Đăng nhập bằng Access Token:** Có thể chuyển sang tab *Access Token* để nhập trực tiếp token được cấp (ví dụ: `founder-h1-token`).
- **Đăng xuất:** Trên thanh toolbar trên cùng, hiển thị huy hiệu người dùng hiện tại (ví dụ: `[FC] Founder / CEO`), nhấp vào nút **Đăng xuất** để xóa phiên làm việc và đóng luồng kết nối thời gian thực.

### 3.2. Đăng nhập qua REST API
```bash
# 1. Đăng nhập lấy Bearer token
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"founder","password":"founder123"}'

# Phản hồi mẫu:
# {
#   "token": "founder-h1-token",
#   "user": {
#     "id": "h1",
#     "displayName": "Founder",
#     "role": "Executive",
#     "kind": "HUMAN",
#     "title": "Founder / CEO"
#   }
# }

# 2. Truy vấn thông tin người dùng hiện tại
curl -s http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer founder-h1-token"

# 3. Sử dụng token cho các API điều hành (bắt buộc qua HTTP Header)
curl -s http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer founder-h1-token"
```

> **Lưu ý bảo mật:** Token bắt buộc phải gửi qua header `Authorization: Bearer <token>`. Hệ thống từ chối mọi yêu cầu truyền token qua URL query parameter (`?token=...`) để triệt tiêu nguy cơ rò rỉ vào access log và proxy log. Mọi query parameter nhạy cảm đều được tự động redact thành `[REDACTED]` trong log máy chủ.

---

## 3. Khởi chạy trong môi trường phát triển (Local Development)

Nếu muốn chạy trực tiếp không qua Docker:

### Bước 1: Khởi chạy Backend (Golang)
```bash
cd backend
go run ./cmd/server
```
*(Nếu không có `DATABASE_URL`, backend sẽ tự động dùng In-Memory repository có sẵn toàn bộ dữ liệu mẫu ban đầu mà không cần cài đặt Postgres)*.

### Bước 2: Khởi chạy Frontend (Angular)
```bash
cd frontend
npm install
npm start
```
Mở trình duyệt tại [http://localhost:4200](http://localhost:4200). Các yêu cầu `/api` sẽ được tự động proxy sang backend tại `http://localhost:8080`.

---

## 4. Các tính năng chính trên Web Console

1. **Tổng quan (Overview)**: Thẻ KPI nhân sự AI, dự án, tasks đang thực hiện, yêu cầu chờ duyệt, thanh tiến độ nghiệm thu và chi phí thực tế.
2. **Sơ đồ tổ chức (Organization Tree)**: Cây phân cấp trực quan hiển thị Founder (HUMAN) và các cấp quản lý / nhân viên AI (1 direct manager / employee). Click vào từng nhân sự để mở chi tiết.
3. **Unified Work Board**:
   - Bộ lọc phạm vi 7 cấp (Scope): Toàn công ty, Phòng ban, Team, Dự án, Nhân sự, Quản lý, Vai trò.
   - 2 chế độ hiển thị: **Kanban Board** và **Bảng dữ liệu (Table)**.
   - Các trạng thái task: `Backlog`, `Ready`, `In progress`, `Blocked`, `In review`, `Completed`, `Accepted`.
4. **Thực thi AI Task (AI Execution Engine)**:
   - Trong drawer chi tiết task, bấm **▶ Chạy AI Execution** để agent tự động giải quyết task thông qua các bước mô phỏng MCP Tool và LLM Model.
   - Xem log thực thi, số token và chi phí USD phát sinh cập nhật trực tiếp vào sổ cái ngân sách.
5. **Nghiệm thu công việc (Task Acceptance)**:
   - Bấm **✓ Nghiệm thu Task** để xác nhận kết quả đạt chuẩn theo quy tắc `BR-WORK-003`.
6. **Danh mục Dự án (Project Portfolio)**: Thẻ dự án, tiến độ hoàn thành các mốc (Milestones), ngân sách trần và đã dùng.
7. **Models, MCP & Skills**: Danh mục Model Providers, MCP Servers (`web-mcp`, `git-mcp`, `shell-mcp`, `budget-mcp`, `docs-mcp`, `task-mcp`), và công thức phân quyền trần vị trí (Position Ceiling ∩ Agent Grant).
8. **Hộp thư Phê duyệt (Approval Inbox)**: Phê duyệt (Approve) hoặc Từ chối (Reject) các hành động nhạy cảm với cơ chế kiểm tra chống tự phê duyệt (`BR-AUTH-001`).
9. **Cập nhật thời gian thực (SSE Stream)**: Giao diện tự động cập nhật ngay lập tức khi có task mới, thay đổi trạng thái hoặc hoàn tất execution mà không cần reload trang.
