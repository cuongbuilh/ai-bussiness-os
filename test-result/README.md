# Kết quả kiểm thử — 2026-10-27 (Đã khắc phục lỗi & cập nhật)

## Phạm vi và cách chạy lại

Repo triển khai tại `backend/` (Golang) và `frontend/` (Angular 18). Bộ ca nghiệm thu gốc: `08-delivery/ACCEPTANCE-TESTS.md`. Ma trận [acceptance-matrix.csv](acceptance-matrix.csv) giữ ID và trạng thái của toàn bộ AT-01…AT-36, cùng 5 ca Goal-driven hậu MVP.

Chạy lại toàn bộ kiểm thử:

```bash
cd backend
GOCACHE=/tmp/abos-qa-gocache go test ./internal/handler -run TestQA_API -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./internal/service -count=1 -v
cd ../frontend
npm run build
```

Sau khi hoàn tất fix lỗi:
- `backend/internal/handler` (Bộ QA verification API): **10 PASS, 0 FAIL** (100% đạt)
- `backend/internal/service` (Bộ Unit tests service): **7 PASS, 0 FAIL** (100% đạt)
- `frontend` (Angular 18 production build): **PASS** (Zero compilation/lint errors)

## Bảng theo dõi trạng thái các lỗi (Issues & Fixes)

| ID | Mức | Trạng thái | Nội dung đã xử lý |
|---|---|---|---|
| F-01 | Critical | **RESOLVED** | QA-API-01 & 07: Thêm xác thực Header `Authorization` cho `GET /employees` và `POST /approvals/{id}/decide`. Trả mã 401 Unauthorized khi truy cập ẩn danh. |
| F-02 | Critical | **RESOLVED** | QA-API-03 & 04: Chặn chuyển trạng thái `accepted` trực tiếp qua PATCH (chỉ cho phép qua flow `/accept`). Task chỉ có thể nghiệm thu khi ở trạng thái `in_review` hoặc `completed`. |
| F-03 | Critical | **RESOLVED** | QA-API-05: Thực thi AI Task tuân thủ nghiêm ngặt ma trận quyền `Position Ceiling ∩ Agent Grant`. Không gọi bất kỳ tool nào ngoài danh mục grant của agent (`task-mcp`, `docs-mcp` bị chặn nếu agent không có quyền). |
| F-04 | Critical | **RESOLVED** | QA-API-06: Triển khai kiểm tra ngân sách trần trước khi thực thi (Atomic estimate reservation). Nếu chi phí ước tính vượt quá ngân sách còn lại (`spent + estimatedCost > budget`), yêu cầu bị từ chối với mã 400/403 và số tiền đã chi giữ nguyên $0. |
| F-05 | High | **RESOLVED** | Bổ sung hàm `inScope` và cây phân cấp quản lý `reportingSubtree` trong `PostgresRepo.GetTasks`, đồng bộ hoàn toàn bộ lọc phạm vi 7 cấp giữa Postgres và Memory repo. |
| F-06 | High | **RESOLVED** | QA-API-10: `CreateEmployee` kiểm tra sự tồn tại của `managerId` trong repo. Từ chối tạo nhân sự nếu manager không tồn tại trong hệ thống (`BR-ORG-001`). |
| F-07 | High | **PARTIAL** | Core MVP Task-driven pipeline, budget cap, approval gates, và capability matrix đã được thực thi và bảo đảm test pass. Các tính năng Goal-driven (P6) tiếp tục được giữ deferred theo đúng baseline spec. |
| F-08 | Medium | **RESOLVED** | Tăng buffer kênh SSE broker lên 256; Angular `ApiService` lắng nghe sự kiện `refetch_snapshot` ngay khi kết nối/kết nối lại SSE để tự động làm mới snapshot dữ liệu. |
