# Baseline quyết định đã chốt — V1.1

**Ngày:** 2026-10-27 · **Trạng thái:** CONFIRMED by product owner · **Phạm vi:** single internal organization. Các giải pháp công nghệ ở các file khác là *Proposed* trừ khi được ghi rõ là Confirmed.

## Ma trận quyết định
| ID | Quyết định đã chốt | Hệ quả kiến trúc / nghiệm thu |
|---|---|---|
| C-01 | Một công ty nội bộ, chưa làm SaaS | Chính sách org thống nhất; tenant-ready identifiers nhưng không làm billing đa tenant ở MVP |
| C-02 | Mọi ngành nghề, position/title/role/task tùy biến | Không hard-code backend dev/research hay mô hình công ty phần mềm |
| C-03 | Agent tự hành động trong quyền; việc quan trọng phải duyệt | Enforcement server-side, không giao quyền quyết định cho prompt |
| C-04 | Một direct manager tại một thời điểm, có cộng tác chéo tổ chức | Reporting tree tách collaboration/project graph |
| C-05 | Kiểm soát đồng thời tiền, token, RPM/TPM, thời gian, số task | Hierarchical/overlapping budget constraints và ledger chống double count |
| C-06 | Platform tập trung provider/API keys | Secret references, centrally-managed model/provider registry |
| C-07 | Benchmark và khuyến nghị tự động; đổi primary model phải duyệt | Versioned model assignments + approval + rollback |
| C-08 | Docker Compose trước, Kubernetes sau | Stateless workers, durable checkpoints, isolation |
| C-09 | Position có MCP/tool/skills; mỗi agent có MCP/tool/skills; model capability gate | Effective permissions là giao của các policy; skill không tự cấp quyền |
| C-10 | Web Console trực quan org-chart, profile nhân viên, manager, task/KPI, giao việc | Web là control surface chính, không chỉ monitoring |
| C-11 | Unified company/department/team/project/person/manager/position task/project boards | Cùng canonical task IDs, scope authorization, no cost double count |
| C-12 | AI Manager được **đề xuất** tuyển/bổ nhiệm/thay đổi cơ cấu; thay đổi quan trọng do human duyệt | OrganizationChangeProposal/approval workflow; AI không tự tạo quyền hoặc cấp ngân sách |
| C-13 | Cho phép công việc do human, manager, schedule hoặc event khởi tạo; chủ động trong policy | Trigger registry, dedup, approval/risk/budget gates; MVP task-driven |
| C-14 | Nhiều AI cùng làm task, có 1 owner chịu trách nhiệm và nhiều collaborators | Participants, subtasks, reviewer, delegation chain; cost attribution rõ ràng |
| C-15 | Position xác lập MCP/tool/skill mặc định và *permission ceiling*; agent có cấu hình riêng | Position ceiling + agent grant + task/runtime policy; elevation chỉ qua exception duyệt |
| C-16 | Memory dài hạn theo Employee, Team, Department, Project, Organization | Namespace + ACL + provenance + retention; không mặc định truy cập chéo |
| C-17 | Budget cảnh báo 80%; hard limit dừng; manager chỉ có thể đề xuất tăng | 80%-threshold alert và hard-stop enforcement; request top-up không auto-apply |
| C-18 | AI được tự đánh giá/đề xuất cải thiện prompt/skill/model; thay đổi quan trọng phải duyệt | KPI evidence + change proposals + offline eval + versioning |
| C-19 | **Cả Task-driven và Goal-driven**; MVP tập trung Task-driven, Goal-driven phát triển sau | Plan theo hai pha; kiến trúc có goal provenance nhưng không bắt xây goal planner ở MVP |

## Quy tắc diễn giải
1. **Confirmed** là ràng buộc sản phẩm. **Proposed** là giải pháp kỹ thuật hoặc chính sách triển khai, có thể điều chỉnh miễn vẫn đáp ứng Confirmed. **Deferred** là tương lai, không được thể hiện trên UI như tính năng đã hoàn thiện.
2. *Task-driven* không có nghĩa chỉ human mới tạo được task: agent/manager, scheduler và event handler vẫn tạo được task trong quyền, nhưng phải có work-order/trigger hợp lệ và ngân sách/capability phù hợp.
3. *Goal-driven* (Deferred) cho phép owner giao business goal, AI xây initiative/plan, liên tục đo KPI và re-plan dưới quyền được giao. **Không đồng nghĩa** tự nâng quyền, tự mua dịch vụ hay tự tuyển nhân viên.
4. Các hạn mức/thời hạn mẫu trong tài liệu là ví dụ thiết kế, không phải budget được owner phê duyệt.

## Vùng chưa cần chốt để xây tài liệu
- Provider/model nào dùng mặc định, giá budget cụ thể và giá token thay đổi theo thị trường.
- Identity provider, thời hạn retention chính xác, catalog MCP đầu tiên.
- Danh mục hành động R3/R4 bổ sung cho từng ngành, thời điểm kích hoạt Goal-driven.
- Default automatic acceptance: đề xuất **human acceptance** cho mọi output chưa có deterministic verifier được duyệt.
