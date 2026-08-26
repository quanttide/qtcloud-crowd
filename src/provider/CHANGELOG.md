# Changelog

> 本文件为 `provider` scope（Go 服务端——FC 3.0 部署）的版本记录。发布使用
> `qtcloud-devops release --version provider/vX.Y.Z --changelog src/provider/CHANGELOG.md`，
> 推送 `provider/*` tag 会触发部署（Docker 镜像 → ACR → Terraform apply → FC）。

## [Unreleased]

### 新增

- OSS 共享数据层：Task 状态机含 published（pending→reviewing→published→accepted→done），审核通过发布黄页快照到公开桶（`public/tasks/{id}.json`）
- 发布层 `internal/publish/`：发布/撤回公开对象（local 模式 `data/public/`，OSS 模式公开桶 `QTCLOUD_OSS_PUBLIC_BUCKET`）
- 写回 API：`POST /api/tasks/{id}/claim`（published→accepted）+ `/deliver`（accepted→reviewing）
- store 扩展：`PutPublic/DeletePublic`

## [0.1.0-alpha.1] - 2026-08-26

### 新增

- 初始版本：任务/执行方/结算三资源 REST API + 本地/OSS 存储
