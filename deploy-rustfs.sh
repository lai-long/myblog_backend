#!/usr/bin/env bash
# 一键部署 RustFS 对象存储（博客图片上传用）。
# 做的事：生成/复用密钥 -> 准备数据目录 -> 启动容器 -> 建 bucket 并设公开读 -> 打印要填进 blog-api.prod.yaml 的配置
#
# 用法：./deploy-rustfs.sh
# 重复执行是安全的（幂等）。密钥首次生成后保存在 rustfs.env（已 gitignore），改了它需先 docker rm -f rustfs 再重跑。
set -euo pipefail
cd "$(dirname "$0")"

CONTAINER="rustfs"
DATA_DIR="${RUSTFS_DATA_DIR:-$HOME/rustfs/data}"
ENV_FILE="rustfs.env"
BUCKET="myblog"
API_PORT=9000
CONSOLE_PORT=9001

info() { printf '\033[1;34m[rustfs]\033[0m %s\n' "$*"; }
err()  { printf '\033[1;31m[rustfs]\033[0m %s\n' "$*" >&2; }

# ---- 0. 前置检查 ----
command -v docker >/dev/null 2>&1 || { err "未安装 docker"; exit 1; }

# ---- 1. 密钥：首次随机生成并落盘，后续复用 ----
if [[ -f "$ENV_FILE" ]]; then
  info "复用已有密钥 $ENV_FILE"
  # shellcheck disable=SC1090
  source "$ENV_FILE"
else
  RUSTFS_ACCESS_KEY=$(openssl rand -hex 8 2>/dev/null || head -c 8 /dev/urandom | od -An -tx1 | tr -d ' \n')
  RUSTFS_SECRET_KEY=$(openssl rand -hex 24 2>/dev/null || head -c 24 /dev/urandom | od -An -tx1 | tr -d ' \n')
  printf 'RUSTFS_ACCESS_KEY=%s\nRUSTFS_SECRET_KEY=%s\n' "$RUSTFS_ACCESS_KEY" "$RUSTFS_SECRET_KEY" > "$ENV_FILE"
  chmod 600 "$ENV_FILE"
  info "已生成随机密钥并保存到 $ENV_FILE（权限 600，勿提交 git）"
fi

# ---- 2. 数据目录（RustFS 容器以 UID 10001 运行，属主不对会权限拒绝） ----
mkdir -p "$DATA_DIR"
if [[ "$(stat -c %u "$DATA_DIR")" != "10001" ]]; then
  info "修正数据目录属主为 10001（可能需要输入 sudo 密码）..."
  sudo chown -R 10001:10001 "$DATA_DIR" || {
    err "chown 失败，请手动执行：sudo chown -R 10001:10001 $DATA_DIR"
    exit 1
  }
fi

# ---- 3. 启动 / 确保容器运行 ----
if [[ "$(docker inspect -f '{{.State.Running}}' "$CONTAINER" 2>/dev/null || true)" == "true" ]]; then
  info "容器 $CONTAINER 已在运行，跳过启动"
elif docker inspect "$CONTAINER" >/dev/null 2>&1; then
  info "容器已存在但未运行，启动..."
  docker start "$CONTAINER"
else
  info "创建并启动 RustFS 容器..."
  docker run -d \
    --name "$CONTAINER" \
    --restart unless-stopped \
    -p "${API_PORT}:9000" \
    -p "${CONSOLE_PORT}:9001" \
    -e "RUSTFS_ACCESS_KEY=$RUSTFS_ACCESS_KEY" \
    -e "RUSTFS_SECRET_KEY=$RUSTFS_SECRET_KEY" \
    -v "$DATA_DIR:/data" \
    rustfs/rustfs:latest
fi

# ---- 4. 等待 S3 API 就绪（根路径返回任何 HTTP 响应都算就绪，S3 对未签名请求返回 403 属正常） ----
info "等待 RustFS 就绪..."
ready=0
for _ in $(seq 1 30); do
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 2 "http://127.0.0.1:${API_PORT}/" 2>/dev/null || true)
  if [[ "$code" != "000" && -n "$code" ]]; then
    ready=1
    break
  fi
  sleep 1
done
[[ "$ready" == "1" ]] || { err "RustFS 未在 30 秒内就绪，日志："; docker logs --tail 30 "$CONTAINER" || true; exit 1; }

# ---- 5. 建 bucket + 公开读（用一次性 mc 容器操作，无需在宿主机装 mc） ----
info "创建 bucket [$BUCKET] 并设置公开读..."
MC_ENV="MC_HOST_local=http://$RUSTFS_ACCESS_KEY:$RUSTFS_SECRET_KEY@127.0.0.1:${API_PORT}"
docker run --rm --network host -e "$MC_ENV" minio/mc mb --ignore-existing "local/$BUCKET"
docker run --rm --network host -e "$MC_ENV" minio/mc anonymous set download "local/$BUCKET"

# ---- 6. 输出下一步 ----
cat <<EOF

$(info "RustFS 部署完成")
  控制台:  http://<服务器IP>:${CONSOLE_PORT}  （账号密码见 $ENV_FILE）
  S3 API:  http://<服务器IP>:${API_PORT}

把下面这段填进 blog-api.prod.yaml 的 OSS 部分（替换同名键），然后运行 ./deploy.sh：

OSS:
  Endpoint: "http://172.17.0.1:${API_PORT}"
  Region: "us-east-1"
  Bucket: "$BUCKET"
  AccessKeyID: "$RUSTFS_ACCESS_KEY"
  SecretAccessKey: "$RUSTFS_SECRET_KEY"
  UsePathStyle: true
  PublicBaseURL: "http://<服务器IP或域名>:${API_PORT}/$BUCKET"

说明：
  Endpoint      用 docker0 网桥地址（172.17.0.1），因为 blog 容器内 127.0.0.1 指容器自己
  PublicBaseURL 是浏览器访问图片的地址；之后上域名走 Nginx 反代时改成 https://域名/files/$BUCKET
EOF
