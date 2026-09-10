# 部署文档

个人博客后端（go-zero + SQLite + RustFS）的部署说明。整体策略：**不走镜像仓库——服务器上 git pull 拉代码，本地 `docker compose build` 构建运行**。

---

## 1. 架构一览

```
开发者 push 代码
        │
服务器 git pull
        │
./deploy.sh  (docker compose up -d --build + 健康检查)
        │
   ┌────┴─────────┐
 blog 容器      RustFS 容器（对象存储）
 /app/data     图片上传目标
 /app/logs     由 ./deploy-rustfs.sh 独立部署（一次性）
```

- 应用二进制：`blog`（go-zero HTTP 服务，端口 8812）
- 数据库：SQLite 文件，落在容器 `/app/data/blog.db`，由 volume 持久化
- 图片：上传到 RustFS（S3 兼容，9000 端口），不在容器里存文件
- 建表/迁移：SQL 已 `go:embed` 进二进制，**服务启动时自动幂等执行**，无需单独跑迁移

---

## 2. 服务器前置准备

- 安装 Docker 与 Docker Compose v2、git
- 2C2G 机器上构建 Go 镜像需要约 1~2GB 内存峰值，**务必先配 1~2GB swap**（zram 亦可），否则构建可能被 OOM kill
- 开放防火墙 8812（或放在反代/Nginx 后）；RustFS 控制台 9001 建议仅内网/SSH 隧道访问

---

## 3. 首次部署

### 3.1 拉代码

```bash
git clone <仓库地址> myblog_backend
cd myblog_backend
```

### 3.2 部署 RustFS（图片存储）

```bash
./deploy-rustfs.sh
```

脚本会：随机生成密钥存入 `rustfs.env`（已 gitignore）→ 启动 RustFS 容器（9000 S3 API / 9001 控制台）→ 自动建 `myblog` bucket 并设公开读 → **最后打印一段 OSS 配置，第 3.3 步要用**。

想先跳过图片功能也行：prod 配置里 OSS 用占位值即可，服务照常启动，仅后台上传会报错，之后再补。

### 3.3 准备生产配置 `blog-api.prod.yaml`

先运行一次 `./deploy.sh`——它发现没有 `blog-api.prod.yaml` 时会**自动生成模板**（JWT 密钥已随机填好）并退出。然后编辑该文件，把 3.2 输出的 OSS 配置块粘贴替换占位段：

```bash
vim blog-api.prod.yaml
```

```yaml
# 只有 OSS 段需要填，其余保持模板原样
OSS:
  Endpoint: "http://172.17.0.1:9000"    # docker0 网桥地址（blog 容器内 127.0.0.1 指容器自己）
  Region: "us-east-1"                    # 自建 RustFS 保持不动
  Bucket: "myblog"
  AccessKeyID: "<rustfs.env 里的值>"
  SecretAccessKey: "<rustfs.env 里的值>"
  UsePathStyle: true                     # 自建服务保持 true
  PublicBaseURL: "http://<服务器IP或域名>:9000/myblog"  # 浏览器访问图片的地址；上域名后改 https://域名/files/myblog
```

### 3.4 构建并启动

```bash
./deploy.sh   # 构建镜像 -> 启动 -> 健康检查（GET /v1/admin/profile 应返回 code=40103）
```

或手动执行：

```bash
docker compose up -d --build
```

首次构建要联网拉 Go 依赖，耗时几分钟；之后有层缓存，增量构建通常 1 分钟内完成。
服务启动时会自动执行未应用的数据库迁移（幂等），空 volume 也能直接拉起。

### 3.5 验证

```bash
# deploy.sh 已内置健康检查；手动验证：
curl -s http://127.0.0.1:8812/v1/admin/profile
# 期望：{"code":40103,"message":"未登录或登录已失效"}
curl -s http://127.0.0.1:8812/v1/articles
# 期望：{"code":0,"message":"ok","data":{"list":[],"total":0}}

# 查看日志
docker compose logs -f blog
```

### 3.6 初始化管理员账号

```bash
# 本地或服务器上生成密码的 bcrypt 哈希
go run ./tools/adminhashpwd '你的密码'

# 然后对 volume 里的 blog.db 执行 SQL（sqlite3 CLI 或任意 SQLite 工具）。
# volume 宿主机路径可用 docker volume inspect myblog_backend_blog-data 查看：
#   INSERT INTO admin(username, password_hash, nickname) VALUES('admin', '<上面的哈希>', '博主')
#   ON CONFLICT(username) DO UPDATE SET password_hash=excluded.password_hash;
```

---

## 4. 日常更新流程

```bash
cd myblog_backend
git pull        # 拉代码这步你自己做
./deploy.sh     # 构建 + 启动 + 健康检查
```

回滚：`git checkout <旧 commit>` 后直接 `./deploy.sh`。

> 注意：数据库迁移是单向的（只有 up 没有 down），回滚代码前确认旧代码兼容新表结构；一期表结构变更少，一般无碍。

---

## 5. CI 说明

`.github/workflows/build.yml` 目前仍会构建并推送镜像到 GHCR。不走镜像仓库后它只剩"验证镜像能构建"的作用：

- 想保留构建验证：把 `push: true` 改成 `push: false`（省掉 GHCR 登录和推送）
- 完全不需要：直接删除该文件

---

## 6. 数据与安全

- **备份**：SQLite 是 WAL 模式，**不要直接 cp 正在写入的库文件**。用在线备份命令：

  ```bash
  sqlite3 /var/lib/docker/volumes/myblog_backend_blog-data/_data/blog.db \
      ".backup '/backup/blog-$(date +%F).db'"
  # 或在容器/应用内执行 VACUUM INTO '/backup/blog.db'
  ```

  建议每日 cron 备份 + 异地存放，保留最近 30 天（见 design.md 5.4）。
- **密钥**：`blog-api.prod.yaml`（OSS Key、JWT 密钥）与 `rustfs.env`（RustFS 管理员密钥）只存在于服务器，均在 `.gitignore` 中，不进 git
- **HTTPS**：给 RustFS 的 `PublicBaseURL` 配好可公网访问的地址；业务建议放在 Nginx/Caddy 反代后统一加 TLS

---

## 7. 常见问题

- **构建被 kill / 卡死**：2C2G 内存不够，先加 swap 再重试。
- **构建卡在拉 Go 依赖 / 报 proxy.golang.org 超时**：Dockerfile 默认用国内代理 `goproxy.cn`；海外服务器构建时覆盖：`docker compose build --build-arg GOPROXY=https://proxy.golang.org,direct`。
- **deploy.sh 提示"已生成 blog-api.prod.yaml"后退出**：这是正常流程——模板已生成（JWT 密钥已随机填好），填入 OSS 真实配置后重跑即可，见 3.3。
- **deploy.sh 健康检查未就绪但日志显示已启动**：健康探针是 `GET /v1/admin/profile`，期望返回 `code:40103`；先手动 `curl` 该地址确认，再看 `docker compose logs blog`。
- **改了 rustfs.env 里的密钥**：RustFS 容器创建后密钥已固化在容器里，需 `docker rm -f rustfs` 后重跑 `./deploy-rustfs.sh` 才会生效（数据不丢，存在 ~/rustfs/data）。
- **首次启动报"表不存在"**：不会。服务启动已自动跑迁移；若你手动换了 DB 文件，确保该 DB 执行过迁移（可用 `go run ./tools/migrate -dsn <path>` 手工补）。
- **上传图片报 TLS / x509 错误**：镜像已内置 CA 证书；若出现，通常是 `OSS.Endpoint` 配错或网络不通。
- **`401` 空响应**：已修复，现在返回 `{"code":40103,"message":"未登录或登录已失效"}`。
