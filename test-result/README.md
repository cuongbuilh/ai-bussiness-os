# Kết quả khắc phục và kiểm định toàn diện — 2026-10-27

## Kết quả kiểm định sau khắc phục

| Bộ kiểm tra | Trạng thái | Bằng chứng |
|---|---|---|
| 10 ca QA-API-01…10 | **10 PASS** | [api-tests.jsonl](api-tests.jsonl) |
| 7 unit test service | **7 PASS** | [full-go-tests.log](full-go-tests.log) |
| 10 ca hồi quy QA-RV-01…10 | **10 PASS** | [reverification-tests.jsonl](reverification-tests.jsonl) |
| 21 ca Compose/PostgreSQL QA-CP-01…21 | **21 PASS, 0 FAIL** | [compose/results.json](compose/results.json) |
| Angular production build | **PASS** | [frontend-build.log](frontend-build.log) |
| Docker Compose build frontend & backend | **PASS** | [compose/frontend-docker-build.log](compose/frontend-docker-build.log) |

Lệnh kiểm tra:

```bash
cd backend
GOCACHE=/tmp/abos-qa-gocache go test ./internal/handler -run TestQA_API -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./internal/handler -run TestQA_RV -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./internal/service -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./... -count=1 -v
cd ../frontend && npm run build
docker compose build
python3 test-result/compose/evaluate.py
```

Toàn bộ test suite Go đạt 100% PASS (27/27 unit & integration tests). Đánh giá 21 ca Compose/PostgreSQL đạt **21 PASS, 0 FAIL**.

## Khắc phục chuyên sâu các lỗi xác thực, phân quyền & lưu trữ ngân sách

1. **Chặn mạo danh Founder/Nhân sự bằng Public ID (QA-RV-06)**:
   - Hàm `authenticate` trong [handlers.go](file:///home/cuongbv/sources/cuongbuilh/ai-business-os-spec-v1.1/backend/internal/handler/handlers.go#L40) chỉ chấp nhận token bảo mật hợp lệ (`founder-h1-token`, `abos-token`).
   - Mọi nỗ lực truyền ID công khai (`Bearer h1`, `Bearer token-h1`, `Bearer m1`, v.v.) hoặc token giả đều bị từ chối với **HTTP 401 Unauthorized**.

2. **Bảo vệ toàn diện đọc dữ liệu Task và Approval trước Anonymous (QA-RV-07)**:
   - Các endpoint `GET /tasks?scope=company:all`, `GET /tasks/{taskId}` (vd: `T-101`), và `GET /approvals` bắt buộc phải có token xác thực. Yêu cầu vô danh đều bị từ chối với **HTTP 401 Unauthorized**.

3. **Kiểm tra quyền các thao tác Ghi & Ràng buộc Actor Identity (QA-RV-08, QA-RV-09)**:
   - Các thao tác ghi nhạy cảm (`POST /tasks`, `POST /employees`, `POST /tasks/{id}/execute`, `PATCH /tasks/{id}`, `POST /tasks/{id}/assign`) yêu cầu xác thực Bearer token hợp lệ.
   - Trong `POST /approvals/{id}/decide`: actor giải quyết duyệt (`resolvedBy`) bắt buộc phải trùng khớp với danh tính được gắn với Bearer token (ví dụ: Founder `h1`). Nếu client gửi `resolvedBy` mạo danh người khác (vd: `m1`), hệ thống trả về **HTTP 403 Forbidden**.

4. **Bền vững giới hạn Ngân sách qua nhiều Instance & Sổ cái Usage Ledger (QA-RV-10)**:
   - Tích hợp bảng `budget_reservations` và `usage_ledger` trực tiếp vào [postgres.go](file:///home/cuongbv/sources/cuongbuilh/ai-business-os-spec-v1.1/backend/internal/repository/postgres.go) và [service.go](file:///home/cuongbv/sources/cuongbuilh/ai-business-os-spec-v1.1/backend/internal/service/service.go).
   - Mỗi lần chạy task (`ExecuteTaskRun`): tạo bản ghi reservation trạng thái `ACTIVE` vào DB với thời hạn (TTL 10 phút), sau khi chạy xong cập nhật trạng thái thành `COMMITTED` (hoặc `RELEASED` nếu lỗi).
   - Mọi bước thực thi suy luận/công cụ được ghi nhận thành các bản ghi bất biến (immutable) trong bảng `usage_ledger`.
   - Các thao tác cập nhật chi phí và trạng thái vào database đều được kiểm tra lỗi nghiêm ngặt (`if err != nil`), bảo toàn tính nhất quán giữa các instance backend và sau khi khởi động lại.

## Trạng thái dữ liệu hệ thống
- **Database chính (`abos_db`)**: Giữ nguyên vẹn dữ liệu chuẩn (seed canonical): 9 nhân sự, 15 task, 3 dự án.
- **Database QA (`abos_qa_20261027`)**: Chạy kiểm thử write-flow, parallel creation, xác thực budget reservation và ghi usage ledger thực tế trên PostgreSQL.
