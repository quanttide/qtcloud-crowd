# 贡献指南

## 提交流程

1. 在子模块仓库中创建分支
2. 进行修改，提交前运行格式检查
3. 提交并推送
4. 回到本仓库更新子模块引用

## 代码规范

| 组件 | 格式化工具 |
|------|-----------|
| provider（Go） | `gofmt` + `go vet` |
| cli（Rust） | `cargo fmt` + `cargo clippy` |
| site（TypeScript） | `oxlint` |
| studio（Dart） | `dart format` |

## CHANGELOG 规范

每个组件维护独立的 CHANGELOG，根目录 `CHANGELOG.md` 仅记录仓库层面变更。

### 何时记录

提交前检查变更是否涉及：

- 用户可见的功能变更（新 API、新页面、行为改变）
- Bug 修复
- 破坏性变更

### 记录到哪里

**这个变更是哪个组件的，就写到哪个组件的 CHANGELOG 里。**

| 组件 | CHANGELOG 路径 |
|------|----------------|
| provider | `src/provider/CHANGELOG.md` |
| site | `src/site/CHANGELOG.md` |
| studio | `src/studio/CHANGELOG.md` |
| cli | `src/cli/CHANGELOG.md` |

格式遵循 [Keep a Changelog](https://keepachangelog.com/)。