# ============================================================
# provider 运行时 OSS 桶：后台私有数据桶（唯一）
# 按架构（docs/dev-guide「前台自有对象存储」）：后台不建公开桶、不投递——
# 公开层（黄页快照）由前台 qtcrowd-provider 拉取数据 API 后写自己的桶
# （qtcrowd 仓库 IaC 管理），与 qtcloud-execute（单私有桶）一致
# ============================================================

# ── 后台数据桶（私有，默认 ACL 即私有）──
resource "alicloud_oss_bucket" "data" {
  bucket = var.oss_data_bucket

  tags = {
    App         = "qtcloud-crowd-data"
    Environment = var.environment
  }
}
