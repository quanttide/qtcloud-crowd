# ============================================================
# provider（qtcloud-crowd 众包管理 API）—— 阿里云 FC 3.0 容器部署
# 数据：运行时 OSS store（QTCLOUD_CROWD_STORE=oss）：
#   - 后台私有桶（QTCLOUD_OSS_BUCKET）：完整数据 data/crowd.json（审核/认证/结算）
#   - 前台桶（QTCLOUD_OSS_PUBLIC_BUCKET=qtcrowd-site）：黄页快照 public/tasks/{id}.json
#     —— 前台已有桶（qtcrowd 仓库 IaC 管理），后台只做投递者角色，不建公开桶
# 凭证：provider 通过环境变量读取静态 AK/SK 访问 OSS（见 internal/store/oss.go 与
#       cmd/server/main.go——注意 crowd 代码读取的是 QTCLOUD_OSS_* 前缀，
#       与 qtcloud-execute 的 ALIYUN_OSS_* 不同）；故 FC 函数环境变量注入
#       QTCLOUD_OSS_ACCESS_KEY_ID/SECRET。此 AK/SK 会明文落入 tfstate，
#       生产环境建议后续改用 FC 密钥管理/配置中心注入。
# ============================================================

# FC 默认角色：允许 FC 服务挂载弹性网卡访问 VPC（应用级，保留与 sibling 模板一致）
resource "alicloud_ram_role" "fc" {
  role_name                   = "${local.app_name_prefix}-fc"
  assume_role_policy_document = <<EOF
{
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": ["fc.aliyuncs.com"]
      }
    }
  ],
  "Version": "1"
}
EOF
  description                 = "Function Compute 默认角色（qtcloud-crowd）"
}

resource "alicloud_ram_role_policy_attachment" "fc_vpc" {
  policy_name = "AliyunECSNetworkInterfaceManagementAccess"
  policy_type = "System"
  role_name   = alicloud_ram_role.fc.role_name
}

# 函数计算（FC 3.0）：custom-container 容器镜像，运行时 OSS store（后台私有桶 + 投递前台桶）
resource "alicloud_fcv3_function" "this" {
  function_name   = local.app_name_prefix
  description     = "qtcloud-crowd 众包管理 API（后台私有桶 + 投递前台 qtcrowd-site 桶，运行时 OSS 数据源）"
  runtime         = "custom-container"
  handler         = "index.handler" # custom-container 必填占位，实际由容器监听端口决定
  cpu             = 0.5
  memory_size     = var.fc_memory
  disk_size       = 512 # FC 3.0 必填（MB）
  timeout         = var.fc_timeout
  internet_access = true
  role            = alicloud_ram_role.fc.arn

  custom_container_config {
    image = var.image
    port  = 8080
  }

  # 对齐 provider 运行时约定（见 src/provider/cmd/server/main.go 与 internal/store/oss.go）：
  # QTCLOUD_CROWD_STORE=oss 走 OSS；QTCLOUD_CROWD_DATA 为后台桶内对象键；
  # QTCLOUD_OSS_PUBLIC_BUCKET 指向前台 qtcrowd-site 桶（投递黄页快照 public/tasks/，不建公开桶）
  environment_variables = {
    QTCLOUD_CROWD_STORE           = "oss"
    QTCLOUD_CROWD_DATA            = "data/crowd.json"
    QTCLOUD_CROWD_ADDR            = ":8080"
    QTCLOUD_OSS_BUCKET            = alicloud_oss_bucket.data.bucket
    QTCLOUD_OSS_PUBLIC_BUCKET     = "qtcrowd-site"
    QTCLOUD_OSS_ENDPOINT          = var.oss_endpoint
    QTCLOUD_OSS_ACCESS_KEY_ID     = var.oss_access_key_id
    QTCLOUD_OSS_ACCESS_KEY_SECRET = var.oss_access_key_secret
  }

  tags = {
    project     = var.project
    environment = var.environment
  }
}

# HTTP 触发器：直接访问（后续经系统级 API 网关统一接入，此触发器保留为直连通道）
resource "alicloud_fcv3_trigger" "http" {
  function_name = alicloud_fcv3_function.this.function_name
  trigger_name  = "http"
  trigger_type  = "http"
  qualifier     = "LATEST"
  trigger_config = jsonencode({
    authType = "anonymous"
    methods  = ["GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"]
  })
}
