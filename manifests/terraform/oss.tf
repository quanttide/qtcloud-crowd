# ============================================================
# provider 运行时 OSS 桶：后台私有数据桶（唯一）
# 按架构（docs/dev-guide「前台自有对象存储」）：后台不建公开桶——
# 公开层（黄页快照）由前台 qtcrowd 自有桶（qtcrowd-site）承载，
# 后台 provider 只做投递者角色（OSS 写入前台桶，见 fc.tf 的
# QTCLOUD_OSS_PUBLIC_BUCKET 指向 qtcrowd-site）。
# 与 qtcloud-execute（单私有桶）的区别：crowd provider 写两个桶——
#   私有桶：完整数据（审核/认证/结算），仅后台可读写
#   前台桶：qtcrowd-site（qtcrowd 仓库 IaC 管理，已有桶不新建），黄页快照
#           （public/tasks/{id}.json，title/reward/报名引导），公共读 + 静态托管
# ============================================================

# ── 后台数据桶（私有，默认 ACL 即私有）──
resource "alicloud_oss_bucket" "data" {
  bucket = var.oss_data_bucket

  tags = {
    App         = "qtcloud-crowd-data"
    Environment = var.environment
  }
}
