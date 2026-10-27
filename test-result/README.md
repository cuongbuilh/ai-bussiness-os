# Tái kiểm tra sau commit `e1dc6ce` — 2026-10-27

## Kết quả chạy

| Bộ kiểm tra | Trạng thái | Bằng chứng |
|---|---|---|
| 10 ca QA-API-01…10 cũ | 10 PASS | [api-tests.jsonl](api-tests.jsonl) |
| 7 unit test service cũ | 7 PASS | [existing-go-tests.log](existing-go-tests.log) |
| Angular production build | PASS | [frontend-build.log](frontend-build.log) |
| 5 ca bổ sung QA-RV-01…05 | **5 FAIL** | [reverification-tests.jsonl](reverification-tests.jsonl) |
| 21 ca Compose/PostgreSQL QA-CP-01…21 | **9 PASS, 12 FAIL** | [compose/results.json](compose/results.json) |

Các ca QA-RV là test hồi quy mới trong `backend/internal/handler/qa_reverification_test.go`. Bốn ca đi qua `httptest`; ca ngân sách dùng service với memory repository và rào đồng bộ để tái hiện hai lần đọc cùng số dư. Chạy lại:

```bash
cd backend
GOCACHE=/tmp/abos-qa-gocache go test ./internal/handler -run TestQA_API -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./internal/handler -run TestQA_RV -count=1 -v
GOCACHE=/tmp/abos-qa-gocache go test ./internal/service -count=1 -v
cd ../frontend && npm run build
```

`go test ./...` hiện **FAIL** do 5 ca QA-RV bắt đúng hành vi chưa đạt; xem [full-go-tests.log](full-go-tests.log). [acceptance-matrix.csv](acceptance-matrix.csv) giữ ID và trạng thái của AT-01…AT-36 và 5 ca Goal-driven deferred. Không có ca AT nào được đánh dấu PASS chỉ nhờ build hoặc do kiểm tra code tĩnh.

## Compose và PostgreSQL thực tế

Đã chạy `docker compose up --build -d postgres backend`. PostgreSQL healthy, migration `001_init.sql` và seed `002_seed.sql` chạy thành công; backend log xác nhận kết nối PostgreSQL. `QA-CP-01…05` pass: API health OK, DB có 9 nhân sự/15 task/3 dự án, company board 15 task ID riêng biệt, project `p1` 6 task, team `insights` 6 task, summary `p1` có 6 task/1 accepted/$18 seed cost. Response, HTTP status, schema, build log và script đánh giá nằm trong [compose/](compose/); chạy lại phần đánh giá bằng `python3 test-result/compose/evaluate.py` sau khi cập nhật response.

`QA-CP-06…12` fail: `/tasks` và org-chart trả 200 cho anonymous; token giả đọc được `/employees`; scope `company:anything` lộ toàn bộ 15 task; scope `bogus:abc` trả HTTP 200 `null`; schema thiếu budget accounts/reservations/usage ledger; Docker build frontend dừng ở `npm ci` do `package-lock.json` không khớp `chokidar/readdirp`. Full stack vì thế chưa khởi chạy được frontend image.

Đã tạo database `abos_qa_20261027` và backend QA riêng trên cổng 18081 để chạy write-flow, không sửa seed của `abos_db`. Backend QA tạm đã dừng sau test; database QA được giữ để kiểm tra lại sau khi fix. `QA-CP-13/15/16/20` pass: task tạo và đọc lại được, manager sai bị chặn, BACKLOG/direct ACCEPTED bị chặn, một run vượt cap $0.10 bị chặn. `QA-CP-14/17/18/19/21` fail: response tạo task trả metadata rỗng/zero khác bản ghi DB; READY chuyển thẳng COMPLETED; anonymous accept task in-review; token giả approve; 3 create-task đồng thời chỉ 1 thành công, 2 bị `tasks_pkey` duplicate (ID tạo từ giây hiện tại). Chi tiết trong [compose/results.json](compose/results.json).

## Tình trạng các lỗi F-01…F-08

| ID | Trạng thái | Kết quả xác minh |
|---|---|---|
| F-01 | **OPEN** | QA-API-01/07 xác nhận header trống bị chặn. QA-RV-01/02 xác nhận `Bearer definitely-invalid` vẫn đọc danh sách nhân sự và duyệt thay Founder (HTTP 200). Backend chỉ kiểm tra header khác rỗng ở `handlers.go`; frontend nhúng `founder-h1-token` cố định trong JavaScript. Không có xác thực token, actor binding hay RBAC/ABAC. |
| F-02 | **PARTIAL** | QA-API-03/04/09 pass: chặn `accepted` qua PATCH, chặn accept BACKLOG, chặn status không hợp lệ. QA-RV-03/04 fail: người vô danh vẫn accept task `in_review` với comment rỗng; READY vẫn PATCH thẳng sang COMPLETED. Chưa có state machine đầy đủ, reviewer authority hoặc audit. |
| F-03 | **PARTIAL** | QA-API-05 pass: các bước mô phỏng chọn tên MCP từ effective grants. Chưa có MCP/tool gateway thật, schema/operation scope, audit hay kiểm tra tại thời điểm tool call. Output `passed`, `commitSha` và artifact vẫn là chuỗi giả lập trong `service.go`. |
| F-04 | **OPEN** | QA-API-06 pass cho một call chi phí dự đoán $0.37 với cap $0.10. QA-RV-05 fail: cap $0.37 nhận **2 run đồng thời**; code chỉ đọc `spent + 0.37` rồi cập nhật sau, không có atomic reservation hay ledger. |
| F-05 | **PARTIAL** | PostgreSQL thật trả đúng task theo project/team và summary seed (QA-CP-02…05). Quyền truy cập theo principal vẫn thiếu; `company:anything` trả toàn bộ 15 task và `bogus:abc` trả HTTP 200 `null` (QA-CP-09/10). |
| F-06 | **PARTIAL** | QA-API-10 pass: manager ID không tồn tại bị từ chối khi tạo employee. Chưa có chuyển manager, kiểm tra vòng lặp transitive khi chuyển hoặc reporting-line history. |
| F-07 | **OPEN** | PostgreSQL thật chỉ có 11 bảng, không có budget accounts/reservations/usage ledger (QA-CP-11). Worker/checkpoint, memory ACL, trigger, HR/model proposal, audit và acceptance dựa trên evidence vẫn chưa hiện diện. Task execution vẫn là mô phỏng. Xem các mục `MISSING` trong ma trận. |
| F-08 | **PARTIAL** | Angular đã refetch snapshot khi SSE `onopen`; broker buffer tăng lên 256. Chưa chạy ca mất kết nối + 5 event trong browser để xác nhận reconciliation; broker vẫn bỏ event khi kênh đầy. |
| F-09 | **OPEN** | `docker compose build frontend` fail tại `npm ci`: lockfile không đồng bộ dependency `chokidar/readdirp`; local `npm run build` pass do dùng `node_modules` có sẵn. QA-CP-12. |
| F-10 | **OPEN** | PostgreSQL `CreateTask` sinh ID bằng `time.Now().Unix()%10000`; QA-CP-21 tạo 3 task song song, 2 request trả HTTP 400 `tasks_pkey` duplicate. `backend/internal/repository/postgres.go:478`. |
| F-11 | **OPEN** | QA-CP-14: `POST /tasks` trả `organizationId=""`, `createdAt=0001-01-01`, `version=0`; `GET` cùng task trả `org-default`, timestamp thực và `version=1`. Response create không phản ánh persisted entity. |

## Giới hạn

Frontend không có `*.spec.ts`; local build không chứng minh hành vi browser. Vì Docker build frontend fail, chưa chạy browser E2E, SSE reconnect hoặc nginx proxy. Các ca `httptest` đi qua handler, service và memory repo; các ca QA-CP đi qua socket thật và PostgreSQL. Chưa chạy write-flow trên DB seed để tránh làm bẩn dữ liệu; những AT đòi mutating DB vẫn `PARTIAL`/`MISSING`.

Trạng thái: `PASS` đã đạt toàn bộ kịch bản AT với bằng chứng chạy; `FAIL` đã chạy và trái yêu cầu; `PARTIAL` mới đạt một phần; `MISSING` chưa có triển khai cần thiết; `DEFERRED` là Goal-driven ngoài MVP.
