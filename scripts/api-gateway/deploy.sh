#!/bin/bash
# crowd API 网关部署脚本（幂等）——将请求转发到 FC
#
# 用法：ALIYUN_ACCESS_KEY_ID / ALIYUN_ACCESS_KEY_SECRET（或本地 aliyun 配置）
#   bash scripts/api-gateway/deploy.sh
set -euo pipefail

GROUP_NAME=qtcloud
REGION=cn-hangzhou

# 后端 FC 地址（从 terraform output 获取，或手动填写）
CROWD_FC="https://qtcloudowd-prod-eqpifjcvoh.cn-hangzhou.fcapp.run"

# retry
aliyun_retry() {
  local out
  for i in $(seq 1 8); do
    out=$("$@" 2>&1) && ! echo "$out" | grep -q "i/o timeout\|lookup" && { echo "$out"; return 0; }
    sleep 2
  done
  echo "FAILED: $*" >&2
  return 1
}

# ── 1. 获取分组 ID ──
GROUP_ID=$(aliyun_retry aliyun cloudapi DescribeApiGroups --PageNumber 1 --PageSize 100 |
  python3 -c "
import json,sys
d=json.load(sys.stdin)
for g in d.get('ApiGroupAttributes',{}).get('ApiGroupAttribute',[]):
    if g['GroupName']=='$GROUP_NAME': print(g['GroupId'])
" || true)
if [ -z "$GROUP_ID" ]; then
  echo "ERROR: group $GROUP_NAME not found"
  exit 1
fi
echo "group: $GROUP_ID"

# ── 2. API 定义（名称幂等） ──
# 每行：ApiName|方法|请求路径|后端路径
APIS=(
  # tasks
  "qtcloud-crowd-tasks|GET|/qtcloud-crowd/api/tasks|/api/tasks"
  "qtcloud-crowd-tasks-put|PUT|/qtcloud-crowd/api/tasks|/api/tasks"
  "qtcloud-crowd-tasks-post|POST|/qtcloud-crowd/api/tasks|/api/tasks"
  "qtcloud-crowd-task-detail|GET|/qtcloud-crowd/api/tasks/{id}|/api/tasks/{id}"
  "qtcloud-crowd-task-detail-put|PUT|/qtcloud-crowd/api/tasks/{id}|/api/tasks/{id}"
  "qtcloud-crowd-task-detail-delete|DELETE|/qtcloud-crowd/api/tasks/{id}|/api/tasks/{id}"
  "qtcloud-crowd-task-review|POST|/qtcloud-crowd/api/tasks/{id}/review|/api/tasks/{id}/review"
  "qtcloud-crowd-task-claim|POST|/qtcloud-crowd/api/tasks/{id}/claim|/api/tasks/{id}/claim"
  "qtcloud-crowd-task-deliver|POST|/qtcloud-crowd/api/tasks/{id}/deliver|/api/tasks/{id}/deliver"
  # partners
  "qtcloud-crowd-partners|GET|/qtcloud-crowd/api/partners|/api/partners"
  "qtcloud-crowd-partners-post|POST|/qtcloud-crowd/api/partners|/api/partners"
  "qtcloud-crowd-partner-detail|GET|/qtcloud-crowd/api/partners/{id}|/api/partners/{id}"
  "qtcloud-crowd-partner-certify|POST|/qtcloud-crowd/api/partners/{id}/certify|/api/partners/{id}/certify"
  # settlements
  "qtcloud-crowd-settlements|GET|/qtcloud-crowd/api/settlements|/api/settlements"
  "qtcloud-crowd-settlements-post|POST|/qtcloud-crowd/api/settlements|/api/settlements"
)

for entry in "${APIS[@]}"; do
  IFS='|' read -r name method reqpath svcpath <<< "$entry"
  API_ID=$(aliyun_retry aliyun cloudapi DescribeApis --GroupId "$GROUP_ID" --ApiName "$name" --PageNumber 1 --PageSize 100 |
    python3 -c "
import json,sys
d=json.load(sys.stdin)
for a in d.get('ApiSummarys',{}).get('ApiSummary',[]):
    if a['ApiName']=='$name': print(a['ApiId'])
" || true)
  if [ -n "$API_ID" ]; then
    echo "api exists: $name ($API_ID)"
  else
    REQ="{\"RequestProtocol\":\"HTTPS\",\"RequestHttpMethod\":\"$method\",\"RequestPath\":\"$reqpath\",\"BodyFormat\":\"STREAM\"}"
    SVC="{\"ServiceProtocol\":\"HTTP\",\"ServiceAddress\":\"$CROWD_FC\",\"ServicePath\":\"$svcpath\",\"ServiceHttpMethod\":\"$method\",\"Mock\":\"FALSE\",\"ContentTypeCatagory\":\"CLIENT\",\"ServiceTimeout\":60}"
    API_ID=$(aliyun_retry aliyun cloudapi CreateApi \
      --GroupId "$GROUP_ID" --ApiName "$name" --Description "$name" \
      --RequestConfig "$REQ" --ServiceConfig "$SVC" \
      --Visibility PUBLIC --AuthType ANONYMOUS --ResultType JSON --ResultSample '{}' |
      python3 -c "import json,sys; print(json.load(sys.stdin)['ApiId'])")
    echo "created api: $name ($API_ID)"
  fi
  aliyun_retry aliyun cloudapi DeployApi --GroupId "$GROUP_ID" --ApiId "$API_ID" --StageName RELEASE --Description "deploy.sh" > /dev/null 2>&1 || true
done

echo "DONE"
