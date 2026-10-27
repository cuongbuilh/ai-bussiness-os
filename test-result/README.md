# Kết quả kiểm thử — 2026-10-27

## Phạm vi và cách chạy lại

Repo không có `./sources`. Mã triển khai nằm tại `backend/` và `frontend/`; tài liệu yêu cầu nằm ở các thư mục `00-governance` đến `08-delivery`. Bộ ca nghiệm thu gốc: `08-delivery/ACCEPTANCE-TESTS.md`. Ma trận [acceptance-matrix.csv](acceptance-matrix.csv) giữ ID và trạng thái của toàn bộ AT-01…AT-36, cùng 5 ca Goal-driven hậu MVP.

Chạy lại các ca HTTP logic (mỗi ca tạo repository memory mới):

```bash
cd backend
GOCACHE=/tmp/abos-qa-gocache go test ./internal/handler -run TestQA_API -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./internal/service -count=1 -v
cd ../frontend
npm run build
```

`api-tests.jsonl` là log máy đọc được của 10 ca QA-API-01…10. `existing-go-tests.log` và `frontend-build.log` giữ bằng chứng test và build. Các test QA có kỳ vọng theo requirement, nên hiện **8 FAIL, 2 PASS**. Test service có sẵn **7 PASS**; Angular build **PASS**. `go test ./...` sẽ FAIL vì các ca hồi quy QA đang bắt lỗi thực tế.

## Lỗi cần sửa theo thứ tự ưu tiên

| ID | Mức | Trạng thái | Bằng chứng và tác động |
|---|---|---|---|
| F-01 | Critical | OPEN | QA-API-01/07: API không có auth middleware; `GET /employees` trả 200 vô danh, `POST /approvals/APR-12/decide` chấp nhận `resolvedBy=h1` do caller tự gửi. `backend/cmd/server/main.go:69`, `backend/internal/handler/handlers.go:274`. |
| F-02 | Critical | OPEN | QA-API-03/04/09: `PATCH /tasks/{id}` chấp nhận chuyển READY → ACCEPTED và trạng thái tùy ý; endpoint accept nhận cả BACKLOG. Không có state machine, reviewer authority, criteria hoặc audit. `backend/internal/service/service.go:177`, `:224`. |
| F-03 | Critical | OPEN | QA-API-05: q1 không có grant `task-mcp`/`docs-mcp` nhưng execution luôn ghi các bước đó là thành công. Đây là mô phỏng, không có gateway MCP/LLM thật. `backend/internal/service/service.go:249`, `:326`. |
| F-04 | Critical | OPEN | QA-API-06: agent cap $0.10 vẫn chạy một lần $0.37, spent thành $0.37. Check chỉ so `spent >= cap` trước khi chạy; không reserve atomically, không ledger. `backend/internal/service/service.go:261`, `:350`. |
| F-05 | High | OPEN | PostgreSQL `GetTasks` nhận `scope` nhưng không sử dụng, nên các board có thể trả toàn bộ task và summary sai. `backend/internal/repository/postgres.go:379`. Chưa chạy DB thật do Docker không khả dụng. |
| F-06 | High | OPEN | QA-API-10: memory repo cho phép manager ID không tồn tại; service chỉ chặn tự quản lý. Không có API đổi manager hay lịch sử reporting line. `backend/internal/service/service.go:40`, `backend/internal/repository/memory.go:262`. |
| F-07 | High | OPEN | Runtime, budget, model, approval, KPI và portfolio dựa nhiều vào seed/simulation; không có agent worker/checkpoint, tool gateway, provider callback, ledger, scoped memory, trigger, HR/model proposal, audit. Các AT tương ứng được đánh dấu `MISSING` trong ma trận. |
| F-08 | Medium | OPEN | SSE client kết nối lại sau lỗi nhưng không fetch snapshot khi kết nối lại; broker bỏ event nếu subscriber chậm, nên có thể mất trạng thái. `frontend/src/app/services/api.service.ts:55`, `backend/internal/service/broker.go:58`. |

## Giới hạn bằng chứng

- `httptest` đi qua handler/service/memory repository, không mở cổng thật. Sandbox chặn `listen tcp` với `socket: operation not permitted`.
- Docker CLI trong WSL báo không được tích hợp, nên chưa chạy PostgreSQL migrations, Compose hay E2E browser. Kết luận về Postgres là từ kiểm tra code, ghi rõ `code-review` trong ma trận.
- Frontend không có file `*.spec.ts`; build thành công chỉ chứng minh biên dịch, không chứng minh UI hoạt động. Prototype QA cũ kiểm tra HTML demo, không phải Angular tích hợp.
- `README-DEPLOY.md` mô tả hệ thống là triển khai đầy đủ; bằng chứng trên không hỗ trợ nhận định đó. Theo `README.md`, tài liệu gốc vẫn xem MVP là mục tiêu cần triển khai.

## Ý nghĩa trạng thái

`PASS`: kỳ vọng được kiểm thử và đạt trong phạm vi nêu. `FAIL`: hành vi được chạy và trái yêu cầu. `PARTIAL`: một phần có bằng chứng, phần còn lại thiếu hoặc chưa chạy. `MISSING`: không thấy đường triển khai cần thiết trong code hiện tại. `DEFERRED`: ngoài MVP theo baseline. Sau khi fix, chạy lại QA tương ứng rồi cập nhật `acceptance-matrix.csv` và trạng thái `F-*`.
