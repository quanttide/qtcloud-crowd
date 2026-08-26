# ROADMAP

## provider Unreleased

- OSS 共享数据层：Task 状态机含 published（pending→reviewing→published→accepted→done），审核通过发布黄页快照到公开桶（`public/tasks/{id}.json`）
- 发布层 `internal/publish/`：发布/撤回公开对象（local 模式 `data/public/`，OSS 模式公开桶 `QTCLOUD_OSS_PUBLIC_BUCKET`）
- 写回 API：`POST /api/tasks/{id}/claim`（published→accepted）+ `/deliver`（accepted→reviewing）
- store 扩展：PutPublic/DeletePublic

## 待办

- [ ] provider 接入 API 网关
- [ ] studio 打 tag 并发布
- [ ] cli 打 tag 并发布
- [ ] provider 发布 Unreleased 内容