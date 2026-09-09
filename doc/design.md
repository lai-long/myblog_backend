# 个人博客系统设计文档 · 后端

| 项目 | 内容 |
|---|---|
| 文档版本 | v1.4（对象存储由 MinIO 更换为 RustFS） |
| 编写日期 | 2026-09-07 |
| 技术栈 | Go (go-zero) + SQLite + Redis + RustFS / Docker |
| 关联文档 | 《博客系统设计文档-前端》（接口契约为两份文档的共同约定） |

---

## 1. 概述

### 1.1 范围

本文档覆盖博客系统的服务端：go-zero API 单服务（模块化单体）、数据存储（SQLite / Redis / RustFS）、认证体系、部署与运维。

### 1.2 用户角色（权限模型的输入）

**分期策略：一期只做"管理员 + 游客"两个角色，账号表保持最小结构（登录即管理员，无角色概念）；注册用户体系二期开放，接口位置与演进方向本文档预留说明，具体表结构届时按需 ALTER / 新建。**

| 角色 | 一期 | 二期 |
|---|---|---|
| 游客（Guest） | 浏览文章、搜索、订阅 RSS；**评论区只读**（能看到管理员回复） | 同左 |
| 注册用户（User） | ——（不存在） | GitHub OAuth / 邮箱注册登录，发表评论、管理自己的评论 |
| 管理员（Admin） | 用户名密码登录；文章增删改、评论管理（可回复）、站点配置 | + 用户管理（封禁/解封） |

> 一期的评论：不开放发表入口，仅管理员可发表/回复（作为"博主回复"展示）。这样评论区 UI 与数据表一期就能建好，二期打开注册后无需改结构。注册通道规划为 GitHub OAuth（主）+ 邮箱注册（兜底）；QQ 互联列入三期（需域名备案）。

---

## 2. 架构设计

### 2.1 服务端架构图

```
                前端 / 外部请求
                      │ /api/*
                      ▼
        ┌─────────────────────────────┐
        │   go-zero API 单服务 :8080   │
        │  group 分组路由：            │
        │  user / post / comment /     │
        │  site / notify               │
        └──────┬──────┬──────┬────────┘
               │      │      │ 进程内调用
               ▼      ▼      ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐     ┌──────────┐
        │ SQLite   │ │ Redis    │ │ RustFS   │     │ SMTP     │
        │ (WAL)    │ │(缓存/限流│ │(图片对象  │     │(QQ 邮箱, │
        │          │ │ /黑名单) │ │  存储)   │     │ 异步发送) │
        └──────────┘ └──────────┘ └──────────┘     └──────────┘
```

### 2.2 架构决策：模块化单体（单服务 + group 分组）

采用**纯模块化单体**，不拆任何微服务——博客低并发、数据强耦合，微服务收益为零却要付出分布式事务与运维成本。（v1.0 曾计划单拆 notify zRPC 服务，v1.1 起放弃：为几封邮件维护一个独立服务不划算。）

- **单服务 + group 分组路由**：go-zero 的 `.api` 文件按领域分组（user / post / comment / site / notify），各组独立中间件与权限声明
- **模块边界**：logic 层按领域划分，模块间只通过接口调用，不共享数据访问层；邮件发送收敛在 notify 模块内，进程内异步执行（goroutine + 重试），调用方无感
- **未来退路**：若流量真实增长，go-zero 的 group 与 logic 边界可低成本平移为独立 zRPC 服务——但那是"真有需要"时的事

### 2.3 技术选型

| 层 | 选型 | 理由 |
|---|---|---|
| 框架 | go-zero | 学习目标框架：goctl 代码生成、限流/熔断/超时等治理能力；group 分组路由天然支持模块化单体 |
| 数据访问 | go-zero 内置 sqlx | goctl 由 DDL 直接生成 model 层代码，类型安全、无反射开销；复杂查询手写 SQL |
| 数据库 | SQLite 3（WAL 模式） | 零依赖内嵌、单文件易备份、内存占用极低，博客读多写少完全够用 |
| 缓存 | Redis | 热点文章缓存、阅读量异步计数、接口限流桶、Token 黑名单 |
| 文件存储 | RustFS（S3 兼容） | Rust 编写的 S3 兼容对象存储，Apache 2.0 无 AGPL 顾虑；空闲内存约 110MB（约为 MinIO 的 1/3），契合 2G 小机；S3 协议是事实标准，未来可切回 MinIO 或上云 OSS/COS/S3。注意：项目尚处 alpha 阶段，仅用于单节点小数据量场景并配合每日备份 |
| 第三方登录 | GitHub OAuth | 读者以开发者为主；零申请门槛；QQ 互联待备案后接入 |
| 邮件服务 | QQ 邮箱 SMTP（免费授权码） | 注册验证邮件；量大了再换 Resend 等专业服务 |
| 认证 | JWT（Access + Refresh Token）+ bcrypt | 无状态，适合前后端分离 |

### 2.4 目录结构

```
server/                         # 单服务：go-zero API（goctl 生成骨架）
├── blog.api                    # API 定义文件（按领域 group 分组：user/post/comment/site/notify）
├── blog.go                     # 入口（goctl 生成）
├── etc/blog.yaml               # 配置文件（go-zero 原生配置）
├── internal/
│   ├── config/                 # 配置结构体
│   ├── handler/                # HTTP 处理器（由 .api 按 group 生成路由）
│   │   ├── user/
│   │   ├── post/
│   │   ├── comment/
│   │   └── site/
│   ├── logic/                  # 业务逻辑层，按领域分模块：
│   │   ├── user/               #   用户/认证模块
│   │   ├── post/               #   文章/标签模块
│   │   ├── comment/            #   评论模块
│   │   ├── site/               #   站点配置模块
│   │   └── notify/             #   通知模块：邮件组装、SMTP 异步发送、失败重试
│   │                           #   （进程内 goroutine，对调用方暴露 SendAsync 接口）
│   ├── svc/                    # 服务上下文（依赖注入：DB、Redis、RustFS、SMTP）
│   ├── model/                  # 数据模型（goctl model sqlite 生成）
│   └── middleware/             # 自定义中间件（JWT、限流等）
├── data/blog.db                # SQLite 数据文件（挂载卷）
├── migrations/                 # 建表与迁移 SQL
└── Dockerfile                  # 注：文件不落本地，无 uploads 目录
```

---

## 3. 数据库设计

### 3.1 ER 关系

```
一期：
admin (1) ───< posts (N) >──< tags (N)      posts (1) ───< comments (N) >── (1) admin
                                                               └── 评论归属登录者（一期仅管理员），支持一级回复（parent_id 自关联）
说明：文章与标签多对多（post_tags 中间表）；一期只有 admin 一张账号表，
登录即管理员，无角色概念。

二期按需演进：admin 表升级/扩展为通用用户表（加 email、role、status 等列），
新建 user_oauth 绑定表。届时通过迁移脚本 ALTER / CREATE 完成，不影响一期数据。
```

### 3.2 表结构

**admin 管理员表**（一期唯一账号表，保持最小结构；用户相关字段二期按需再加）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | INTEGER PRIMARY KEY AUTOINCREMENT | 主键 |
| username | TEXT UNIQUE NOT NULL | 登录名 |
| password_hash | TEXT NOT NULL | bcrypt 哈希 |
| nickname | TEXT | 显示昵称 |
| avatar_url | TEXT | 头像 |
| created_at / updated_at | DATETIME | 时间戳 |

> 二期演进（仅记录方向，届时再定稿）：邮箱/验证状态/角色/封禁等列 ALTER 加入；新建 `user_oauth`（user_id / provider / openid，(provider,openid) 唯一）支撑 GitHub / QQ 登录。

**posts 文章表**

| 字段 | 类型 | 说明 |
|---|---|---|
| id | INTEGER PRIMARY KEY AUTOINCREMENT | 主键 |
| title | TEXT NOT NULL | 标题 |
| slug | TEXT UNIQUE NOT NULL | URL 友好别名（SEO） |
| summary | TEXT | 摘要（列表页/SEO description） |
| content | TEXT NOT NULL | Markdown 原文 |
| cover_url | TEXT | 封面图 |
| status | INTEGER NOT NULL DEFAULT 0 | 0=草稿 1=已发布 2=隐藏 |
| views | INTEGER NOT NULL DEFAULT 0 | 阅读量 |
| published_at | DATETIME | 发布时间 |
| created_at / updated_at | DATETIME | 时间戳 |

> 索引：`status + published_at DESC` 联合索引（列表查询）；`slug` 唯一索引；全文搜索使用 **SQLite FTS5 虚拟表**（`posts_fts`，同步 title/content），中文分词一期用 trigram 或简单 LIKE 兜底，二期可评估结巴分词插件。

**tags 标签表**

| 字段 | 类型 | 说明 |
|---|---|---|
| id | INTEGER PRIMARY KEY AUTOINCREMENT | 主键 |
| name | TEXT UNIQUE NOT NULL | 标签名 |
| created_at | DATETIME | 创建时间 |

**post_tags 文章-标签中间表**

| 字段 | 类型 | 说明 |
|---|---|---|
| post_id | INTEGER FK → posts(id) ON DELETE CASCADE | 复合主键之一 |
| tag_id | INTEGER FK → tags(id) ON DELETE CASCADE | 复合主键之一 |

**comments 评论表**

| 字段 | 类型 | 说明 |
|---|---|---|
| id | INTEGER PRIMARY KEY AUTOINCREMENT | 主键 |
| post_id | INTEGER FK → posts(id) ON DELETE CASCADE | 所属文章 |
| parent_id | INTEGER FK → comments(id) NULL | 回复目标（仅一级） |
| author_id | INTEGER FK → admin(id) NOT NULL | 评论者（一期只有管理员会写入；二期随用户表演进调整外键） |
| content | TEXT NOT NULL | 评论内容（纯文本，防 XSS，长度 ≤1000） |
| status | INTEGER NOT NULL DEFAULT 1 | 0=待审核 1=通过 2=垃圾 |
| ip | TEXT | 记录 IP（限流与风控） |
| created_at | DATETIME | 时间戳 |

**site_configs 站点配置表**（KV 结构，避免硬编码）

| 字段 | 类型 | 说明 |
|---|---|---|
| key | TEXT PRIMARY KEY | 配置键（site_title / icp / github_url …） |
| value | TEXT | 配置值 |

### 3.3 SQLite 配置与迁移策略

- **WAL 模式**：`PRAGMA journal_mode=WAL`，读写不互斥，并发读性能显著提升
- **外键约束**：SQLite 默认关闭，连接串启用 `_foreign_keys=on`
- **busy 处理**：设置 `busy_timeout=5000`，避免写锁冲突直接报错
- **迁移**：版本化 SQL 迁移文件，服务启动时检查 `schema_migrations` 表自动执行；`blog.db` 文件随容器卷持久化

### 3.4 Redis 数据规划

| Key 模式 | 用途 | TTL | 分期 |
|---|---|---|---|
| `token_blacklist:{jti}` | 登出后的 RefreshToken 黑名单 | = Token 剩余有效期 | 一期 |
| `rate:{action}:{id}` | 限流桶（登录/评论） | 按窗口 | 一期 |
| `post:hot` | 热点文章缓存 | 10min | 一期 |
| `views:{post_id}` | 阅读量计数（异步落库） | 持久，定时同步 | 一期 |
| `verify:{token}` | 邮箱验证 token → user_id | 24h | 二期（预留） |
| `oauth_state:{state}` | OAuth state 防 CSRF | 5min | 二期（预留） |
| `reset:{token}` | 找回密码 token → user_id | 30min | 二期（预留） |

---

## 4. API 设计（前后端契约）

### 4.1 设计规范

- 风格：RESTful，资源名复数，统一前缀 `/api/v1`
- 认证：`Authorization: Bearer <access_token>`
- 统一响应结构：

```json
{ "code": 0, "message": "ok", "data": { } }
```

- 错误码：`0` 成功；`400xx` 参数错误；`401xx` 未认证/Token 失效；`403xx` 无权限；`404xx` 资源不存在；`429xx` 触发限流；`500xx` 服务端错误
- 分页参数：`?page=1&size=10`，响应含 `total`

### 4.2 接口清单

**前台公开接口**

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/v1/posts | 文章分页列表，支持 `?tag=&keyword=` 过滤 |
| GET | /api/v1/posts/:slug | 文章详情（按 slug，SEO 友好） |
| POST | /api/v1/posts/:slug/view | 阅读量 +1（前端进入详情页时调用，IP 去重） |
| GET | /api/v1/tags | 标签列表（含文章数） |
| GET | /api/v1/posts/:slug/comments | 评论列表 |
| GET | /api/v1/site/config | 站点公开配置 |
| GET | /rss.xml | RSS 订阅输出 |

**认证接口**

| 方法 | 路径 | 分期 | 说明 |
|---|---|---|---|
| POST | /api/v1/auth/login | **一期** | 管理员用户名 + 密码登录 |
| POST | /api/v1/auth/refresh | **一期** | 刷新 access token |
| POST | /api/v1/auth/logout | **一期** | 登出（RefreshToken 加入 Redis 黑名单） |
| POST | /api/v1/auth/register | 二期 | 邮箱注册（限流：同 IP 每小时 5 次） |
| GET | /api/v1/auth/verify-email?token= | 二期 | 邮箱验证（注册后激活） |
| POST | /api/v1/auth/resend-verify | 二期 | 重发验证邮件（限流） |
| GET | /api/v1/auth/oauth/github | 二期 | 跳转 GitHub 授权（带 state 防 CSRF） |
| GET | /api/v1/auth/oauth/github/callback | 二期 | GitHub 回调，换取 token 完成登录/注册 |
| POST | /api/v1/auth/forgot-password | 二期 | 发送重置密码邮件 |
| POST | /api/v1/auth/reset-password | 二期 | 凭邮件 token 重置密码 |

**登录用户接口（需 JWT）**

| 方法 | 路径 | 分期 | 说明 |
|---|---|---|---|
| GET | /api/v1/user/profile | **一期** | 当前用户信息（一期即 admin 自己） |
| PUT | /api/v1/user/profile | **一期** | 修改昵称 / 头像 |
| PUT | /api/v1/user/password | **一期** | 修改密码 |
| POST | /api/v1/posts/:slug/comments | **一期**（仅 admin） | 一期仅管理员可发表（博主回复）；二期开放注册用户（限流：同用户 1 分钟 1 条） |
| DELETE | /api/v1/comments/:id | 二期 | 删除自己的评论 |

**后台管理接口（需 JWT 登录，一期只有管理员，登录即授权）**

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/v1/admin/posts | 管理列表（含草稿，支持状态过滤） |
| POST | /api/v1/admin/posts | 新建文章 |
| PUT | /api/v1/admin/posts/:id | 更新文章（含发布/草稿切换） |
| DELETE | /api/v1/admin/posts/:id | 删除文章 |
| POST | /api/v1/admin/upload | 图片上传（multipart，限制 5MB，jpg/png/webp），后端转存 RustFS 并返回访问 URL |
| GET / PUT | /api/v1/admin/site/config | 读取 / 修改站点配置 |
| PUT | /api/v1/admin/comments/:id | 评论审核 / 标记垃圾 |
| DELETE | /api/v1/admin/comments/:id | 删除评论 |

> 上传流程：前端 → 后端（校验类型/大小、生成缩略图）→ RustFS `blog-images` bucket → 返回 `https://<域名>/files/<bucket>/<key>`。文件访问由 Nginx 反代 RustFS，不经过 Go 服务。
>
> 二期预留接口（届时再实现）：`GET /api/v1/admin/users` 用户列表、`PUT /api/v1/admin/users/:id/status` 封禁/解封。

### 4.3 认证流程

**一期：管理员登录**

```
登录:  username+password ──► 验证 bcrypt ──► 签发 AccessToken(2h) + RefreshToken(14d)
请求:  前端携带 AccessToken ──► JWT 中间件校验（一期登录即管理员）──► 过期则前端静默调 /auth/refresh 重试
登出:  RefreshToken 加入 Redis 黑名单（TTL=剩余有效期）
存储:  RefreshToken 存 httpOnly Cookie；AccessToken 由前端存内存，不落 localStorage
```

**二期预留：邮箱注册**

```
注册 ──► 写用户表(email_verified=0) ──► 调 notify 模块 SendAsync 异步发验证邮件
   （token 存 Redis，24h 有效）──► 用户点链接验证 ──► email_verified=1，正式激活
未验证账号：允许登录但不可评论，7 天未验证自动清理
```

**二期预留：GitHub OAuth**

```
前端点"GitHub 登录" ──► 302 到 GitHub（带 state 防 CSRF，state 存 Redis 5 分钟）
──► 回调拿 code 换 access_token ──► 拉取 GitHub 用户信息
──► 查 user_oauth：已绑定则直接登录；未绑定则自动创建用户（头像/昵称同步）
```

---

## 5. 非功能性设计

### 5.1 安全

| 风险 | 对策 |
|---|---|
| 密码安全 | bcrypt(cost=12) 哈希存储，登录限流（同 IP 每分钟 5 次） |
| SQL 注入 | 全程参数化查询（goctl 生成的 model 与手写 SQL 均禁止字符串拼接） |
| CSRF | API 使用 Bearer Token（非 Cookie 认证）天然免疫；RefreshToken Cookie 加 SameSite=Strict |
| Token 窃取 | AccessToken 短时效 2h；RefreshToken httpOnly + Secure |
| 接口滥用 | go-zero 内置限流中间件（评论、登录、view 接口）；上传限制类型与大小 |
| XSS（服务端侧） | 评论纯文本存储，输出不拼接 HTML；响应设置安全头 |
| 二期新增关注 | 批量注册（限流 + 邮箱验证 + 未验证 7 天清理）、OAuth state 防 CSRF——随注册功能落地时启用 |

### 5.2 性能

- 列表接口必分页（默认 10 条）；列表不返回 content 字段
- 数据库索引见 3.2；全文搜索用 SQLite FTS5 虚拟表，万级文章内性能绰绰有余
- go-zero 内置超时、熔断、自适应限流，直接获得生产级韧性
- 图片上传时生成缩略图（列表页用），原图仅详情页加载
- Redis 缓存热点文章；阅读量 Redis 计数、定时批量落库，避免高频写 SQLite

### 5.3 日志

- 结构化日志（go-zero 内置 logx / zap），按天切割
- notify 模块记录邮件发送结果，失败重试 3 次后写错误日志留档

### 5.4 备份

- 每日 cron：SQLite 在线备份（`sqlite3 blog.db ".backup ..."` 或 `VACUUM INTO`，避免直接 cp 正在写入的库文件）+ RustFS 数据目录打包（或用 `mc` / `rclone` 同步，S3 兼容工具均可），保留最近 30 天
- 备份文件异地存储——SQLite 单文件 + RustFS 目录结构让备份与迁移都极其简单

---

## 6. 部署与配置

### 6.1 全 Docker 化部署（docker-compose 一套编排）

**所有组件均以容器运行，宿主机只安装 Docker 本身**，不在宿主机上直接装 Nginx / Go / Node 等任何环境。前端构建产物也以镜像或挂载卷形式交给 Nginx 容器托管。

```yaml
services:
  nginx:      # 入口：80/443；/api → server，/files → rustfs，其余 → 前端静态产物
  web:        # 前端构建产物（镜像内打包 dist，或构建卷挂载给 nginx）
  server:     # go-zero 单服务，内嵌 SQLite，挂载数据卷（blog.db）
  rustfs:     # 对象存储（S3 兼容），挂载数据卷；控制台端口仅内网开放
  redis:      # 缓存/限流/Token 黑名单，仅 Docker 内网，不暴露端口
```

要点：
- **对外只暴露 Nginx 的 80/443**，server / rustfs / redis 全部只在 Docker 内网可达
- 数据全部走挂载卷：`blog.db`（SQLite）、RustFS 数据目录、Redis（可选 AOF）
- 每个服务设 `mem_limit`；`restart: unless-stopped` 保证重启自愈
- 镜像全部由 GitHub Actions 构建推送（含前端），服务器只 `pull && up -d`——2G 内存不在本机构建

### 6.2 服务器资源预算（2 核 2G，全栈）

| 服务 | 常驻内存（估算） | 说明 |
|---|---|---|
| Nginx | ~20 MB | 静态资源 + 反代 |
| web 前端产物 | ~0 | 静态文件由 Nginx 直接托管，无独立运行时 |
| go-zero 后端（含 SQLite + 邮件异步任务） | 50~150 MB | 单服务进程 |
| RustFS | ~110 MB（空载） | compose `mem_limit` 约束；轻量模式下单机占用远低于 MinIO |
| Redis | ~30 MB | `maxmemory 128mb` + `allkeys-lru`；持久化只开 `appendonly everysec` |
| 系统 + Docker 守护进程 | ~300 MB | — |
| **合计** | **约 0.5~0.8 GB** | **余量 1.2GB+，2C2G 充裕（RustFS 比 MinIO 再省约 150MB）** |

配套措施：配置 1~2GB swap（或 zram）应对备份等瞬时高峰；备份脚本低峰串行执行。

### 6.3 发布流程（CI/CD）

```
git push main
   └── GitHub Actions: 后端 go test → 构建 server 镜像 → SSH 到 VPS
       └── docker compose pull && up -d（零停机可后续优化为蓝绿）
```

### 6.4 配置项

走 go-zero 的 `etc/blog.yaml` + 环境变量覆盖（敏感项不入库）：

| 配置 | 说明 |
|---|---|
| SQLite 文件路径 | data/blog.db |
| JWT 密钥 | Access / Refresh 签名 |
| 管理员初始密码 | 首次启动初始化 admin 账号 |
| RustFS endpoint / access key / bucket | 文件存储（S3 兼容，SDK 无需更换） |
| Redis 地址 | 缓存与限流 |
| SMTP 授权码 | QQ 邮箱发信（二期 notify 模块启用，一期可先不配置） |
| GitHub OAuth client id/secret | 第三方登录（二期再配置） |

本地开发零依赖：SQLite 内嵌，启动后端即可联调。

---

## 7. 服务端里程碑

| 阶段 | 交付物 |
|---|---|
| M2 | goctl 生成骨架，文章/标签/配置 API + 管理员 JWT 登录 |
| M3 | 上传与 RustFS、评论区（管理员可发表/回复） |
| M4 | Docker 部署至 VPS，HTTPS 可访问 |
| M5 | CI/CD 流水线、RSS/sitemap 输出 |

### 二期 backlog

- **开放注册与用户体系**（优先级最高）：用户表演进 + user_oauth、邮箱注册/验证、GitHub OAuth、评论开放给注册用户、管理员用户管理接口
- 评论邮件通知（notify 模块扩展）、邮件订阅
- QQ 互联登录（待域名备案完成后申请）
- 全文搜索增强（FTS5 + 中文分词扩展，或迁独立搜索引擎）
- 对象存储平滑迁移至云 OSS/COS/S3 或成熟版 MinIO（仅改配置）+ CDN 加速
- 仅当流量真实增长且出现明确瓶颈时，才评估将模块平移为 zRPC 服务（group/logic 边界已预留）

---

## 8. 服务端风险

| 风险 | 说明 | 对策 |
|---|---|---|
| 评论垃圾信息 | 一期评论仅管理员可写，无风险；二期开放注册后上升 | 二期：评论强制登录 + 邮箱验证 + 限流 + 管理员封禁，必要时加 Turnstile |
| SQLite 并发写入上限 | 单写者模型，高并发写入排队 | 写操作极少；WAL + busy_timeout 已覆盖；流量暴涨再迁 PostgreSQL |
| RustFS 单节点无冗余 | 挂盘即丢数据 | 每日备份 + `mc`/`rclone` 异地同步 |
| RustFS 处于 alpha 阶段 | 项目较新，生产验证有限，大文件读性能暂弱 | 仅限单节点小数据量场景；每日备份兜底；出严重问题可随时切回 MinIO 或直上云 OSS（S3 兼容，业务代码零改动） |
| SMTP 发信失败 | QQ 邮箱 SMTP 偶发限制 | notify 模块异步重试；监控日志；量大后换 Resend |
| 单点 VPS 故障 | 单机部署无高可用 | 每日备份 + 可复现的 compose 环境，重装成本 < 30 分钟 |

---

*本文档与《博客系统设计文档-前端》共同构成完整设计；API 契约（第 4 章）变更需同步前端文档。*
