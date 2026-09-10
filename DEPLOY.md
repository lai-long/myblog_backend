# 部署文档

个人博客后端（go-zero + SQLite + RustFS）的部署说明。整体策略：**不走镜像仓库——服务器上 git pull 拉代码，本地 `docker compose build` 构建运行**。

---

## 1. 架构一览

```
开发者 push 代码
        │
服务器 git pull
        │
docker compose up -d --build        # 本地构建镜像并重启
        │
   ┌────┴─────────┐
 blog 容器      RustFS（对象存储）
 /app/data     图片上传目标
 /app/logs     （独立部署，见 RustFS 文档）
```

- 应用二进制：`blog`（go-zero HTTP 服务，端口 8812）
- 数据库：SQLite 文件，落在容器 `/app/data/blog.db`，由 volume 持久化
- 图片：上传到 RustFS（S3 兼容），不在容器里存文件
- 建表/迁移：SQL 已 `go:embed` 进二进制，**服务启动时自动幂等执行**，无需单独跑迁移

---

## 2. 服务器前置准备

- 安装 Docker 与 Docker Compose v2、git
- 2C2G 机器上构建 Go 镜像需要约 1~2GB 内存峰值，**务必先配 1~2GB swap**（zram 亦可），否则构建可能被 OOM kill
- 开放防火墙 8812（或放在反代/Nginx 后）
- RustFS 独立部署（不在本 compose 里），拿到 endpoint 和 access key 备用

---

## 3. 首次部署

### 3.1 拉代码

```bash
git clone <仓库地址> myblog_backend
cd myblog_backend
```

### 3.2 准备生产配置 `blog-api.prod.yaml`

在 `docker-compose.yml` 同级目录创建该文件（已在 `.gitignore` 中，含密钥，禁止提交）：

```yaml
Name: blog-api
Host: 0.0.0.0
Port: 8812

ZeroLog:
  Path: ../logs/blog.log
  MaxSize: 100
  MaxBackups: 14
  MaxAge: 14
  Level: info

Auth:
  AccessSecret: "换成一段随机长字符串(用于 JWT 签名)"   # 生成：openssl rand -hex 32
  AccessExpire: 7200

SqliteDSN: "../data/blog.db?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"

# RustFS / S3 兼容对象存储（图片上传用），以下填你的真实值
OSS:
  Endpoint: "https://s3.your-rustfs.com"
  Region: "us-east-1"
  Bucket: "myblog"
  AccessKeyID: "YOUR_ACCESS_KEY"
  SecretAccessKey: "YOUR_SECRET_KEY"
  UsePathStyle: true          # 自建 S3 兼容服务保持 true
  PublicBaseURL: "https://s3.your-rustfs.com/myblog"  # 返回给前端的 URL 前缀
```

### 3.3 构建并启动

```bash
./deploy.sh   # 一键脚本：拉代码 -> 检查配置 -> 构建启动 -> 健康检查
```

或手动执行：

```bash
docker compose up -d --build
```

首次构建要联网拉 Go 依赖，耗时几分钟；之后有层缓存，增量构建通常 1 分钟内完成。
服务启动时会自动执行未应用的数据库迁移（幂等），空 volume 也能直接拉起。

### 3.4 验证

```bash
# 健康检查：未带 token 访问受保护接口应返回统一错误体（不再是空 body）
curl -s http://127.0.0.1:8812/v1/admin/upload
# 期望：{"code":40103,"message":"未登录或登录已失效"}

# 查看日志
docker compose logs -f blog
```

### 3.5 初始化管理员账号

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
./deploy.sh    # 等价于 git pull + docker compose up -d --build + 健康检查
```

回滚：`git checkout <旧 commit>` 后 `./deploy.sh --skip-pull`（跳过 pull，直接按当前代码重建）。

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
- **密钥**：`blog-api.prod.yaml` 含 OSS Key 与 JWT 密钥，只在服务器存在，不进 git
- **HTTPS**：给 RustFS 的 `PublicBaseURL` 配好可公网访问的地址；业务建议放在 Nginx/Caddy 反代后统一加 TLS

---

## 7. 常见问题

- **构建被 kill / 卡死**：2C2G 内存不够，先加 swap 再重试。
- **构建卡在拉 Go 依赖 / 报 proxy.golang.org 超时**：Dockerfile 默认用国内代理 `goproxy.cn`；海外服务器构建时覆盖：`docker compose build --build-arg GOPROXY=https://proxy.golang.org,direct`。
- **首次启动报"表不存在"**：不会。服务启动已自动跑迁移；若你手动换了 DB 文件，确保该 DB 执行过迁移（可用 `go run ./tools/migrate -dsn <path>` 手工补）。
- **上传图片报 TLS / x509 错误**：镜像已内置 CA 证书；若出现，通常是 `OSS.Endpoint` 配错或网络不通。
- **`401` 空响应**：已修复，现在返回 `{"code":40103,"message":"未登录或登录已失效"}`。
