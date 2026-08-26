# 众包云开发指南

> 面向：开发者（实现众包管理云）
> 来源：docs/user-guide/（人财事三视角的"最精简的样子"是实施起点）

## 定位

众包管理云 = **管理方后台**（不是交易平台）：审核任务、管理执行方、结算记录。
站点（qtcrowd）只做信息展示——本应用是管理工具。

## 架构：OSS 共享数据层（两端一数据层）

```
qtcloud-crowd studio → 后台 provider（私有桶：审核/认证/结算）
                          │ 审核通过 = 发布（OSS CopyObject → 公开桶）
                          ▼
                      公开桶（公共读 + CDN：任务池）
                          ▲ 认领/交付（写回 API——前台依赖后台）
                          │
qtcrowd provider（轻量：只做写操作转发）→ qtcrowd site/studio（公开）
```

### 依赖方向（硬约束）

- **前台可以依赖后台，反之不行**
- 后台（qtcloud-crowd provider）只依赖 OSS——不知道前台存在
- 前台（qtcrowd）依赖：OSS（读公开桶）+ 后台 API（认领/交付写回）

### 桶设计

| 桶 | 权限 | 内容 | 消费者 |
|----|------|------|--------|
| 后台桶（私有） | 仅后台 | 完整数据（审核/认证/结算） | 后台 provider |
| 公开桶（公共读+CDN） | 公共读 | 黄页快照（可接任务：title/reward/报名引导） | qtcrowd site/studio 直接读 |

### 数据流

- **发布**：后台审核通过 → 写公开桶对象（CopyObject）——"推"的动作在后台（带审计）
- **展示**：site/studio 直接读公开桶（CDN 静态分发）——无需读 API
- **认领/交付**：写操作 → qtcrowd provider → 调后台 API（前台依赖后台）
- **撤回**：后台删除公开桶对象（任务关闭/被认领）

### 公开桶语义

"当前可接任务"（published 且未被认领）——认领后后台更新/移除对象。
公开桶内容是黄页模型视图（title/reward/报名引导）——内部数据（验收准则等）留后台——模型不同构由桶边界天然解决。

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
