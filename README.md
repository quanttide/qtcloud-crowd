# 量潮众包云（`qtcloud-crowd`）

量潮众包管理云——自营发销售众包，面向外部渠道与代理，按量潮标准结算。

## 入口

| 服务 | 地址 |
|------|------|
| API | `https://api.quanttide.com/qtcloud-crowd/` |
| site | `https://crowd.cloud.quanttide.com/` |

## 组成

本项目包含四个独立组件，各自独立发布：

- **provider**（Go）— 服务端，提供任务/执行方/结算 REST API，支持本地存储和 OSS 共享数据层
- **site**（React + Vite）— 产品介绍页，展示核心功能
- **studio**（Flutter）— 管理前端，支持任务审核、执行方管理、结算记录
- **cli**（Rust）— 管理方命令行工具，功能与 studio 对齐

## 目录结构

```
qtcloud-crowd/
├── src/
│   ├── provider/     # Go 服务端
│   ├── site/         # React 首页
│   ├── studio/       # Flutter 工作台
│   └── cli/          # Rust 命令行
├── docs/             # 项目文档
├── manifests/        # 部署清单
├── CHANGELOG.md      # 主仓库变更记录
├── CONTRIBUTING.md   # 贡献指南
└── AGENTS.md         # Agent 工作指引
```

## 快速开始

### provider

```bash
cd src/provider
go run ./cmd/server
```

### site

```bash
cd src/site
npm install
npm run dev
```

### studio

```bash
cd src/studio
flutter pub get
flutter run -d chrome
```

### cli

```bash
cd src/cli
cargo build --release
./target/release/qtcloud-crowd --help
```

## 贡献

详见 [CONTRIBUTING.md](CONTRIBUTING.md)。
