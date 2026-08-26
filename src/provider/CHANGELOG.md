# Changelog

## [Unreleased]

### 移除

- 删除发布投递层 `internal/publish/`（审核通过不再写前台公开桶）——新架构下前台只从
  qtcrowd-provider 数据 API 读（`GET /api/tasks?status=published` 上架拉取），后台不建/不投递公开桶；
  published 状态机保留
- store 接口移除 `PutPublic/DeletePublic`（local/OSS 实现及测试同步清理）
- 环境变量 `QTCLOUD_OSS_PUBLIC_BUCKET` 移除（main.go / terraform fc.tf / README 同步）

### 新增

- `GET /api/tasks` 支持 `?status=` 过滤（如 `?status=published`）——前台 qtcrowd-provider 上架拉取契约：
  只返回指定状态任务；不带 status 返回全部；非法 status 返回 400

## [0.1.0-alpha.2] - 2026-08-26

### 修复

- 使用阿里云官方 OSS SDK 替代手动签名，修复 403 SignatureDoesNotMatch 错误
- 修复 OSS URL 缺少 https:// 前缀的问题

### 新增

- OSS 共享数据层：Task 状态机含 published（pending→reviewing→published→accepted→done），审核通过发布黄页快照到公开桶（`public/tasks/{id}.json`）
- 发布层 `internal/publish/`：发布/撤回公开对象（local 模式 `data/public/`，OSS 模式公开桶 `QTCLOUD_OSS_PUBLIC_BUCKET`）
- 写回 API：`POST /api/tasks/{id}/claim`（published→accepted）+ `/deliver`（accepted→reviewing）
- store 扩展：`PutPublic/DeletePublic`

## [0.1.0-alpha.1] - 2026-08-26

### 新增

- 初始版本：任务/执行方/结算三资源 REST API + 本地/OSS 存储