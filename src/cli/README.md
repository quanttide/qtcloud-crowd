# qtcloud-crowd CLI

量潮众包管理云 CLI —— **管理方工具**（不是交易平台）：审核任务、管理执行方、记录结算。

## 构建

```bash
cd src/cli
cargo build --release
```

## 数据文件

- 默认路径：`data/crowd.json`（相对当前工作目录），可用环境变量 `QTCLOUD_CROWD_DATA` 或全局参数 `--data <path>` 覆盖；
- 文件不存在时自动初始化：`{ "tasks": [], "partners": [], "settlements": [] }`；
- 格式：2 空格缩进 JSON，空值字段省略，原子写入（临时文件 + rename）。

## 用法

```bash
# 事：任务审核
qtcloud-crowd tasks list                     # 列出任务（--status pending/reviewing/done 过滤）
qtcloud-crowd tasks review <id> --action approve   # 通过验收 → done（验收准则为空不能通过）
qtcloud-crowd tasks review <id> --action reject    # 打回 → pending

# 人：执行方认证
qtcloud-crowd partners list                   # 列出执行方（名单 + 认证状态）
qtcloud-crowd partners certify <id>           # 认证 → 已认证

# 财：结算记录
qtcloud-crowd settlements list                # 列出结算记录
qtcloud-crowd settlements add <task_id> <partner_id> <amount>   # 记一笔（任务必须已通过验收）
```

## 数据模型

与 `src/studio/lib/models/` 对齐：

| 实体 | 字段 |
|------|------|
| Task | id / title / content / acceptance_criteria / status（pending→reviewing→done） |
| Partner | id / name / type（channel\|agent\|training）/ certified |
| Settlement | id / task_id / partner_id / amount / settled_at（ISO 8601） |

## 模型层校验

- **任务审核**：验收准则为空 → 不能通过（说不清验收 = 不发单）；已完成的不能重复审核/打回
- **结算**：任务必须已通过验收（status=done）；同一任务只结算一次；执行方必须存在；金额必须为正数
