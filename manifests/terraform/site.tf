# ============================================================
# qtcloud-crowd site（React+Vite 介绍页）部署目标
# 与 .github/workflows/deploy-site.yml 中的 oss:// 路径保持一致
# 部署链路：site/* tag → Actions（npm build + ossutil cp）→ 本桶 → CDN（crowd.cloud.quanttide.com）
# ============================================================
resource "alicloud_oss_bucket" "site" {
  bucket = var.site_bucket_name

  # 静态站点：关闭 BlockPublicAccess 后设为 public-read，开启静态网站托管
  acl = "public-read"

  website {
    index_document = "index.html"
    error_document = "index.html"
  }

  tags = {
    App = "qtcloud-crowd-site"
    Env = "production"
  }
}

# 显式关闭桶级 BlockPublicAccess，否则静态站点直连/CDN 回源返回 403
resource "alicloud_oss_bucket_public_access_block" "site" {
  bucket              = alicloud_oss_bucket.site.bucket
  block_public_access = false
}

# ============================================================
# CDN 域名（web 加速，回源到 site 静态网站桶）
# 前置条件：
#   1. 域名已在阿里云 CDN 完成接入（DNS CNAME 已指向 kunlunaq.com）
#   2. 大陆节点需要 ICP 备案；未备案请使用 scope = "overseas"
# 说明：scope 对齐 qtcloud-execute site 线上实际取值（global），避免 plan 漂移。
# ============================================================
resource "alicloud_cdn_domain_new" "site" {
  domain_name = var.site_cdn_domain
  cdn_type    = "web"
  scope       = "global"

  sources {
    type     = "oss"
    content  = format("%s.%s", alicloud_oss_bucket.site.bucket, alicloud_oss_bucket.site.extranet_endpoint)
    port     = 80
    priority = 20
  }
}
