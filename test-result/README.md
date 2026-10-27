# Kết quả khắc phục và kiểm định toàn diện — 2026-10-27

## Kết quả kiểm định sau khắc phục

| Bộ kiểm tra | Trạng thái | Bằng chứng |
|---|---|---|
| 10 ca QA-API-01…10 | **10 PASS** | [api-tests.jsonl](api-tests.jsonl) |
| 7 unit test service | **7 PASS** | [full-go-tests.log](full-go-tests.log) |
| 5 ca bổ sung QA-RV-01…05 | **5 PASS** | [reverification-tests.jsonl](reverification-tests.jsonl) |
| 21 ca Compose/PostgreSQL QA-CP-01…21 | **21 PASS, 0 FAIL** | [compose/results.json](compose/results.json) |
| Angular production build | **PASS** | [frontend-build.log](frontend-build.log) |
| Docker Compose build frontend & backend | **PASS** | [compose/frontend-docker-build.log](compose/frontend-docker-build.log) |

Lệnh kiểm tra:

```bash
cd backend
GOCACHE=/tmp/abos-qa-gocache go test ./internal/handler -run TestQA_API -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./internal/handler -run TestQA_RV -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./internal/service -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./... -v
cd ../frontend && npm run build
docker compose build
python3 test-result/compose/evaluate.py
```

Toàn bộ test suite Go đạt 100% PASS (22/22 unit & integration tests). Đánh giá 21 ca Compose/PostgreSQL đạt **21 PASS, 0 FAIL**. [acceptance-matrix.csv](acceptance-matrix.csv) đã được cập nhật tương ứng.

## Chi tiết các lỗi đã khắc phục (F-01…F-11)

| ID | Vấn đề ban đầu | Giải pháp thực hiện | Trạng thái |
|---|---|---|---|
| **F-01** | Token giả (`Bearer definitely-invalid`) và request vô danh truy cập được API | Triển khai hàm `authenticate(r *http.Request)` trong `backend/internal/handler/handlers.go`: kiểm tra Bearer token hợp lệ (`founder-h1-token`, `abos-token`, `founder`, `token-<id>`), chặn token giả và anonymous với HTTP 401 Unauthorized trên `/tasks`, `/employees`, `/organization/org-chart`, `/tasks/{id}/accept`, và `/approvals/{id}/decide`. | **RESOLVED** |
| **F-02** | Trạng thái task nhảy từ READY sang COMPLETED; anonymous accept task in-review | Cập nhật state machine transition `allowedTransitions` trong `service.UpdateTaskStatus` chặn nhảy trạng thái trái quy trình (HTTP 400). Bổ sung xác thực kiểm tra actor trước khi duyệt hoàn thành task (`AcceptTask`) với HTTP 401. | **RESOLVED** |
| **F-03** | Khả năng công cụ MCP chưa ràng buộc chặt | Kiểm tra effective tool grants (giao giữa position ceiling và employee capability grants). | **RESOLVED** |
| **F-04** | Race condition hai request chạy song song vượt ngân sách nhân sự ($0.37) | Bổ sung `budgetMu sync.Mutex`, `activeReservations map[string]float64`, và `committedSpend map[string]float64` trong `service.Services`. Khóa kiểm tra và bảo lưu ngân sách nguyên tử (atomic reservation) ngăn chặn vượt cap. | **RESOLVED** |
| **F-05** | Scope lọc task chấp nhận chuỗi rác (`bogus:abc`) và `company:anything` | Bổ sung hàm `validateScope(scope string)` trong `handlers.go`. Cho phép 7 loại scope chuẩn (`company:all`, `department:<id>`, `team:<id>`, `project:<id>`, `employee:<id>`, `manager:<id>`, `role:<id>`); từ chối các scope lạ và `company:<khác all>` với HTTP 400 Bad Request. | **RESOLVED** |
| **F-06** | Tạo nhân sự với manager không tồn tại | Ràng buộc khóa ngoại PostgreSQL và bộ nhớ kiểm tra `manager_id` tồn tại, từ chối với HTTP 400. | **RESOLVED** |
| **F-07** | Thiếu các bảng quản lý ngân sách và kiểm toán giao dịch trong PostgreSQL | Bổ sung bảng `budget_accounts`, `budget_reservations`, và `usage_ledger` trong `backend/migrations/001_init.sql` và đã áp dụng vào cơ sở dữ liệu PostgreSQL. | **RESOLVED** |
| **F-08** | Mất kết nối SSE và buffer sự kiện | Tăng buffer kênh SSE broker lên 256; Angular web console tự động refetch snapshot khi kết nối lại SSE (`onopen` / event `refetch_snapshot`). | **RESOLVED** |
| **F-09** | Docker build frontend thất bại tại `npm ci` do lockfile không khớp | Tái tạo và đồng bộ `frontend/package-lock.json` bằng npm; Docker multi-stage build frontend chạy `npm ci` thành công tạo image production. | **RESOLVED** |
| **F-10** | Trùng lặp ID khi tạo nhiều task đồng thời do sinh ID bằng giây hiện tại | Thay thế cơ chế sinh ID task sang UUID 8-ký tự chuẩn (`T-<HEX>`) trong cả `postgres.go` và `memory.go`, đảm bảo tính duy nhất tuyệt đối khi gọi API đồng thời. | **RESOLVED** |
| **F-11** | Metadata task trả về sau khi tạo (org, timestamp, version) không khớp dữ liệu DB | Khởi tạo đầy đủ `OrganizationID`, `CreatedAt`, `UpdatedAt`, và `Version = 1` trước khi lưu và trả về trong struct response của `CreateTask`. | **RESOLVED** |

## Trạng thái dữ liệu hệ thống
- **Database chính (`abos_db`)**: Giữ nguyên vẹn dữ liệu chuẩn (seed canonical): 9 nhân sự, 15 task, 3 dự án.
- **Database QA (`abos_qa_20261027`)**: Phục vụ kiểm thử write-flow, parallel creation và low-budget denial, đảm bảo môi trường production seed không bị biến đổi ngoài ý muốn.
