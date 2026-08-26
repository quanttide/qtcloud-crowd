# qtcloud-crowd 基础设施（Terraform）

管理「量潮众包管理云」的云上基础设施，对齐 qtcloud-execute 已跑通的部署模式
（FC 3.0 + Terraform + ACR）：

## 一、site（React+Vite 介绍页）

- **site OSS 桶** `qtcloud-crowd-site`：介绍页构建产物部署目标（`deploy-site.yml` 上传路径）
- **site CDN 域名** `crowd.cloud.quanttide.com`：web 加速，回源到上述 OSS 桶
- 桶 ACL `public-read` + 静态网站托管（`index.html` 为根默认页）

## 二、provider（FC 3.0 容器，众包管理 API）

- **FC 函数** `qtcloud-crowd-prod`：custom-container，承载 `src/provider`（任务/执行方/结算 REST API）
- **后台数据 OSS 桶** `qtcloud-crowd-provider`（私有）：完整数据 `data/crowd.json`
  （`QTCLOUD_CROWD_STORE=oss`；见 `fc.tf` 的 `environment_variables`）
- **HTTP 触发器**：`https://<fc-fn>.<region>.fcapp.run`（直连入口，后续可上系统级 API 网关）

> 前台（qtcrowd site/studio）只经 qtcrowd-provider 数据 API 读任务（上架拉取
> `GET /api/tasks?status=published` + 认领/交付写回 API），后台不建公开桶、不投递；
> 黄页快照由 qtcrowd-provider 拉取后写自己的桶（qtcrowd 仓库 IaC 管理）。

## 远程状态（OSS backend）

本配置使用 OSS 远程 state（`provider.tf` 的 `backend "oss"`），本机与 CI 共用。

### 首次：从本地 state 迁移到远程（如存在本地 state）

```bash
cd manifests/terraform
export ALICLOUD_ACCESS_KEY_ID=xxx
export ALICLOUD_ACCESS_KEY_SECRET=xxx

terraform init -migrate-state \
  -backend-config="bucket=quanttide-terraform-state" \
  -backend-config="key=qtcloud-crowd/terraform.tfstate" \
  -backend-config="region=cn-hangzhou"
terraform plan
```

### 之后（本机/CI）

```bash
terraform init \
  -backend-config="bucket=quanttide-terraform-state" \
  -backend-config="key=qtcloud-crowd/terraform.tfstate" \
  -backend-config="region=cn-hangzhou"
export TF_VAR_image=<ACR>/quanttide/qtcloud-crowd-provider:latest
export TF_VAR_oss_access_key_id=xxx
export TF_VAR_oss_access_key_secret=xxx
terraform plan
terraform apply
```

> 注：`oss_access_key_id/secret` 会明文落入 tfstate（FC 环境变量注入），生产建议后续改用 FC 密钥管理。

## 发布（provider）

`provider/*` tag（如 `provider/v0.1.0-alpha.1`）推送触发 `.github/workflows/deploy-provider.yml`：
构建 `src/provider` 镜像 → 推 ACR（`quanttide/qtcloud-crowd-provider`）→ Terraform apply 到 FC。

前置：ACR 仓库已创建（PUBLIC）、GitHub org secrets 已配置（`ALIYUN_ACCESS_KEY_ID/SECRET`、
`ALIYUN_ACR_USERNAME/PASSWORD/REGISTRY`）、OSS 状态桶 `quanttide-terraform-state` 已存在。

## 发布（site）

`site/*` tag 推送触发 `.github/workflows/deploy-site.yml`：构建 `src/site` → 上传
`oss://qtcloud-crowd-site/` → 刷新 CDN（`crowd.cloud.quanttide.com`）。

## 前置条件（公共）

| 项 | 说明 |
|----|------|
| DNS | `crowd.cloud.quanttide.com` 需 CNAME 到 CDN 分配的地址（当前指向 `*.kunlunaq.com`） |
| ICP 备案 | 大陆 CDN 节点要求备案；如未备案，将 `site.tf` 的 `scope` 设为 `overseas` |
| Secrets | GitHub Actions 部署需要仓库配置 `ALIYUN_ACCESS_KEY_ID` / `ALIYUN_ACCESS_KEY_SECRET`（org 级已有） |

## 说明

- site 桶 ACL 为 `public-read`（CDN 回源需要）；provider 后台数据桶为私有（不建公开桶）
- 命名与 `.github/workflows/deploy-site.yml`、`.github/workflows/deploy-provider.yml` 保持一致
