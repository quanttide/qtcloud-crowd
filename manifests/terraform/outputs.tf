output "site_bucket_name" {
  description = "site OSS 桶名"
  value       = alicloud_oss_bucket.site.bucket
}

output "site_bucket_domain" {
  description = "site OSS 桶访问域名（CDN 回源地址）"
  value       = format("%s.%s", alicloud_oss_bucket.site.bucket, alicloud_oss_bucket.site.extranet_endpoint)
}

output "site_cdn_domain" {
  description = "site CDN 加速域名"
  value       = alicloud_cdn_domain_new.site.domain_name
}

output "site_cdn_cname" {
  description = "site CDN CNAME，需在 DNS 中配置指向（当前已配置为 kunlunaq.com）"
  value       = alicloud_cdn_domain_new.site.cname
}

# ============================================================
# provider（FC）输出
# ============================================================
output "fc_function_name" {
  description = "函数计算函数名"
  value       = alicloud_fcv3_function.this.function_name
}

output "fc_http_url" {
  description = "FC HTTP 触发器公网地址（系统级 API 网关接入前的直连入口）"
  value       = try(alicloud_fcv3_trigger.http.http_trigger[0].url_internet, "尚未创建")
}

output "oss_data_bucket" {
  description = "provider 运行时后台数据 OSS 桶名（私有）"
  value       = alicloud_oss_bucket.data.bucket
}

output "oss_public_bucket" {
  description = "provider 运行时公开黄页 OSS 桶名（公共读 + 静态托管）"
  value       = alicloud_oss_bucket.public.bucket
}

# ============================================================
# API 网关输出
# ============================================================
output "apigateway_domain" {
  description = "API 网关子域名（需在 DNS 中配置 CNAME）"
  value       = "34c138c4bec1405d942a57d9bb5ede37-cn-hangzhou.alicloudapi.com"
}

output "apigateway_apis" {
  description = "API 网关 API 列表"
  value = {
    tasks        = "/qtcloud-crowd/api/tasks"
    task_detail  = "/qtcloud-crowd/api/tasks/{id}"
    task_claim   = "/qtcloud-crowd/api/tasks/{id}/claim"
    task_deliver = "/qtcloud-crowd/api/tasks/{id}/deliver"
    partners     = "/qtcloud-crowd/api/partners"
    partner_certify = "/qtcloud-crowd/api/partners/{id}/certify"
    settlements  = "/qtcloud-crowd/api/settlements"
  }
}
