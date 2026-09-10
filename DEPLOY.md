# 部署文档

个人博客后端（go-zero + SQLite + RustFS）的部署说明。整体策略：**只在 CI（GitHub Actions）里编译镜像，服务器只负责拉镜像运行**，绝不在 2C2G 的小机器上编译。

---

## 1. 架构一览

```
开发者 push 代码 ──► GitHub Actions ──► 构建多阶段镜像 ──► 推到镜像仓库(GHCR/ACR)
                                                        │
                                             服务器 docker pull
                                                        │
                                              docker compose up -d
                                                        │
                                            ┌────────────┴─────────────┐
                                         blog 容器                RustFS(对象存储)
                                    /app/data  (SQLite)        图片上传目标
                                    /app/logs
```

- 应用二进制：`blog`（go-zero HTTP 服务，端口 8812）
- 数据库：SQLite 文件，落在容器 `/app/data/blog.db`，由 volume 持久化
- 图片：上传到 RustFS（S3 兼容），不在容器里存文件

---

## 2. 服务器前置准备

- 安装 Docker 与 Docker Compose v2
- （若用 GHCR 且仓库为私有）在服务器执行 `docker login ghcr.io`，用有 `read:packages` 权限的 Personal Access Token
- 开放防火墙 8812（或放在反代/Nginx 后）

---

## 3. 镜像构建（CI 自动）

推送到 `main` 或打 `v*` tag 时，`.github/workflows/build.yml` 自动：

1. 多阶段 `docker build`（构建阶段联网拉 Go 依赖，运行阶段只留二进制 + 配置 + CA 证书）
2. 推送到 GHCR，标签：`main` / `sha-xxxx` / `v1.2.3`

> 想换镜像仓库（如阿里云 ACR，国内拉取更快）：改 `build.yml` 里的 `IMAGE` 常量和 `login-action` 的 registry/账号即可。

如需在本地手动构建（需要能访问 Go 模块代理的网络）：

```bash
docker build -t myblog-backend:local .
```

---

## 4. 服务器部署步骤

### 4.1 准备生产配置 `blog-api.prod.yaml`

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
  AccessSecret: "换成一段随机长字符串(用于 JWT 签名)"
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

### 4.2 拉起服务

```bash
# 确认 docker-compose.yml 里的 image 地址改成你自己的
docker compose pull
docker compose up -d
```

容器启动命令是 `./migrate && exec ./blog -f etc/blog-api.yaml`：
**每次启动会先幂等执行 `migrations/*.sql` 自动建表**，所以空 volume 也能直接跑，无需手动建表。

### 4.3 验证

```bash
# 健康检查：未带 token 访问受保护接口应返回统一错误体（不再是空 body）
curl -s http://127.0.0.1:8812/v1/admin/upload
# 期望：{"code":40103,"message":"未登录或登录已失效"}

# 查看日志
docker compose logs -f blog
```

---

## 5. 日常更新流程

1. 本地改代码 → `git push`
2. GitHub Actions 自动出新版镜像（同一 `:main` 标签会更新）
3. 服务器上：
   ```bash
   docker compose pull
   docker compose up -d
   ```

---

## 6. 数据与安全

- **备份**：定期备份 volume 对应的 `/app/data`（SQLite 单文件，停服或拷贝 WAL 一致的快照即可）
- **密钥**：`blog-api.prod.yaml` 含 OSS Key 与 JWT 密钥，只在服务器存在，不进 git
- **HTTPS**：给 RustFS 的 `PublicBaseURL` 配好可公网访问的地址；业务建议放在 Nginx/Caddy 反代后统一加 TLS

---

## 7. 常见问题

- **首次启动报“表不存在”**：不会。容器启动已自动跑迁移；若你手动在容器外换了 DB，确保该 DB 也执行过 `migrations/*.sql`。
- **上传图片报 TLS / x509 错误**：镜像已内置 CA 证书；若出现，通常是 `OSS.Endpoint` 配错或网络不通。
- **`401` 空响应**：已修复，现在返回 `{"code":40103,"message":"未登录或登录已失效"}`。
