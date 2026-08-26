# 众包云开发指南

> 面向：开发者（实现众包管理云）
> 来源：docs/user-guide/（人财事三视角的"最精简的样子"是实施起点）

## 定位

众包管理云 = **管理方后台**（不是交易平台）：审核任务、管理执行方、结算记录。
站点（qtcrowd）只做信息展示——本应用是管理工具。

## 架构：前台唯一服务端（qtcrowd-provider 拉取上架）

```
qtcloud-crowd studio → 后台 provider（私有桶：审核/认证/结算——唯一数据源）
                          │ 审核通过 = published（状态；GET /api/tasks?status=published 可查）
                          ▼
qtcrowd-provider（前台唯一服务端：拉取上架 + 数据 API + 写操作转发）
                          ▲ site/studio 读数据 API（{PROVIDER}/api/tasks——不再直读 OSS/CDN）
                          │ 认领/交付（写回 API——经 qtcrowd-provider 转发后台）
                          │
                      qtcrowd site/studio（公开）
```

### 依赖方向（硬约束）

- **前台可以依赖后台，反之不行**
- 后台（qtcloud-crowd provider）只有**私有数据桶**——不建任何公开桶（公开层不在后台图纸中）
- 前台（qtcrowd）拥有自己的对象存储（qtcrowd-provider 桶）——qtcrowd-provider 上架时写自己的桶
- 后台发布 = **状态置为 published + API 可查**（`GET /api/tasks?status=published` 供前台上架拉取）
- 前台依赖后台两处：上架拉取（published 列表）+ 认领/交付写回 API（均经 qtcrowd-provider）

### 桶设计

| 桶 | 归属 | 权限 | 内容 |
|----|------|------|------|
| 后台数据桶（私有） | qtcloud-crowd | 仅后台 | 完整数据（审核/认证/结算） |
| **qtcrowd-provider 桶** | **qtcrowd（前台自有）** | qtcrowd-provider 读写 | 黄页快照（可接任务）`public/tasks/{id}.json` |

### 数据流

- **上架**：qtcrowd-provider 拉取后台 `GET /api/tasks?status=published` → 写自己桶黄页快照
  （title/description/reward/apply_guide）；被认领/关闭的任务下次同步清理（撤回语义）
- **展示**：site/studio 读 qtcrowd-provider 数据 API（`{PROVIDER}/api/tasks`）——前台自治
- **认领/交付**：写操作 → qtcrowd-provider → 转发后台 API（前台依赖后台）

### 公开桶语义

"当前可接任务"（published 且未被认领）——qtcrowd-provider 自己桶中只放 published 任务的
黄页快照（title/description/reward/报名引导）——内部数据（验收准则等）留后台——
模型不同构由桶边界天然解决。

### 多租户扩展预留（设计预留，代码不做）

架构目标包含后续多租户化——现在不做多租户代码（验证先行），但路径/目录/路由**不写死单租户假设**：

| 层 | 当前（单租户） | 多租户扩展（结构留缝） |
|----|---------------|----------------------|
| 公开层 | `public/tasks/{id}.json` | `public/{tenant}/tasks/{id}.json`——租户市场任务池隔离 |
| 私有层 | `data/crowd.json` | `data/{tenant}/crowd.json`——租户数据隔离 |
| 写回 API | `/api/tasks/{id}/claim` | `/api/{tenant}/tasks/...`——后台服务多租户 |

**约束**：后台不依赖前台（一个后台可服务多个前台市场）；前台是薄壳（读 qtcrowd-provider 数据 API + 写回 API）——新租户市场 = 新前缀 + 新前台实例。

**实施时**：发布接口路径参数化（租户维度可加）、存储层目录支持租户维度、API 路由前缀可加租户——不写多租户逻辑，只留扩展缝。

## 开发结构（人-财-事三视角）

```
src/
├── models/            # 领域模型
│   ├── task.dart          # 任务（title/内容清单/验收准则/状态）
│   ├── partner.dart       # 执行方（姓名/类型/认证状态）
│   └── settlement.dart    # 结算（任务+执行方+金额+时间）
├── repositories/      # 数据访问（DDD 仓储）
│   ├── task_repository.dart
│   ├── partner_repository.dart
│   └── settlement_repository.dart
├── screens/           # 页面
│   ├── task_review_screen.dart   # 事：任务审核（发布/验收两个判断）
│   ├── partner_screen.dart       # 人：执行方（名单+认证状态）
│   └── settlement_screen.dart    # 财：结算（记录一笔）
└── main.dart          # 入口接线
```

## 数据模型（最精简）

```
Task：id / title / content / acceptance_criteria / status（pending→reviewing→done）
Partner：id / name / type（channel|agent|training）/ certified（bool）
Settlement：id / task_id / partner_id / amount / settled_at
```

## 实施顺序（最精简优先）

1. **事：task_review**——任务能不能发（审核）+ 算不算过（验收）——两个判断
2. **人：partner**——名单 + 认证状态（一个动作：认证）
3. **财：settlement**——验收通过记一笔（任务+金额+时间）

每步完成 = 最小闭环可用（真实使用后再加"想象中的样子"功能）。

## 约定

- 技术栈：对齐量潮惯例（Flutter studio + 本地/服务端数据；数据文件 JSON——provider 只用 OSS 不引数据库）
- 命名：models/repositories/screens 三件套（不发明 data/domain/services 层）
- 验收准则兜底：任务说不清验收 = 不能发布（模型层约束）
- 测试：仓储注入 InMemory、失败路径必测

## 参考

- 用法：docs/user-guide/（index + partners + settlement + review）
- 领域边界：quanttide-crowd 根 README（发单/接单/标准交易/信用沉淀）
