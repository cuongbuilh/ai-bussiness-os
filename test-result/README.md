# Kết quả khắc phục và kiểm định toàn diện — 2026-10-27

## Kết quả kiểm định sau khắc phục

| Bộ kiểm tra | Trạng thái | Bằng chứng |
|---|---|---|
| 10 ca QA-API-01…10 | **10 PASS** | [api-tests.jsonl](api-tests.jsonl) |
| 7 unit test service | **7 PASS** | [full-go-tests.log](full-go-tests.log) |
| 12 ca hồi quy QA-RV-01…12 | **12 PASS** | [reverification-tests.jsonl](reverification-tests.jsonl) |
| 1 ca concurrency Postgres đa instance | **1 PASS** | [full-go-tests.log](full-go-tests.log) |
| 21 ca Compose/PostgreSQL QA-CP-01…21 | **21 PASS, 0 FAIL** | [compose/results.json](compose/results.json) |
| Angular production build | **PASS (0 token leak)** | [frontend-build.log](frontend-build.log) |
| Docker Compose build frontend & backend | **PASS** | [compose/frontend-docker-build.log](compose/frontend-docker-build.log) |

Lệnh kiểm tra:

```bash
cd backend
GOCACHE=/tmp/abos-qa-gocache go test ./internal/handler -run TestQA_API -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./internal/handler -run TestQA_RV -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./internal/repository -run TestPostgres -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./internal/service -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test -race ./... -count=1 -v
cd ../frontend && npm run build
docker compose build
python3 test-result/compose/evaluate.py
```

Toàn bộ test suite Go đạt 100% PASS (30/30 unit & integration tests). Đánh giá 21 ca Compose/PostgreSQL đạt **21 PASS, 0 FAIL**.

## Khắc phục chuyên sâu các lỗi xác thực, phân quyền & lưu trữ ngân sách

1. **Loại bỏ token khỏi Frontend Build (Zero Token Leak)**:
   - Trong [api.service.ts](file:///home/cuongbv/sources/cuongbuilh/ai-business-os-spec-v1.1/frontend/src/app/services/api.service.ts), token được đọc động từ `localStorage` thông qua `getAuthToken()`, `setAuthToken()`, `clearAuthToken()`.
   - Loại bỏ toàn bộ hardcoded header `founder-h1-token` và placeholder text nhạy cảm trong template.
   - Thêm modal cấu hình Bearer token trực quan trên thanh toolbar của [app.component.html](file:///home/cuongbv/sources/cuongbuilh/ai-business-os-spec-v1.1/frontend/src/app/app.component.html).
   - Kiểm tra `grep -rn "founder-h1-token" frontend/dist/` trả về **0 kết quả**.

2. **Bảo vệ toàn diện các endpoint đọc dữ liệu trước truy cập Vô danh (Anonymous)**:
   - Tất cả các endpoint đọc (`GET /employees/{id}`, `GET /employees/{id}/capabilities`, `GET /projects`, `GET /projects/{id}`, `GET /departments`, `GET /teams`, `GET /positions`, `GET /capabilities/mcp-servers`, `GET /capabilities/models`) đều bắt buộc xác thực Bearer token. Yêu cầu vô danh đều trả về **HTTP 401 Unauthorized**.
   - Endpoint realtime Server-Sent Events `GET /boards/stream` chặn kết nối vô danh (401), hỗ trợ xác thực client thông qua query parameter `?token=...`.

3. **Chặn tạo Project vô danh trước khi parse body**:
   - Trong [handlers.go](file:///home/cuongbv/sources/cuongbuilh/ai-business-os-spec-v1.1/backend/internal/handler/handlers.go), thao tác `POST /projects` thực hiện `authenticate(r)` ngay dòng đầu tiên. Yêu cầu vô danh với payload `{}` lập tức trả về **HTTP 401 Unauthorized** thay vì 400 Bad Request.

4. **Giới hạn ngân sách an toàn đa Instance bằng khóa dòng PostgreSQL (Atomic Row Locking)**:
   - Triển khai phương thức `AtomicReserveBudget` trong [postgres.go](file:///home/cuongbv/sources/cuongbuilh/ai-business-os-spec-v1.1/backend/internal/repository/postgres.go) sử dụng transaction `SELECT ... FROM employees WHERE id = $1 FOR UPDATE`.
   - Cơ chế khóa dòng mức DB đảm bảo tuần tự hóa các yêu cầu đặt chỗ ngân sách giữa nhiều backend instances chạy đồng thời.
   - Kiểm thử thực tế qua `TestPostgres_MultiInstance_AtomicBudgetConcurrency` với 2 instance backend kết nối đồng thời vào PostgreSQL thật: chặn tuyệt đối hiện tượng vượt trần ngân sách ($0.50 cap chặn thành công 2 request $0.37 đồng thời).
   - Xử lý triệt để tất cả lỗi DB (không bỏ qua lỗi khi ghi reservation, cập nhật employee/task/execution, và ghi usage ledger).

## Trạng thái dữ liệu hệ thống
- **Database chính (`abos_db`)**: Giữ nguyên vẹn dữ liệu chuẩn (seed canonical): 9 nhân sự, 15 task, 3 dự án.
- **Database QA (`abos_qa_20261027`)**: Chạy kiểm thử write-flow, parallel creation, xác thực budget reservation và ghi usage ledger thực tế trên PostgreSQL.
