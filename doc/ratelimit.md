# 接口限流方案（登录 / 评论）

| 项目 | 内容 |
|---|---|
| 针对问题 | review 高危 #3（登录可无限爆破）、#4（游客评论无频率限制） |
| 现状 | 项目无 Redis 依赖，单进程部署；`go.mod` 无限流相关库 |
| 编写日期 | 2026-09-10 |

---

## 1. 目标与限额

| 接口 | 限额 | 维度 | 理由 |
|---|---|---|---|
| POST /v1/auth/login | 5 次/分钟 | 客户端 IP | 设计文档 5.1 既定值，挡在线爆破 |
| POST /v1/articles/:slug/comments | 1 条/分钟 | 客户端 IP | 游客场景无 user_id，按 IP；对应设计"同用户 1 分钟 1 条" |
| 登录失败附加策略 | 同 IP+用户名 连续失败 5 次锁定 10 分钟 | IP+用户名 | 防止换着 IP 慢速爆破固定账号（单机内存计数即可） |

不做全局 QPS 兜底：go-zero 自适应限流（shedding）已默认生效，足够。

## 2. 技术选型：进程内令牌桶（一期）

用 `golang.org/x/time/rate`（新增唯一依赖），不用 go-zero 内置 `core/limit.TokenLimiter`——那个依赖 Redis，一期没有 Redis。

- 单进程部署，内存限流无一致性问题
- 重启后计数清零，对限流场景可接受
- 二期引入 Redis 后平移：`Allow(key)` 接口不变，实现换成 Redis + Lua（见第 6 节）

## 3. 实现结构

```
blog/internal/middleware/
└── ratelimitmiddleware.go    # 中间件
pkg/ratelimit/
└── ratelimit.go              # 与 go-zero 无关的限流器本体（易测试）
```

**`pkg/ratelimit/ratelimit.go`**——带清理的 per-key 令牌桶：

```go
type Limiter struct {
    mu     sync.Mutex
    buckets map[string]*entry  // key -> {*rate.Limiter, lastSeen}
    rate   rate.Limit
    burst  int
}

func New(r rate.Limit, burst int) *Limiter  // 启动时建，后台 goroutine 每 5min 清理 10min 未活跃的 key（防内存膨胀）
func (l *Limiter) Allow(key string) bool
```

**中间件**：取客户端 IP 作 key，超限直接写统一错误体并中断：

- IP 提取：请求来自 Nginx 反代，优先 `X-Real-IP`（仅当 `RemoteAddr` 是回环/内网地址时信任，防伪造），否则用 `RemoteAddr` 去端口
- 响应复用统一格式：`{"code":42901,"message":"操作太频繁，请稍后再试"}`（HTTP 仍 200，与项目约定一致）
- `pkg/errx/code.go` 新增 `RateLimited = 42901`
- 注意：统一响应体结构目前定义在 `blog.go` 的 main 包里，中间件要用得先把它下沉到 `pkg/errx`（加一个 `WriteError(w, err)` 帮助函数），`blog.go` 的 handler 一并改用它

**登录失败锁定**放在 `loginlogic.go` 内（不属于中间件）：

```go
// svc 或 loginlogic 持有一个 ratelimit.Limiter 风格的失败计数器
key := ip + "|" + username
if 已锁定 { return errx.New(42901, "尝试次数过多，请 10 分钟后再试") }
if 密码错误 { 计数+1；满 5 次设置 10 分钟锁定；return ErrLoginFailed }
登录成功 { 清空该 key 计数 }
```

## 4. 挂载方式

`.api` 文件里给对应 group 声明中间件，goctl 重新生成路由：

```
@server (
    group:      auth
    middleware: RateLimit   // 登录/刷新都在这组，按需可再细分
)
```

`interact` 组同理。生成后在 `svc/servicecontext.go` 或 handler 注册处把 `*ratelimit.Limiter` 实例注进去（不同路由不同限额 → 中间件构造函数带参数，或拆 `LoginRateLimit` / `CommentRateLimit` 两个实例）。

> 若不想重新生成路由，也可在 `blog.go` 用 `server.Use()` 全局挂载 + 中间件内按 `r.URL.Path` 匹配，但不推荐：路径字符串散落两处，易腐化。

## 5. 验证清单

- 单测：`pkg/ratelimit` 的 Allow 边界（burst 内通过、耗尽拒绝、清理后重建）
- 手测：`for i in {1..6}; do curl -X POST .../auth/login ...; done` 第 6 次返回 42901
- 评论接口同理，1 分钟内第 2 条被拒
- 登录锁定：连续 5 次错误密码后，即使密码正确也提示锁定；10 分钟后（或重启）恢复

## 6. 二期平移 Redis

`pkg/ratelimit.Limiter` 抽成接口：

```go
type Limiter interface { Allow(ctx context.Context, key string) bool }
```

新增 Redis 实现（`core/limit.TokenLimiter` 或自写 Lua 固定窗口），`svc` 按配置开关选实现，logic/中间件零改动。届时 key 维度可从 IP 升级为"登录用户按 user_id、游客按 IP"。
