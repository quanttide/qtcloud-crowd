# ============================================================
# provider 运行时 OSS 桶：双桶分离（后台私有 + 公开黄页）
# 与 qtcloud-execute（单私有桶）的区别：crowd provider 按架构
# （docs/dev-guide「OSS 共享数据层」）同时写两个桶——
#   私有桶：完整数据（审核/认证/结算），仅后台可读写
#   公开桶：黄页快照（public/tasks/{id}.json，title/reward/报名引导），公共读 + 静态托管
# ============================================================

# ── 后台数据桶（私有，默认 ACL 即私有）──
resource "alicloud_oss_bucket" "data" {
  bucket = var.oss_data_bucket

  tags = {
    App         = "qtcloud-crowd-data"
    Environment = var.environment
  }
}

# ── 公开黄页桶（公共读 + 静态网站托管；与私有数据桶分开建）──
# 关键经验（参考 qtcloud-data / qtcloud-execute 踩坑记录）：
# 1. 新桶默认开启【桶级 BlockPublicAccess】→ 需先关闭才能设置公共读
# 2. 关闭后设置 acl = public-read（阿里云禁止的是 API 设置，关闭 BPA 后允许）
# 3. 静态网站托管：index.html 为根默认页（黄页无真实首页时占位/错误页复用）
resource "alicloud_oss_bucket" "public" {
  bucket = var.oss_public_bucket

  acl = "public-read"

  website {
    index_document = "index.html"
    error_document = "index.html"
  }

  tags = {
    App         = "qtcloud-crowd-public"
    Environment = var.environment
  }
}

# 同 site 桶：显式关闭桶级 BlockPublicAccess，否则公共读 / CDN 回源返回 403
resource "alicloud_oss_bucket_public_access_block" "public" {
  bucket              = alicloud_oss_bucket.public.bucket
  block_public_access = false
}
