#!/usr/bin/env bash
# 一键部署脚本：检查生产配置 -> docker compose 构建并启动 -> 健康检查
#
# 用法：先自己 git pull（或 checkout 到目标版本），再运行：
#   ./deploy.sh
set -euo pipefail
cd "$(dirname "$0")"

PROD_CONF="blog-api.prod.yaml"
PORT=8812
HEALTH_URL="http://127.0.0.1:${PORT}/v1/admin/profile"

info() { printf '\033[1;34m[deploy]\033[0m %s\n' "$*"; }
err()  { printf '\033[1;31m[deploy]\033[0m %s\n' "$*" >&2; }

# ---- 0. 前置检查 ----
command -v docker >/dev/null 2>&1 || { err "未安装 docker"; exit 1; }
docker compose version >/dev/null 2>&1 || { err "未安装 docker compose v2"; exit 1; }

# ---- 1. 生产配置 ----
# 不存在则从默认配置生成模板：JWT 密钥自动生成随机值，OSS 部分需手动填写
if [[ ! -f "$PROD_CONF" ]]; then
  info "未找到 $PROD_CONF，正在从默认配置生成模板..."
  SECRET=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')
  sed -e "s|AccessSecret: .*|AccessSecret: \"$SECRET\"   # 已自动生成随机密钥|" blog/etc/blog-api.yaml > "$PROD_CONF"
  err "已生成 $PROD_CONF（JWT 密钥已随机生成），请编辑它填写 OSS 真实配置后重新运行本脚本"
  exit 1
fi

# ---- 2. 小内存机器提示：本机构建 Go 镜像需要约 1~2G 内存峰值 ----
if [[ -r /proc/meminfo ]]; then
  mem_kb=$(awk '/MemTotal/{print $2}' /proc/meminfo)
  swap_kb=$(awk '/SwapTotal/{print $2}' /proc/meminfo)
  if (( mem_kb < 3*1024*1024 && swap_kb < 1024*1024 )); then
    err "警告：内存 < 3G 且 swap < 1G，本机构建可能 OOM，建议先配 1~2G swap（见 DEPLOY.md 第 2 节）"
  fi
fi

# ---- 3. 构建并启动 ----
info "构建镜像并启动（首次需拉依赖，耗时几分钟；之后有缓存）..."
docker compose up -d --build

# ---- 4. 健康检查：未带 token 访问受保护接口，应返回统一错误体 code=40103 ----
info "等待服务就绪..."
ok=0
for _ in $(seq 1 30); do
  body=$(curl -s --max-time 2 "$HEALTH_URL" 2>/dev/null || true)
  if [[ "$body" == *'"code":40103'* ]]; then
    ok=1
    break
  fi
  sleep 1
done

if [[ "$ok" == "1" ]]; then
  info "部署完成，服务已就绪（端口 $PORT）"
  info "常用命令：docker compose logs -f blog 查看日志；docker compose restart blog 重启"
else
  err "服务未在 30 秒内就绪，最近日志如下："
  docker compose logs --tail 30 blog || true
  exit 1
fi
