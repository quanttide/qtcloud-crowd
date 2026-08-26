# provider / backend 配置（对齐 qtcloud-execute 模板；qtcrowd 的 site/公开层由 qtcrowd 仓库自管，
# 本仓库只管 qtcloud-crowd 应用自身：provider（FC）+ site 介绍页（qtcloud-crowd-site）——
# 后台不建公开桶，公开层由前台 qtcrowd 自有桶承载）

terraform {
  # 远程状态：OSS（本机与 CI 共用，CI 必须持久化状态）。初始化时通过 -backend-config 指定：
  #   terraform init \
  #     -backend-config="bucket=quanttide-terraform-state" \
  #     -backend-config="key=qtcloud-crowd/terraform.tfstate" \
  #     -backend-config="region=cn-hangzhou"
  # 首次从本地 state 迁移：使用 -migrate-state（见 README）
  backend "oss" {}
}

provider "alicloud" {
  region = var.region
}
